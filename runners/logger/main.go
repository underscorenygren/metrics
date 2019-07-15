package main

import (
	"github.com/underscorenygren/metrics/ingest/server"
	"github.com/underscorenygren/metrics/producer/logger"
	"go.uber.org/zap"
	"log"
)

func main() {

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level.SetLevel(zap.DebugLevel)
	zapper, err := zapCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	(&server.Config{}).
		SetProducer(logger.New(zapper)).
		SetLogger(zapper).
		RunForever()
}
