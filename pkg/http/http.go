/*
Package http provides a webserver that converts requests into events,
as well as some minimal routing of events to different event processing pipelines.
*/
package http

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/pipeline"
	"github.com/underscorenygren/partaj/pkg/programmatic"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	//DefaultHost listens on all interfaces
	DefaultHost = "0.0.0.0"
	//DefaultPort listens on web port
	DefaultPort = 80
	//DefaultReadHeaderTimeout is 1 second
	DefaultReadHeaderTimeout = 1 * time.Second
	//DefaultReadTimeout is 2 seconds
	DefaultReadTimeout = 2 * time.Second
	//DefaultWriteTimeout is 4 seconds
	DefaultWriteTimeout = 4 * time.Second
	//DefaultSuccessCode writes 201 on success
	DefaultSuccessCode = http.StatusNoContent
)

/*
EventMakerFn is the function signature for making an event from a request.

body is provided as an argument, it's read in-advance from the http.Request.Body.
As such, attempts to read req.Body will always return "" and should not be used.
*/
type EventMakerFn func(body []byte, req *http.Request) (*types.Event, error)

//SuccessWriterFn is the function signature for writing a successful response.
type SuccessWriterFn func(w http.ResponseWriter)

//DefaultEventMaker implements EventMakerFn type,  writes request body bytes as the event bytes.
func DefaultEventMaker(body []byte, req *http.Request) (*types.Event, error) {
	evt := types.NewEventFromBytes(body)
	return &evt, nil
}

//DefaultSuccessFn implements SuccessWriterFn type, writes the default response code and no content.
func DefaultSuccessFn(w http.ResponseWriter) {
	w.WriteHeader(DefaultSuccessCode)
}

/*
Server accepts web request and turns them into events.

Call ListenAndServe like a regular net/http server to start it.

	server, _ := NewServer(Config{})
	log.Fatal(server.ListenAndServe())

*/
type Server struct {
	eventServer *eventServer
	httpServer  *http.Server
	Router      *mux.Router //allows access to the gorilla/mux Router
}

/*
Config is the input arguments to NewServer.

All fields are optional, and will be filled in with their
corresponding default.

If no Sink is provided, at least one Sink must be added
using MakeHandleFunc for the server to run correctly (see below for example).

If a Sink is provided, it will be registered as a catch-all sink, that receives
all events not covered by other routes registered by MakeHandleFunc.
*/
type Config struct {
	Port              *int            //listen on port
	Host              *string         //listen on host interface
	ReadHeaderTimeout *time.Duration  //passed to net/http
	ReadTimeout       *time.Duration  //passed to net/http
	WriteTimeout      *time.Duration  //passed to net/http
	EventMaker        EventMakerFn    //How to make events from requests
	SuccessWriter     SuccessWriterFn //what to write on event success
	Sink              types.Sink      //sink to handle received events
}

/*
eventServer internal class for handling the events internally.

Extrapolated away from actual http handling, which is covered
by the http module
*/
type eventServer struct {
	EventMaker    EventMakerFn
	SuccessWriter SuccessWriterFn
	stages        []types.Stage
	sources       []types.Source
}

//NewServer makes a new server from the config.
func NewServer(cfg Config) (*Server, error) {
	host := DefaultHost
	port := DefaultPort
	readHeaderTimeout := DefaultReadHeaderTimeout
	readTimeout := DefaultReadTimeout
	writeTimeout := DefaultWriteTimeout
	eventMaker := DefaultEventMaker
	successWriter := DefaultSuccessFn
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

/*
ListenAndServe starts the Server and starts accepting incoming requests
accordinging to the supplied configuration.

Will run indefintely, and returns an error if underlying server fails or event processing
pipeline processing fails.
*/
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

/*
MakeHandleFunc is used to create a function usable with the http.Handler interface, that
sends events to the specified sink.

Used with the server Router to route events to different sink:
	server.Router.HandleFunc("/some-path", server.MakeHandleFunc(someSink))
*/
func (srv *Server) MakeHandleFunc(sink types.Sink) func(w http.ResponseWriter, req *http.Request) {
	return srv.eventServer.makeHandleFunc(sink)
}

//Shutdown stops the server gracefully.
//See http.Shutdown for context usage.
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

/*
makeHandleFunc is the internal handling for adding a sink to the
eventServer. Uses a programmatic source to put events to the sink.
*/
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
