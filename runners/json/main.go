package main

import (
	"context"
	"fmt"
	"github.com/underscorenygren/metrics/ingest/server"
	"github.com/underscorenygren/metrics/middleware/json"
	"github.com/underscorenygren/metrics/producer/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type contextKey int

const (
	headerKey contextKey = iota
	pathKey
	headerValue = "header"
)

func contextParser(req *http.Request) (context.Context, error) {

	ctx := req.Context()
	if req.Header != nil {
		ctx = context.WithValue(ctx, headerKey, req.Header.Get(headerValue))
	}

	if req.URL == nil {
		return nil, fmt.Errorf("req url is nil")
	}
	fmt.Println(req.URL.Path)
	ctx = context.WithValue(ctx, pathKey, req.URL.Path[1:])

	return ctx, nil
}

func main() {
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level.SetLevel(zap.DebugLevel)
	zapper, err := zapCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	(&server.Config{}).
		SetProducer(logger.New(zapper)).
		SetContextMaker(contextParser).
		SetMiddleware(json.New([]json.Processor{
			json.AddFromContext(headerKey, headerValue),
			json.AddFromContext(pathKey, "path"),
			json.UTCTimestamp("@timestamp"),
		})).
		SetLogger(zapper).
		RunForever()
}
