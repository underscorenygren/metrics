package blackhole_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/underscorenygren/metrics/internal"
	"github.com/underscorenygren/metrics/pkg/sink/blackhole"
)

var _ = Describe("Blackhole", func() {

	testEvents := [][]byte{
		[]byte("a"),
		[]byte("b"),
	}

	It("drains test events", func() {

		Expect(
			blackhole.Drain(
				internal.ToEvents(testEvents))).
			To(BeNil())
	})
})
