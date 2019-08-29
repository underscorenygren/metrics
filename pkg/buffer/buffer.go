/*
Package buffer is an event sink that stores events in an
internally allocatted infinitely extending slice.

Primarily used in testing to easily check correct events are
received/processed.
*/
package buffer

import (
	"github.com/underscorenygren/partaj/pkg/types"
)

//Sink fulfills Sink interface
type Sink struct {
	Events []types.Event //the events received
}

/*
NewSink creates a new Buffer Sink.

Returns a pointer, because drain manipulates the
event array internally.
*/
func NewSink() *Sink {
	return &Sink{
		Events: []types.Event{}}
}

/*
Drain appends all events received to the internal buffer.

Does no range checking, so will always return a nil error,
and will crash if memory is not available
*/
func (buffer *Sink) Drain(events []types.Event) []error {
	for _, evt := range events {
		buffer.Events = append(buffer.Events, evt)
	}
	return nil
}
