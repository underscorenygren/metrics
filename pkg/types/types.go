package pkg

//Event a single event in the system. Passed around as a typed
//object instead of a native type to make additions to class easier.
type Event struct {
	bytes []byte
}

//Sink A generic class for capturing all implementations
// where events end up and are processed.
// Specific implementations of the Sink interface provide
// functionality for different "backends", such as sending events
// to storage, logging, etc.
type Sink interface {
	//Processes a sequence of events.
	//if any events fail, returns a slice where
	//the index of the failure corresponds to the original
	//event.
	//If there are no failures, returns nil
	Drain([]event.Event) []error
}

//Stage stage
type Stage interface{}

//MapFn mapfn
type MapFn func(Event) Event

//Pipeline pipeline
type Pipeline interface {
	Source(Source)
	Map(MapFn)
	DrainTo(Sink)
}

//Source source
type Source interface {
	DrawOne() Event
	Close() error
}
