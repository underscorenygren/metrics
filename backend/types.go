package sink

import (
	"github.com/underscorenygren/metrics/event"
)

//Sink A generic class for capturing all implementations
// where events end up and are processed.
// Specific implementations of the Sink interface provide
// functionality for different "backends", such as sending events
// to storage, logging, etc.
type Sink interface {
	//Processes a sequence of events.
	//returns error if any events fail,
	//and an error array where the index of the error
	//corresponds to the failed event
	Drain([]event.Event) (error, []error)
}
