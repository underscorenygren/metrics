package main

import (
	"fmt"
	"github/underscorenygren/producer/logger"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type config struct {
	logger *zap.Logger
	p      producer.Producer
}

func (cfg *config) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		cfg.logger.Error("error reading body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
	} else {
		cfg.p.PutRecord(body)
		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level.SetLevel(zap.DebugLevel)
	logger, err := zapCfg.Build()

	if err != nil {
		log.Fatal(fmt.Sprintf("couldn't init logger: %s", err))
	}

	p := producer.Log(logger)
	cfg := &config{logger: logger}

	logger.Info("starting logger")

	s := &http.Server{
		Addr:              "0.0.0.0:80",
		Handler:           cfg,
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      4 * time.Second,
	}

	logger.Fatal(s.ListenAndServe().Error())
}
