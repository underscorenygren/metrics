package http_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/underscorenygren/metrics/pkg/http"
	"github.com/underscorenygren/metrics/pkg/sink/blackhole"
)

var _ = Describe("Http", func() {
	It("creates a server", func() {
		srv, err := http.NewServer(http.Config{
			Sink: blackhole.Sink()})
		Expect(err).To(BeNil())
		Expect(srv).ToNot(BeNil())
	})
})
