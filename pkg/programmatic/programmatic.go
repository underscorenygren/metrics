/*
Package programmatic provides a Source that can receive events using Put function calls.
*/
package programmatic

import (
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//ChannelBufferSize is the max number of events that can be in flight.
const ChannelBufferSize = 100000

/*
Source fullfills the Source interface. It accepts events using Put commands,
and stores them in an internal channel.
*/
type Source struct {
	c      chan *types.Event
	closed bool
}

//NewSource creates a new programmatic Source.
func NewSource() *Source {
	return &Source{
		c:      make(chan *types.Event, ChannelBufferSize),
		closed: false,
	}
}

//PutBytes constructs a new event from bytes and adds it to the source.
func (manual *Source) PutBytes(bytes []byte) error {
	return manual.Put(types.NewEventFromBytes(bytes))
}

//PutString constructs an event from a string and adds it to the source.
func (manual *Source) PutString(str string) error {
	return manual.PutBytes([]byte(str))
}

/*
Put adds an event to the source.

Returns an error if source has been closed,
or if the underlying channel is broken.
*/
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

/*
DrawOne draws one event from the underlying channel.
*/
func (manual *Source) DrawOne() (*types.Event, error) {
	logger := logging.Logger()

	logger.Debug("DrawOne: blocking on channel")
	e, more := <-manual.c
	if e != nil {
		logger.Debug("DrawOne: unblocked from channel", zap.ByteString("event", e.Bytes()))
	} else {
		logger.Debug("DrawOne: unblocked from channel")
	}

	var err error
	if !more {
		logger.Debug("DrawOne: channel closed")
		err = errors.ErrSourceClosed
	}

	return e, err
}

//Close closes the source, the underlying channel, causing further puts to error.
func (manual *Source) Close() error {
	close(manual.c)
	manual.closed = true
	return nil
}
