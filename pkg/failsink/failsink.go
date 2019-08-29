/*
Package failsink is a sink that drains all events arriving to
it into it's primary sink, and re-drains all failed events into
the secondary sink.
*/
package failsink

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//Sink implements the Sink interface.
type Sink struct {
	sink types.Sink
	fail types.Sink
}

/*
NewSink creates a Sink which attempts to drain all
events to the primary "sink" Sink, and re-drains
failed events to secondary "fail" sink.

Will fail if either sinks provided are nil.
*/
func NewSink(sink types.Sink, fail types.Sink) (*Sink, error) {
	if sink == nil {
		return nil, fmt.Errorf("sink cannot be nil")
	}

	if fail == nil {
		return nil, fmt.Errorf("failure sink cannot be nil")
	}

	return &Sink{
		sink: sink,
		fail: fail,
	}, nil
}

//Drain sends failed events to the configured "fail" sink.
func (fs *Sink) Drain(events []types.Event) []error {

	logger := logging.Logger()

	failed := fs.sink.Drain(events)

	if failed == nil {
		return nil
	}

	failures := []types.Event{}
	for i, failErr := range failed {
		if failErr != nil {
			e := events[i]
			logger.Debug("failSink.Drain: event failed",
				zap.Int("i", i),
				zap.Error(failErr),
				zap.ByteString("event", e.Bytes()))
			failures = append(failures, e)
		}
	}

	if len(failures) == 0 {
		return nil
	}

	failed = fs.fail.Drain(failures)
	if failed != nil {
		logger.Debug("failsink.Drain: failure sink returned errors")
	}

	return failed
}
