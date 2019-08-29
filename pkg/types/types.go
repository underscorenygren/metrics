/*
Package types contains generic types and interfaces for partaj.
*/
package types

/*
Event represents a single event in the system.

Declared as a type to make function signatures generic, and
so that future additions and changes can be made more easily.
*/
type Event struct {
	bytes []byte //the raw bytes of the event
}

/*
Sink defines an interface for where events end up.
*/
type Sink interface {
	/*
		Drain processes a sequence of events.

		If there are any failures, must return a slice of errors, where
		the index of the failed event corresponds with it's index in the error array.

		If there are no errors, returns nil
	*/
	Drain([]Event) []error
}

/*
Source defines an interface for where events originate.
*/
type Source interface {
	/*
		DrawOne draws a single event from the source.

		Expected to block until an event is available, or an error occurs.

		Can return nil events, which should be ignored if there are no errors.
	*/
	DrawOne() (*Event, error)
	/* Close closes a source and any of it's associated resources */
	Close() error
}

/*
Stage defines a connection of sources and sinks.
*/
type Stage interface {
	/*
		Flow pumps events through the stage continuously by calling
		Draw and Drain methods, until they return an error.
	*/
	Flow() error
}

//NewEventFromBytes is a convenience method for creating an Event from a set of bytes.
func NewEventFromBytes(bytes []byte) Event {
	return Event{bytes: bytes}
}

//Bytes returns the underlying bytes.
func (evt *Event) Bytes() []byte {
	return evt.bytes
}

//NewBytes creates a copy of the event with the new bytes.
func (evt *Event) NewBytes(bytes []byte) *Event {
	cpy := Event(*evt)
	cpy.bytes = bytes
	return &cpy
}
