package pipeline_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/blackhole"
	"github.com/underscorenygren/metrics/pkg/buffer"
	"github.com/underscorenygren/metrics/pkg/errors"
	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/programmatic"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

//failSink fails after certain number of events
type failSink struct {
	received  int
	failAfter int
	logger    *zap.Logger
}

//implements Sink interface
func (fail *failSink) Drain(events []types.Event) []error {
	errs := []error{}
	failures := false
	for range events {
		fail.received = fail.received + 1
		if fail.received > fail.failAfter {
			failures = true
			fail.logger.Debug("failSink: failing", zap.Int("received", fail.received), zap.Int("failAfter", fail.failAfter))
			errs = append(errs, fmt.Errorf("failureSink %d > %d", fail.received, fail.failAfter))
		} else {
			fail.logger.Debug("failSink: not failing", zap.Int("received", fail.received))
			errs = append(errs, nil)
		}
	}
	if failures {
		return errs
	}
	return nil
}

var _ = Describe("Pipeline", func() {
	logger := logging.ConfigureDevelopment(GinkgoWriter)

	var testSource *programmatic.Source
	var total int
	var p *pipeline.Pipeline
	nEvents := 3
	eventBytes := []byte("a")

	BeforeEach(func() {
		testSource = programmatic.NewSource()
		total = 0
		p = pipeline.NewPipeline(testSource, blackhole.Sink())

		for i := 0; i < nEvents; i++ {
			testSource.PutBytes(eventBytes)
		}
	})

	It("drained events are counted", func() {

		counter := func(evt *types.Event) (*types.Event, error) {
			total = total + 1
			logger.Debug("called counter fn", zap.ByteString("eventBytes", evt.Bytes()), zap.Int("total", total))
			return evt, nil
		}

		p.MapFn = counter
		//processing flow in goroutine, ensures we get expected error
		//use channel to handle concurrency
		drained := p.AsyncFlow()

		//After closing source, no more events are counted
		Expect(testSource.Close()).To(BeNil())
		Expect(testSource.PutString("a")).Should(HaveOccurred())
		err := <-drained
		Expect(err).To(Equal(errors.ErrSourceClosed))
		Expect(total).To(Equal(3))
	})

	It("uses map to limit number of processed events", func(done Done) {

		logger.Debug("starting limit test")

		max := 3
		limitErr := fmt.Errorf("limit")

		limitor := func(evt *types.Event) (*types.Event, error) {
			logger.Debug("got event", zap.ByteString("event_bytes", evt.Bytes()))
			if total >= 3 {
				return nil, limitErr
			}
			total = total + 1
			return evt, nil
		}

		p.MapFn = limitor

		//Adds more event than max
		testSource.PutBytes(eventBytes)

		//Flow should end naturally from error in map fn
		err := p.Flow()
		Expect(err).To(Equal(limitErr))

		//should have only processed max events
		Expect(total).To(Equal(max))
		Expect(testSource.Close()).To(BeNil())

		close(done)
	})

	It("handles failed events", func(done Done) {
		//sink will fail on last event
		failer := &failSink{failAfter: nEvents - 1, logger: logger}
		ref := []types.Event{
			types.NewEventFromBytes(eventBytes)}

		p = pipeline.NewPipeline(testSource, failer)
		//track failures in buffer
		failed := buffer.Sink()
		p.FailSink = failed

		//run pipeline
		Expect(testSource.Close()).To(BeNil())
		Expect(p.Flow()).To(Equal(errors.ErrSourceClosed))

		//last event should end up in failure sink
		Expect(failed.Events).To(Equal(ref))

		close(done)
	})
})
