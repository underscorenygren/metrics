package server

import (
	"fmt"
	"github.com/underscorenygren/metrics/middleware"
	"github.com/underscorenygren/metrics/producer"
	"go.uber.org/zap"
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
)

//SetPort set port
func (cfg *Config) SetPort(port int) *Config {
	copy := Config(*cfg)
	copy.Port = &port
	return &copy
}

//SetHost set host
func (cfg *Config) SetHost(host string) *Config {
	copy := Config(*cfg)
	copy.Host = &host
	return &copy
}

//SetReadHeaderTimeout set read header timeout
func (cfg *Config) SetReadHeaderTimeout(dur time.Duration) *Config {
	copy := Config(*cfg)
	copy.ReadHeaderTimeout = &dur
	return &copy
}

//SetReadTimeout set read timeout
func (cfg *Config) SetReadTimeout(dur time.Duration) *Config {
	copy := Config(*cfg)
	copy.ReadTimeout = &dur
	return &copy
}

//SetWriteTimeout set write timeout
func (cfg *Config) SetWriteTimeout(dur time.Duration) *Config {
	copy := Config(*cfg)
	copy.WriteTimeout = &dur
	return &copy
}

//SetProducer set the producer
func (cfg *Config) SetProducer(p producer.Producer) *Config {
	copy := Config(*cfg)
	copy.Producer = p
	return &copy
}

//SetLogger set logger
func (cfg *Config) SetLogger(logger *zap.Logger) *Config {
	copy := Config(*cfg)
	copy.Logger = logger
	return &copy
}

//SetContextMaker set logger
func (cfg *Config) SetContextMaker(contextMaker ContextMaker) *Config {
	copy := Config(*cfg)
	copy.ContextMaker = contextMaker
	return &copy
}

//SetMiddleware sets middleware
func (cfg *Config) SetMiddleware(transformer middleware.Transformer) *Config {
	copy := Config(*cfg)
	copy.Middleware = transformer
	return &copy
}

//RunForever Build server from config and start serving requests
func (cfg *Config) RunForever() error {

	host := DefaultHost
	port := DefaultPort
	readHeaderTimeout := DefaultReadHeaderTimeout
	readTimeout := DefaultReadTimeout
	writeTimeout := DefaultWriteTimeout
	var logger *zap.Logger
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
	if cfg.Logger != nil {
		logger = cfg.Logger
	} else {
		zapCfg := zap.NewProductionConfig()
		zapCfg.Level.SetLevel(zap.DebugLevel)
		var err error
		logger, err = zapCfg.Build()
		if err != nil {
			return err
		}
	}

	s := server{
		logger:       logger,
		p:            cfg.Producer,
		middleware:   cfg.Middleware,
		contextMaker: cfg.ContextMaker,
	}
	addr := fmt.Sprintf("%s:%d", host, port)

	s.logger.Info("starting server", zap.String("addr", addr))

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           &s,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
	}

	return httpServer.ListenAndServe()
}
