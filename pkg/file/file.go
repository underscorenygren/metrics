/*
Package file reads events from a file.
*/
package file

import (
	"github.com/underscorenygren/partaj/pkg/stream"
	"github.com/underscorenygren/partaj/pkg/types"
	"os"
)

//Source fulfills the source interface. Reads events from a file.
type Source struct {
	stream *stream.Source
	f      *os.File
}

//Sink fulffils the sink interface. Writes events to a file
type Sink struct {
	stream *stream.Sink
	f      *os.File
}

/*
NewSource creates a file Source by opening the file
supplied at `path`.

File is read using a `stream` Source, and assumes newlines for separators.
In a future version, this will be configurable.

Will return underlying opening error if file open fails.
*/
func NewSource(path string) (*Source, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &Source{
		f:      f,
		stream: stream.NewSource(f),
	}, nil
}

/*
NewSink creates a sink Source by openening the file
supplied at `path`.

Will return underlying opening error if it fails
*/
func NewSink(path string) (*Sink, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &Sink{
		f:      f,
		stream: stream.NewSink(f),
	}, nil
}

//Close closes the source and the file handle with it.
func (source *Source) Close() error {
	err := source.f.Close()
	streamErr := source.stream.Close()
	if err != nil {
		return err
	}
	return streamErr
}

//DrawOne reads one event from the underlying file.
func (source *Source) DrawOne() (*types.Event, error) {
	return source.stream.DrawOne()
}

//Drain writes one event to the file.
func (sink *Sink) Drain(events []types.Event) []error {
	return sink.stream.Drain(events)
}

//Close closes the sink and the underlying file
func (sink *Sink) Close() error {
	sink.stream.Close()
	sink.f.Sync()
	return sink.f.Close()
}
