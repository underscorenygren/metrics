package blackhole_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/underscorenygren/partaj/internal"
	"github.com/underscorenygren/partaj/pkg/blackhole"
	"github.com/underscorenygren/partaj/pkg/types"
)

func Example() {
	blackholeSink := blackhole.NewSink()
	evts := []types.Event{
		types.NewEventFromBytes([]byte{1})}
	if blackholeSink.Drain(evts) == nil {
		fmt.Printf("nil")
	}
	//Output: nil
}

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
