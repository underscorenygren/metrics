package transformer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/blackhole"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/pipeline"
	"github.com/underscorenygren/partaj/pkg/programmatic"
	"github.com/underscorenygren/partaj/pkg/transformer"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

var _ = Describe("Transformer", func() {

	var testSource *programmatic.Source
	var total int
	var p *pipe.Pipe
	nEvents := 3
	eventBytes := []byte("a")

	logger := logging.ConfigureDevelopment(GinkgoWriter)

	BeforeEach(func() {
		testSource = programmatic.NewSource()
		total = 0

		for i := 0; i < nEvents; i++ {
			testSource.PutBytes(eventBytes)
		}
	})

	It("drained events are counted", func() {

		logger.Debug("***counted***")

		counter := func(evt *types.Event) (*types.Event, error) {
			total = total + 1
			logger.Debug("called counter fn", zap.ByteString("eventBytes", evt.Bytes()), zap.Int("total", total))
			return evt, nil
		}

		t, err := transformer.NewSource(testSource, counter)
		Expect(err).To(BeNil())

		//processing flow in goroutine, ensures we get expected error
		//use channel to handle concurrency
		p, err = pipe.Stage(t, blackhole.NewSink())
		Expect(err).To(BeNil())

		drained := pipeline.AsyncFlow(p)

		//After closing source, no more events are counted
		Expect(t.Close()).To(BeNil())
		Expect(testSource.PutString("a")).Should(HaveOccurred())
		err = <-drained
		Expect(err).To(Equal(errors.ErrSourceClosed))
		Expect(total).To(Equal(3))
	})

	It("uses a transformation to limit number of events", func(done Done) {
		max := 3
		limitErr := fmt.Errorf("limit")
		logger.Debug("***limit***")

		limitor := func(evt *types.Event) (*types.Event, error) {
			logger.Debug("got event", zap.ByteString("event_bytes", evt.Bytes()))
			if total >= 3 {
				return nil, limitErr
			}
			total = total + 1
			return evt, nil
		}

		t, err := transformer.NewSource(testSource, limitor)
		Expect(err).To(BeNil())

		p, err = pipe.Stage(t, blackhole.NewSink())
		Expect(err).To(BeNil())

		//Adds more event than max
		testSource.PutBytes(eventBytes)

		//Flow should end naturally from error in map fn
		err = p.Flow()
		Expect(err).To(Equal(limitErr))

		//should have only processed max events
		Expect(total).To(Equal(max))
		Expect(t.Close()).To(BeNil())

		close(done)
	})

	It("allows mapping to yield nil events", func(done Done) {
		max := 3
		mapErr := fmt.Errorf("done")
		logger.Debug("***nil events***")

		mapper := func(evt *types.Event) (*types.Event, error) {
			logger.Debug("got event", zap.ByteString("event_bytes", evt.Bytes()))
			if total >= 3 {
				return nil, mapErr
			}
			if total > 0 {
				evt = nil
			}
			total = total + 1
			return evt, nil
		}

		sink := buffer.NewSink()
		t, err := transformer.NewSource(testSource, mapper)
		Expect(err).To(BeNil())

		p, err = pipe.Stage(t, sink)
		Expect(err).To(BeNil())

		//Adds more event than max
		testSource.PutBytes(eventBytes)

		//Flow should end naturally from error in map fn
		err = p.Flow()
		Expect(err).To(Equal(mapErr))

		//should have only processed max events
		Expect(total).To(Equal(max))
		ref := []types.Event{
			types.NewEventFromBytes(eventBytes)}
		Expect(sink.Events).To(Equal(ref))

		close(done)
	})
})
