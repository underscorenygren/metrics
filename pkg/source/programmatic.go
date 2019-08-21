package source

import (
	"github.com/underscorenygren/metrics/pkg/types"
)

//ProgrammaticSource a source that accept events programatically, through calls to Put
type ProgrammaticSource struct {
	c      chan *types.Event
	closed bool
}

//NewProgrammaticSource creates a ProgrammaticSource
func NewProgrammaticSource() *ProgrammaticSource {
	return &ProgrammaticSource{
		c:      make(chan *types.Event),
		closed: false,
	}
}

//PutBytes Constructs an event from bytes and adds it to the source
func (manual *ProgrammaticSource) PutBytes(bytes []byte) error {
	return manual.Put(types.NewEventFromBytes(bytes))
}

//PutString Constructs an event from a string and adds it to the source
func (manual *ProgrammaticSource) PutString(str string) error {
	return manual.PutBytes([]byte(str))
}

//Put Adds an event to the source
func (manual *ProgrammaticSource) Put(event types.Event) error {
	if manual.closed {
		return ErrSourceClosed
	}

	manual.c <- &event

	return nil
}

//DrawOne draws an event from the source
func (manual *ProgrammaticSource) DrawOne() (*types.Event, error) {
	if manual.closed {
		return nil, ErrSourceClosed
	}
	return <-manual.c, nil
}

//Close Closes the source, disallowing further events
func (manual *ProgrammaticSource) Close() error {
	close(manual.c)
	manual.closed = true
	return nil
}
