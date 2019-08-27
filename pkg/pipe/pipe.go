package pipe

import (
	"fmt"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/internal/stage"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

//Pipe A pipeline is the plumbing that connects Sources to Drains,
//and registers processing on events passing through it.
type Pipe struct {
	source types.Source
	sink   types.Sink
}

//Stage creates a pipe that links a source and a sink
func Stage(source types.Source, sink types.Sink) (*Pipe, error) {
	if source == nil {
		return nil, fmt.Errorf("source cannot be nil")
	}
	if sink == nil {
		return nil, fmt.Errorf("sink cannot be nil")
	}

	return &Pipe{
		source: source,
		sink:   sink,
	}, nil
}

//Flow the executing loop that draws events from the source,
//processes and drains them. Ends when Source is closed, and
//blocks on internal channels, so best run as a goroutine.
func (pipe *Pipe) Flow() error {

	logger := logging.Logger()
	for {
		logger.Debug("pipe.Flow: drawing")
		e, err := pipe.source.DrawOne()
		logger.Debug("pipe.Flow: drew")

		//drawing errors stop the flow
		if err != nil {
			logger.Debug("pipe.Flow: Draw error", zap.Error(err))
			return err
		}
		//nil events dont stop the flow. This allows
		//things like mappers nullifying events, or other
		//types of event filtering
		if e == nil {
			logger.Debug("pipe.Flow: Empty event")
		} else {
			logger.Debug("pipe.Flow: draining", zap.ByteString("event", e.Bytes()))
			events := []types.Event{*e}
			errs := pipe.sink.Drain(events)
			logger.Debug("pipe.Flow: drained")
			//End flow iff errors from drain
			if err = stage.FlattenErrors(errs, logger); err != nil {
				return err
			}
		}
	}

	return nil
}

//Source gets the source for the pipeline
func (pipe *Pipe) Source() types.Source {
	return pipe.source
}

//Sink gets the sink for the pipeline
func (pipe *Pipe) Sink() types.Sink {
	return pipe.sink
}
