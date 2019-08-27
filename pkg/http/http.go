package http

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/pipe"
	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/programmatic"
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
	stages        []types.Stage
	sources       []types.Source
}

//Server accepts events over HTTP
//Can be configured to route requests to different sinks, will
//ship all to default sink unless configured otherwise
type Server struct {
	eventServer *eventServer
	httpServer  *http.Server
	Router      *mux.Router
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
	Router            *mux.Router
}

//NewServer makes a new server instance
func NewServer(cfg Config) (*Server, error) {
	host := DefaultHost
	port := DefaultPort
	readHeaderTimeout := DefaultReadHeaderTimeout
	readTimeout := DefaultReadTimeout
	writeTimeout := DefaultWriteTimeout
	eventMaker := defaultEventMaker
	successWriter := defaultSuccessFn
	router := mux.NewRouter()

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
		stages:        []types.Stage{},
	}

	//all servers are created with routing enabled, but as
	//a shortcut we allow providing just a sink to route
	//there by default without using handlefunc
	if sink := cfg.Sink; sink != nil {
		router.NotFoundHandler = http.HandlerFunc(eventServer.makeHandleFunc(sink))
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
	}

	return &Server{
		eventServer: eventServer,
		httpServer:  httpServer,
		Router:      router,
	}, nil
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
		channel := pipeline.ParalellFailFirst(srv.eventServer.stages, logger)
		for err := range channel {
			logger.Error("event pipeline error", zap.Error(err))
			errChan <- err
			break
		}
		logger.Debug("http.ListenAndServe: pipeline ended")
	}()

	return <-errChan
}

//MakeHandleFunc Used to configure routing. Provide as argument to
//mux.HandleFunc by accessing the underlying router
func (srv *Server) MakeHandleFunc(sink types.Sink) func(w http.ResponseWriter, req *http.Request) {
	return srv.eventServer.makeHandleFunc(sink)
}

//Shutdown stops the server
func (srv *Server) Shutdown(ctx context.Context) {
	logger := logging.Logger()
	logger.Debug("http.Shutdown starting")
	for _, src := range srv.eventServer.sources {
		if err := src.Close(); err != nil {
			logger.Debug("http.Shutdown src close err", zap.Error(err))
		}
	}
	srv.httpServer.Shutdown(ctx)
	logger.Debug("http.Shutdown finished")
}

//internally, all handlefunc logic is done on eventServer
//Adds a sink and routes request to it using programmatic sink
func (s *eventServer) makeHandleFunc(sink types.Sink) func(w http.ResponseWriter, req *http.Request) {

	src := s.addSink(sink)
	s.sources = append(s.sources, src)
	//This is the ServeHTTP request
	return func(w http.ResponseWriter, req *http.Request) {

		logger := logging.Logger()

		body, err := readBody(req)
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
			err = src.Put(*evt)
			if err != nil {
				s.handleError(err, w)
				return
			}
		} else {
			logger.Debug("event pruned")
		}

		s.SuccessWriter(w)
	}
}

//writes en arror when request is malformatted
func (s *eventServer) handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	logging.Logger().Error("error on request", zap.Error(err))
}

//adds sink into internal structure and creates corresponding source
func (s *eventServer) addSink(sink types.Sink) *programmatic.Source {
	src := programmatic.NewSource()
	stage, err := pipe.Stage(src, sink)
	if err != nil {
		logging.Logger().Fatal("couldn't create stage", zap.Error(err))
	}
	s.stages = append(s.stages, stage)
	return src
}

//readBody reads body from a request
func readBody(req *http.Request) ([]byte, error) {
	return ioutil.ReadAll(req.Body)
}
