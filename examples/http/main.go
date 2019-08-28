package main

import (
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/blackhole"
	"github.com/underscorenygren/partaj/pkg/http"
	"github.com/underscorenygren/partaj/pkg/types/optional"
	"os"
)

func main() {
	logger := logging.ConfigureDevelopment(os.Stderr)
	sink := blackhole.NewSink()
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
