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
