package file

import (
	"github.com/underscorenygren/metrics/pkg/stream"
	"github.com/underscorenygren/metrics/pkg/types"
	"os"
)

//Source reads events from an open file
type Source struct {
	stream *stream.Source
	f      *os.File
}

//NewSource Creates a file source by openeing file at path
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

//Close closes file source
func (source *Source) Close() error {
	err := source.f.Close()
	streamErr := source.stream.Close()
	if err != nil {
		return err
	}
	return streamErr
}

//DrawOne reads one event
func (source *Source) DrawOne() (*types.Event, error) {
	return source.stream.DrawOne()
}
