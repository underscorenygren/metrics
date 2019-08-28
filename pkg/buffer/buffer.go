package buffer

import (
	"github.com/underscorenygren/partaj/pkg/types"
)

//Buffer stores drained event in memory
type Buffer struct {
	Events []types.Event
}

//Sink creates a new buffer sink
func Sink() *Buffer {
	return &Buffer{
		Events: []types.Event{}}
}

//Drain drains events into internal buffer
func (buffer *Buffer) Drain(events []types.Event) []error {
	for _, evt := range events {
		buffer.Events = append(buffer.Events, evt)
	}
	return nil
}
