package http_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"context"
	"fmt"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/http"
	"github.com/underscorenygren/metrics/pkg/optional"
	"github.com/underscorenygren/metrics/pkg/sink/buffer"
	"github.com/underscorenygren/metrics/pkg/types"
	nethttp "net/http"
	"time"
)

var _ = Describe("Http", func() {
	var s *http.Server
	var bufferSink *buffer.Buffer
	var err error
	testPort := 10329
	testHost := "127.0.0.1"
	testURL := fmt.Sprintf("http://%s:%d/", testHost, testPort)
	testBytes := []byte("hello world")
	//shutdown is closure to make easy to call from tests
	shutdown := func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(800)*time.Millisecond)
		defer cancel()
		s.Shutdown(ctx)
	}
	logging.ConfigureDevelopment(GinkgoWriter)

	BeforeEach(func() {
		bufferSink = buffer.Sink()
		s, err = http.NewServer(http.Config{
			Port: optional.Int(testPort),
			Host: optional.String(testHost),
			Sink: bufferSink,
		})
		Expect(err).ToNot(HaveOccurred())

		go s.ListenAndServe()
	})

	It("server handles events", func(done Done) {
		ref := types.NewEventFromBytes(testBytes)
		//queue a few events
		for i := 0; i < 2; i++ {
			resp, err := nethttp.Post(testURL, "text/plain", bytes.NewReader(testBytes))
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.DefaultSuccessCode))
		}

		//events are there
		Expect(bufferSink.Events).To(Equal([]types.Event{ref, ref}))
		shutdown()
		close(done)
	})

	It("a second server started doesn't fail", func(done Done) {
		shutdown()
		close(done)
	})
})
