package http_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"context"
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/http"
	"github.com/underscorenygren/partaj/pkg/types"
	"github.com/underscorenygren/partaj/pkg/types/optional"
	"go.uber.org/zap"
	nethttp "net/http"
	"time"
)

func shutdown(s *http.Server) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(800)*time.Millisecond)
	defer cancel()
	s.Shutdown(ctx)
}

var _ = Describe("Http", func() {
	var s *http.Server
	var bufferSink *buffer.Buffer
	var err error
	testPort := 10329
	testHost := "127.0.0.1"
	testURL := func(p int) string {
		return fmt.Sprintf("http://%s:%d/", testHost, p)
	}
	testBytes := []byte("hello world")
	ref := types.NewEventFromBytes(testBytes)
	startupTime := 200 * time.Millisecond

	logger := logging.ConfigureDevelopment(GinkgoWriter)

	It("server handles events", func(done Done) {
		bufferSink = buffer.Sink()
		s, err = http.NewServer(http.Config{
			Port: optional.Int(testPort),
			Host: optional.String(testHost),
			Sink: bufferSink,
		})
		Expect(err).ToNot(HaveOccurred())

		go s.ListenAndServe()
		//We use sleeps to let things progress through pipeline, but should probably do better here
		time.Sleep(startupTime)

		//queue a few events
		for i := 0; i < 2; i++ {
			url := testURL(testPort)
			logger.Debug("posting to", zap.String("url", url))
			resp, err := nethttp.Post(url, "text/plain", bytes.NewReader(testBytes))
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.DefaultSuccessCode))
		}
		time.Sleep(startupTime)

		//events are there
		Expect(bufferSink.Events).To(Equal([]types.Event{ref, ref}))
		shutdown(s)
		close(done)
	})

	It("routes two different request to different sinks", func(done Done) {
		bufferSink = buffer.Sink()
		port := testPort + 1
		s, err = http.NewServer(http.Config{
			Port: optional.Int(port),
			Host: optional.String(testHost),
			Sink: bufferSink,
		})
		Expect(err).ToNot(HaveOccurred())
		otherSink := buffer.Sink()
		otherSuffix := "other"
		//adds different sink to route
		s.Router.
			PathPrefix(fmt.Sprintf("/%s", otherSuffix)).
			Subrouter().
			HandleFunc("/", s.MakeHandleFunc(otherSink))

		go s.ListenAndServe()
		time.Sleep(startupTime)

		//put an event on each route
		for _, suffix := range []string{"", "other/"} {
			url := fmt.Sprintf("%s%s", testURL(port), suffix)
			logger.Debug("posting to", zap.String("url", url))
			resp, err := nethttp.Post(url, "text/plain", bytes.NewReader(testBytes))
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.DefaultSuccessCode))
		}
		//We use sleeps to let things progress through pipeline, but should probably do better here
		time.Sleep(startupTime)

		Expect(bufferSink.Events).To(Equal([]types.Event{ref}))
		Expect(otherSink.Events).To(Equal([]types.Event{ref}))
		shutdown(s)
		close(done)
	})
})
