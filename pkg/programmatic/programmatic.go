package programmatic

import (
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/errors"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

//ChannelBufferSize the max number of events that can be in flight
const ChannelBufferSize = 100000

//Source a source that accept events programatically, through calls to Put
type Source struct {
	c      chan *types.Event
	closed bool
}

//NewSource creates a Source
func NewSource() *Source {
	return &Source{
		c:      make(chan *types.Event, ChannelBufferSize),
		closed: false,
	}
}

//PutBytes Constructs an event from bytes and adds it to the source
func (manual *Source) PutBytes(bytes []byte) error {
	return manual.Put(types.NewEventFromBytes(bytes))
}

//PutString Constructs an event from a string and adds it to the source
func (manual *Source) PutString(str string) error {
	return manual.PutBytes([]byte(str))
}

//Put Adds an event to the source
func (manual *Source) Put(event types.Event) error {
	if manual.closed {
		return errors.ErrSourceClosed
	}
	logger := logging.Logger()

	logger.Debug("Put: pre-channel send")
	e := &event
	select {
	case manual.c <- &event:
		logger.Debug("Put: channel success", zap.ByteString("event", e.Bytes()))
	default:
		logger.Debug("Put: channel fail")
		return errors.ErrChannelBroken
	}

	return nil
}

//DrawOne draws an event from the source
func (manual *Source) DrawOne() (*types.Event, error) {
	logger := logging.Logger()

	logger.Debug("DrawOne: blocking on channel")
	e, more := <-manual.c
	logger.Debug("DrawOne: unblocked from channel")

	var err error
	if !more {
		logger.Debug("DrawOne: channel closed")
		err = errors.ErrSourceClosed
	}

	return e, err
}

//Close Closes the source, disallowing further events
func (manual *Source) Close() error {
	close(manual.c)
	manual.closed = true
	return nil
}
