package main

import (
	"github.com/underscorenygren/metrics/ingest/server"
	"github.com/underscorenygren/metrics/producer/kinesis"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level.SetLevel(zap.DebugLevel)
	zapper, err := zapCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	name := os.Getenv("NAME")
	if name == "" {
		log.Fatal("Must set NAME env var")
	}
	zapper.Debug("starting kinesis", zap.String("name", name))

	(&server.Config{}).
		SetProducer(kinesis.New(name, zapper)).
		SetLogger(zapper).
		RunForever()
}
