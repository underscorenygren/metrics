package pipeline_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/sink/blackhole"
	"github.com/underscorenygren/metrics/pkg/source"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

var _ = Describe("Pipeline", func() {
	logger := logging.ConfigureDevelopment(GinkgoWriter)

	It("drained events are counted", func() {

		testSource := source.NewProgrammaticSource()

		total := 0
		counter := func(evt *types.Event) (*types.Event, error) {
			total = total + 1
			logger.Debug("called counter fn", zap.ByteString("eventBytes", evt.Bytes()), zap.Int("total", total))
			return evt, nil
		}

		p := pipeline.NewPipeline(
			testSource,
			blackhole.Sink())

		p.MapFn = counter
		//processing flow in goroutine, ensures we get expected error
		//use channel to handle concurrency
		drained := p.AsyncFlow()

		//Flow takes care of the draining
		testSource.PutString("a")
		testSource.PutString("a")
		testSource.PutString("a")

		//After closing source, no more events are counted
		Expect(testSource.Close()).To(BeNil())
		Expect(testSource.PutString("a")).Should(HaveOccurred())
		err := <-drained
		Expect(err).To(Equal(source.ErrSourceClosed))
		Expect(total).To(Equal(3))
	})

	It("uses map to limit number of processed events", func(done Done) {

		logger.Debug("starting limit test")
		testSource := source.NewProgrammaticSource()

		total := 0
		max := 3
		limitErr := errors.New("limit")

		limitor := func(evt *types.Event) (*types.Event, error) {
			logger.Debug("got event", zap.ByteString("event_bytes", evt.Bytes()))
			if total >= 3 {
				return nil, limitErr
			}
			total = total + 1
			return evt, nil
		}

		p := pipeline.NewPipeline(
			testSource,
			blackhole.Sink())
		p.MapFn = limitor

		//Adds more event than max
		for i := 0; i < max+1; i++ {
			testSource.PutString("a")
		}

		//Flow should end naturally from error in map fn
		err := p.Flow()
		Expect(err).To(Equal(limitErr))

		//should have only processed max events
		Expect(total).To(Equal(max))
		Expect(testSource.Close()).To(BeNil())

		close(done)
	})
})
