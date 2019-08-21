package pipeline

import (
	"github.com/underscorenygren/metrics/pkg/types"
)

//Pipeline A pipeline is the plumbing that connects Sources to Drains,
//and registers processing on events passing through it.
type Pipeline struct {
	MapFn   types.MapFn
	source  types.Source
	drainTo types.Sink
}

//defaultMapFn the default map function passes events through unchanged
func defaultMapFn(evt *types.Event) *types.Event {
	return evt
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

	for {
		e, err := pipe.source.DrawOne()
		if err != nil || e == nil {
			return err
		}
		e = pipe.MapFn(e)
		if e != nil {
			pipe.drainTo.Drain([]types.Event{*e})
		}
	}

	return nil
}
