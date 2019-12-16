/*
Package stream provides a source that reads events from an io.Reader.

bufio.Scanner is exposed, which can be used to configure stream parsing:
https://golang.org/pkg/bufio/#Scanner
*/
package stream

import (
	"bufio"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/types"
	"io"
)

/*
Source implements the Source interface.

Use the Scanner field to configure stream scanning.
*/
type Source struct {
	Scanner *bufio.Scanner
}

//Sink implements Sink interface for writing events to a stream
type Sink struct {
	Writer    *bufio.Writer
	separator []byte
}

//NewSource creates a new stream Source that reads events from the supplied io.Reader.
func NewSource(r io.Reader) *Source {
	return &Source{
		Scanner: bufio.NewScanner(r),
	}
}

//NewSink  create a new stream Sink that writes events to a stream
func NewSink(w io.Writer) *Sink {
	return &Sink{
		Writer:    bufio.NewWriter(w),
		separator: []byte("\n"),
	}
}

//Close closes this Source. Does not close the underlying stream.
func (source *Source) Close() error {
	//scanners don't need to be closed
	return nil
}

//DrawOne reads one event from the stream.
func (source *Source) DrawOne() (*types.Event, error) {
	if ok := source.Scanner.Scan(); ok {
		evt := types.NewEventFromBytes(source.Scanner.Bytes())
		return &evt, nil
	}
	err := source.Scanner.Err()
	if err == nil {
		err = errors.ErrStreamEnd
	}
	return nil, err
}

//Drain sends events to the stream
func (sink *Sink) Drain(events []types.Event) []error {
	for _, e := range events {
		bytes := e.Bytes()
		if _, err := sink.Writer.Write(bytes); err != nil {
			return []error{err}
		}
		sink.Writer.Write(sink.separator)
	}
	sink.Writer.Flush()
	return nil
}

//Close closes the stream
func (sink *Sink) Close() error {
	sink.Writer.Flush()
	return nil
}
