package http

import (
	"context"
	"fmt"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/source"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	//DefaultHost 0.0.0.0
	DefaultHost = "0.0.0.0"
	//DefaultPort port 80
	DefaultPort = 80
	//DefaultReadHeaderTimeout 1 second
	DefaultReadHeaderTimeout = 1 * time.Second
	//DefaultReadTimeout 2 seconds
	DefaultReadTimeout = 2 * time.Second
	//DefaultWriteTimeout 4 seconds
	DefaultWriteTimeout = 4 * time.Second
	//DefaultSuccessCode code to write on success
	DefaultSuccessCode = http.StatusNoContent
)

//EventMakerFn function signature for making an event from a request
//body is read ahead of time, so provided as an arg and not available as req.Body
type EventMakerFn func(body []byte, req *http.Request) (*types.Event, error)

//SuccessWriterFn function signature for wirting success responses
type SuccessWriterFn func(w http.ResponseWriter)

//defaultEventMaker writes body as event
func defaultEventMaker(body []byte, req *http.Request) (*types.Event, error) {
	evt := types.NewEventFromBytes(body)
	return &evt, nil
}

//default success function writes no content response code
func defaultSuccessFn(w http.ResponseWriter) {
	w.WriteHeader(DefaultSuccessCode)
}

//eventServer inner class used to register ServeHTTP
type eventServer struct {
	EventMaker    EventMakerFn
	SuccessWriter SuccessWriterFn
	p             *pipeline.Pipeline
	src           *source.ProgrammaticSource
}

//Server accepts events over HTTP and drains to configured sink
type Server struct {
	eventServer *eventServer
	httpServer  *http.Server
}

//Config server configuration
type Config struct {
	Port              *int
	Host              *string
	ReadHeaderTimeout *time.Duration
	ReadTimeout       *time.Duration
	WriteTimeout      *time.Duration
	EventMaker        EventMakerFn
	SuccessWriter     SuccessWriterFn
	Sink              types.Sink
}

//NewServer makes a new server instance
func NewServer(cfg Config) (*Server, error) {
	src := source.NewProgrammaticSource()
	host := DefaultHost
	port := DefaultPort
	readHeaderTimeout := DefaultReadHeaderTimeout
	readTimeout := DefaultReadTimeout
	writeTimeout := DefaultWriteTimeout
	eventMaker := defaultEventMaker
	successWriter := defaultSuccessFn
	sink := cfg.Sink
	if sink == nil {
		return nil, fmt.Errorf("cannot have nil sink")
	}

	if cfg.Host != nil {
		host = *cfg.Host
	}
	if cfg.Port != nil {
		port = *cfg.Port
	}
	if cfg.ReadHeaderTimeout != nil {
		readHeaderTimeout = *cfg.ReadHeaderTimeout
	}
	if cfg.ReadTimeout != nil {
		readTimeout = *cfg.ReadTimeout
	}
	if cfg.WriteTimeout != nil {
		writeTimeout = *cfg.WriteTimeout
	}
	if cfg.EventMaker != nil {
		eventMaker = cfg.EventMaker
	}
	if cfg.SuccessWriter != nil {
		successWriter = cfg.SuccessWriter
	}

	eventServer := &eventServer{
		EventMaker:    eventMaker,
		SuccessWriter: successWriter,
		p:             pipeline.NewPipeline(src, sink),
		src:           src,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           eventServer,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
	}

	return &Server{
		eventServer: eventServer,
		httpServer:  httpServer,
	}, nil
}

//ReadBody reads body from a request
func ReadBody(req *http.Request) ([]byte, error) {
	return ioutil.ReadAll(req.Body)
}

//ListenAndServe starts server listening on vents
//returns error if http server or event processor fails
func (srv *Server) ListenAndServe() error {
	errChan := make(chan error)
	logger := logging.Logger()

	go func() {
		logger.Debug("http.ListenAndServe: starting http")
		err := srv.httpServer.ListenAndServe()
		logger.Error("http server error", zap.Error(err))
		errChan <- err
	}()

	go func() {
		logger.Debug("http.ListenAndServe: starting pipeline")
		err := srv.eventServer.p.Flow()
		logger.Error("event pipeline error", zap.Error(err))
		errChan <- err
	}()

	return <-errChan
}

//ServeHTTP Fulfills http interface for webserver
//reads request and submits it as an event to
//configured sink
func (s *eventServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := logging.Logger()

	body, err := ReadBody(req)
	logger.Debug("http.ServeHTTP: request received", zap.ByteString("body", body))
	if err != nil {
		logger.Debug("http.ServeHTTP: ReadBody error")
		s.handleError(err, w)
		return
	}

	evt, err := s.EventMaker(body, req)
	if err != nil {
		logger.Debug("http.ServeHTTP: EventMaker error")
		s.handleError(err, w)
		return
	}

	if evt != nil {
		logger.Debug("made event", zap.ByteString("event", evt.Bytes()))
		err = s.src.Put(*evt)
		if err != nil {
			s.handleError(err, w)
			return
		}
	} else {
		logger.Debug("event pruned")
	}
	s.SuccessWriter(w)
}

//error handling function
func (s *eventServer) handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	logging.Logger().Error("error on request", zap.Error(err))
}

//Shutdown stops the server
func (srv *Server) Shutdown(ctx context.Context) {
	logger := logging.Logger()
	logger.Debug("http.Shutdown starting")
	srv.eventServer.src.Close()
	srv.httpServer.Shutdown(ctx)
	logger.Debug("http.Shutdown finished")
}
