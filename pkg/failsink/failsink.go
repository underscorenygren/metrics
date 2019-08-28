package failsink

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

type failsink struct {
	sink types.Sink
	fail types.Sink
}

//Sink creates new failsink, which sends failed events to secondary sink
func Sink(sink types.Sink, fail types.Sink) (types.Sink, error) {
	if sink == nil {
		return nil, fmt.Errorf("sink cannot be nil")
	}

	if fail == nil {
		return nil, fmt.Errorf("failure sink cannot be nil")
	}

	return &failsink{
		sink: sink,
		fail: fail,
	}, nil
}

//Drain sends failed events to a secondary sink
func (fs *failsink) Drain(events []types.Event) []error {

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
