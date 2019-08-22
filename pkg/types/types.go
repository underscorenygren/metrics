package types

import "context"

//Event a single event in the system. Passed around as a typed
//object instead of a native type to make additions to class easier.
type Event struct {
	//the raw bytes of the event, unprocessed
	bytes []byte
	//stores context for event
	ctx *context.Context
}

//Sink A generic class for capturing all implementations
// where events end up and are processed.
// Specific implementations of the Sink interface provide
// functionality for different "backends", such as sending events
// to storage, logging, etc.
type Sink interface {
	//Processes a sequence of events.
	//If any events fail, returns a slice where the index of the failure corresponds to the original event.
	//If there are no failures, returns nil
	Drain([]Event) []error
}

//MapFn mapfn
type MapFn func(*Event) (*Event, error)

//Source source
type Source interface {
	DrawOne() (*Event, error)
	Close() error
}

//Pipeline pipeline
type Pipeline interface {
	Source(Source)
	Map(MapFn)
	DrainTo(Sink)
}

//NewEventFromBytes creates an event object from a set of bytes
func NewEventFromBytes(bytes []byte) Event {
	return Event{bytes: bytes}
}

//Bytes returns the bytes of the event
func (evt *Event) Bytes() []byte {
	return evt.bytes
}

//Context gets the event context
func (evt *Event) Context() *context.Context {
	return evt.ctx
}
