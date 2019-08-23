package blackhole

import (
	"github.com/underscorenygren/metrics/pkg/types"
)

//Drain discard all events
func Drain(evts []types.Event) []error {
	return nil
}

//blackholeSink internal-only struct to register Drain method on a Sink object
type blackholeSink struct{}

//Sink creates a new sink that discards all events received
func Sink() types.Sink {
	return blackholeSink{}
}

//Drain the drain method on the blackhole sink. Discards all events
func (sink blackholeSink) Drain(evts []types.Event) []error {
	return Drain(evts)
}
