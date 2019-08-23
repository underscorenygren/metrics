package pipeline

import (
	"fmt"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/blackhole"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

//Pipeline A pipeline is the plumbing that connects Sources to Drains,
//and registers processing on events passing through it.
type Pipeline struct {
	MapFn   types.MapFn
	source  types.Source
	drainTo types.Sink
	//Failed drains go to a separate sink, a blackhole by default.
	FailSink types.Sink
}

//defaultMapFn the default map function passes events through unchanged
func defaultMapFn(evt *types.Event) (*types.Event, error) {
	return evt, nil
}

//NewPipeline Creates a Pipeline
func NewPipeline(source types.Source, sink types.Sink) *Pipeline {
	return &Pipeline{
		source:   source,
		MapFn:    defaultMapFn,
		drainTo:  sink,
		FailSink: blackhole.Sink(),
	}
}

//Flow the executing loop that draws events from the source,
//processes and drains them. Ends when Source is closed, and
//blocks on internal channels, so best run as a goroutine.
func (pipe *Pipeline) Flow() error {

	logger := logging.Logger()
	for {
		logger.Debug("Flow: drawing")
		e, err := pipe.source.DrawOne()
		logger.Debug("Flow: drew")

		//drawing errors stop the flow
		if err != nil {
			logger.Debug("Flow: Draw error", zap.Error(err))
			return err
		}
		//nil events also stop the flow
		if e == nil {
			logger.Debug("Flow: Empty event")
			return nil
		}
		//map functions can return errors to stop flow
		logger.Debug("Flow: mapping")
		e, err = pipe.MapFn(e)
		if err != nil {
			logger.Debug("Flow: mapping error", zap.Error(err))
			return err
		}
		//mappers can return nil events, but this doesn't stop flow
		if e != nil {
			logger.Debug("Flow: draining", zap.ByteString("event", e.Bytes()))
			events := []types.Event{*e}
			err = pipe.drain(events)
			//we cannot recover from failures to drain to failure sink
			if err != nil {
				return err
			}
		} else {
			logger.Debug("Flow: mapped nil")
		}
	}

	return nil
}

//internal drain handling function that handles draining failures
func (pipe *Pipeline) drain(events []types.Event) error {
	logger := logging.Logger()

	failed := pipe.drainTo.Drain(events)

	if failed == nil {
		return nil
	}

	failures := []types.Event{}
	for i, failErr := range failed {
		if failErr != nil {
			e := events[i]
			logger.Debug("Flow: event failed",
				zap.Int("i", i),
				zap.Error(failErr),
				zap.ByteString("event", e.Bytes()))
			failures = append(failures, e)
		}
	}

	if len(failures) == 0 {
		return nil
	}

	if errs := pipe.FailSink.Drain(failures); errs != nil {
		return fmt.Errorf("received error on failure drain: %v", errs)
	}

	return nil
}

/*AsyncFlow calls flow in a goroutine and retuns a channel
 * with the result of the flow when it ends
 */
func (pipe *Pipeline) AsyncFlow() <-chan error {
	drained := make(chan error)
	logger := logging.Logger()
	go func() {
		logger.Debug("starting async flow")
		drained <- pipe.Flow()
		logger.Debug("ended async flow")
	}()
	return drained
}

//Source gets the source for the pipeline
func (pipe *Pipeline) Source() types.Source {
	return pipe.source
}

//Sink gets the sink for the pipeline
func (pipe *Pipeline) Sink() types.Sink {
	return pipe.drainTo
}
