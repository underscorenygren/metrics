package pipeline_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/sink/blackhole"
	"github.com/underscorenygren/metrics/pkg/source"
	"github.com/underscorenygren/metrics/pkg/types"
)

var _ = Describe("Pipeline", func() {

	It("drained events are counted", func() {

		testSource := source.NewProgrammaticSource()

		total := 0
		counter := func(evt *types.Event) *types.Event {
			total = total + 1
			return evt
		}

		p := pipeline.NewPipeline(
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
		Expect(total).To(Equal(3))

		//After closing source, no more events are counted
		Expect(testSource.Close()).To(BeNil())
		Expect(testSource.PutString("a")).Should(HaveOccurred())
		Expect(total).To(Equal(3))
	})
})
