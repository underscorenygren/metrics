package source

import (
	"errors"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

//ErrChannelBroken when we can't put on the channel
var ErrChannelBroken = errors.New("channel broken")

//ChannelBufferSize the max number of events that can be in flight
const ChannelBufferSize = 100000

//ProgrammaticSource a source that accept events programatically, through calls to Put
type ProgrammaticSource struct {
	c      chan *types.Event
	closed bool
}

//NewProgrammaticSource creates a ProgrammaticSource
func NewProgrammaticSource() *ProgrammaticSource {
	return &ProgrammaticSource{
		c:      make(chan *types.Event, ChannelBufferSize),
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
	logger := logging.Logger()

	logger.Debug("Put: pre-channel send")
	e := &event
	select {
	case manual.c <- &event:
		logger.Debug("Put: channel success", zap.ByteString("event", e.Bytes()))
	default:
		logger.Debug("Put: channel fail")
		return ErrChannelBroken
	}

	return nil
}

//DrawOne draws an event from the source
func (manual *ProgrammaticSource) DrawOne() (*types.Event, error) {
	logger := logging.Logger()

	logger.Debug("DrawOne: blocking on channel")
	e, more := <-manual.c
	logger.Debug("DrawOne: unblocked from channel")

	var err error
	if !more {
		logger.Debug("DrawOne: channel closed")
		err = ErrSourceClosed
	}

	return e, err
}

//Close Closes the source, disallowing further events
func (manual *ProgrammaticSource) Close() error {
	close(manual.c)
	manual.closed = true
	return nil
}
