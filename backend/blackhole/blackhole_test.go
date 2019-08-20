package blackhole_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "UNKNOWN_PACKAGE_PATH"
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
			ToEqual(nil)
	})

	It("drained events are counted", func() {

		testSource := pkg.NewManualSource()

		total := 0
		counter := func(evt Event) Event {
			total = total + 1
			return evt
		}

		p := pkg.NewChannelPipeline(
			testSource,
			blackhole.Sink())
		p.MapFn = counter
		go p.Flow()

		//Flow takes care of the draining
		testSource.PutString("a")
		Expect(total).To(Equal(1))

		//more events get counted
		testSource.PutString("a")
		testSource.PutString("a")
		Expect(counter.total).To(Equal(3))

		//After closing source, no more events are counted
		Expect(testSource.Close()).To(BeNil())
		Expect(testSource.PutString("a")).To(BeError())
		Expect(counter.total).To(Equal(3))
	})
})
