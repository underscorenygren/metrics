package blackhole_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/underscorenygren/partaj/internal"
	"github.com/underscorenygren/partaj/pkg/blackhole"
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
