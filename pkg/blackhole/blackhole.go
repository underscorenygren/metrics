/*Package blackhole is an event sink that discards all events it receives. */
package blackhole

import (
	"github.com/underscorenygren/partaj/pkg/types"
)

//Drain discards all events it receives and never fails. Can be used to easily fulfill Sink interface.
func Drain(evts []types.Event) []error {
	return nil
}

//Sink fulfills the Sink interface and discards all events it receives.
type Sink struct{}

//NewSink creates a Sink that discards all events it receives.
func NewSink() Sink {
	return Sink{}
}

//Drain fulfills the Sink interface for the blackhole.
func (sink Sink) Drain(evts []types.Event) []error {
	return Drain(evts)
}
