package failsink_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/blackhole"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/failsink"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/programmatic"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//Example is just used to show the specs in godoc
func Example() {}

//failSink is a dummy sink that starts returning errors
//after a certain number of events have been received.
type failSink struct {
	received  int
	failAfter int
	logger    *zap.Logger
}

//Drain implements the Sink interface
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

var _ = Describe("Failsink", func() {

	var testSource *programmatic.Source
	var p *pipe.Pipe
	var err error
	nEvents := 3
	eventBytes := []byte("a")

	logger := logging.ConfigureDevelopment(GinkgoWriter)

	BeforeEach(func() {
		testSource = programmatic.NewSource()
		p, err = pipe.Stage(testSource, blackhole.NewSink())
		Expect(err).To(BeNil())

		for i := 0; i < nEvents; i++ {
			testSource.PutBytes(eventBytes)
		}
	})

	It("handles failed events", func(done Done) {
		//sink will fail on last event
		failer := &failSink{failAfter: nEvents - 1, logger: logger}
		ref := []types.Event{
			types.NewEventFromBytes(eventBytes)}

		//track failures in buffer
		failed := buffer.NewSink()
		fs, err := failsink.NewSink(failer, failed)
		Expect(err).To(BeNil())
		p, err = pipe.Stage(testSource, fs)
		Expect(err).To(BeNil())

		//run pipe
		Expect(testSource.Close()).To(BeNil())
		Expect(p.Flow()).To(Equal(errors.ErrSourceClosed))

		//last event should end up in failure sink
		Expect(failed.Events).To(Equal(ref))

		close(done)
	})
})
