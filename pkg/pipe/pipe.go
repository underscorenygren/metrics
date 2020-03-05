/*
Package pipe provides a simple stage that connects a source with a sink.
*/
package pipe

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/internal/stage"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//Pipe connects a source to a sink, by sending all events from source to the sink. Implements Stage interface.
type Pipe struct {
	source types.Source
	sink   types.Sink
}

//NewStage creates a Pipe that connects a source with a sink.
func NewStage(source types.Source, sink types.Sink) (*Pipe, error) {
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

/*
Flow implements the Stage interface. Continually draws one event from
the source and sends it to the sink.

Runs continually until sink.Drain returns an error.
*/
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
}

//Source gets the source for the pipe.
func (pipe *Pipe) Source() types.Source {
	return pipe.source
}

//Sink gets the sink for the pipe.
func (pipe *Pipe) Sink() types.Sink {
	return pipe.sink
}
