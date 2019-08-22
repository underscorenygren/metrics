package pipeline

import (
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

//Pipeline A pipeline is the plumbing that connects Sources to Drains,
//and registers processing on events passing through it.
type Pipeline struct {
	MapFn   types.MapFn
	source  types.Source
	drainTo types.Sink
}

//defaultMapFn the default map function passes events through unchanged
func defaultMapFn(evt *types.Event) (*types.Event, error) {
	return evt, nil
}

//NewPipeline Creates a Pipeline
func NewPipeline(source types.Source, sink types.Sink) *Pipeline {
	return &Pipeline{
		source:  source,
		MapFn:   defaultMapFn,
		drainTo: sink,
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
		//nil events and errors stop the flow
		if err != nil {
			logger.Debug("Flow: Draw error", zap.Error(err))
			return err
		}
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
		if e != nil {
			logger.Debug("Flow: draining", zap.ByteString("event", e.Bytes()))
			pipe.drainTo.Drain([]types.Event{*e})
		} else {
			logger.Debug("Flow: mapped nil")
		}
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
