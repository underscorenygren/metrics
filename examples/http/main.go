package main

import (
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/http"
	"github.com/underscorenygren/metrics/pkg/optional"
	"github.com/underscorenygren/metrics/pkg/sink/blackhole"
	"os"
)

func main() {
	logger := logging.ConfigureDevelopment(os.Stderr)
	sink := blackhole.Sink()
	s, err := http.NewServer(http.Config{
		Port: optional.Int(3033),
		Host: optional.String("127.0.0.1"),
		Sink: sink,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Fatal(s.ListenAndServe().Error())
}
