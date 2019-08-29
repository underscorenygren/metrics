/*
Package main is a simple webserver that discard all events
and logs it's internal processing.
*/
package main

import (
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/blackhole"
	"github.com/underscorenygren/partaj/pkg/http"
	"github.com/underscorenygren/partaj/pkg/types/optional"
	"os"
)

//Example runs the example webserver.
func Example() {
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

func main() {
	Example()
}
