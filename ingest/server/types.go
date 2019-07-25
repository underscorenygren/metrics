package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/underscorenygren/metrics/middleware"
	"github.com/underscorenygren/metrics/producer"
	"go.uber.org/zap"
	"net/http"
	"time"
)

//ContextMaker parses and generates context per-request
type ContextMaker func(*http.Request) (context.Context, error)

type server struct {
	logger       *zap.Logger
	p            producer.Producer
	middleware   middleware.Transformer
	contextMaker ContextMaker
	healthcheck  *Healthcheck
	r            *mux.Router
}

//Healthcheck configures healthcheck handling on path/method
type Healthcheck struct {
	Path   string
	Method string
}

//Config server configuration
type Config struct {
	Port              *int
	Host              *string
	ReadHeaderTimeout *time.Duration
	ReadTimeout       *time.Duration
	WriteTimeout      *time.Duration
	Producer          producer.Producer
	Logger            *zap.Logger
	Middleware        middleware.Transformer
	ContextMaker      ContextMaker
	Healthcheck       *Healthcheck
	Router            *mux.Router
}

//Server simpler interface for server
type Server interface {
	ListenAndServe() error
}
