package stream

import (
	"bufio"
	"github.com/underscorenygren/metrics/pkg/types"
	"io"
)

//Source reads events from a buffered stream
type Source struct {
	//Scanner the underlying scanner object.
	//Can be accessed to change stream reading semantics,
	//e.g. not read by lines
	Scanner *bufio.Scanner
}

//NewSource reads events from a file-like stream
func NewSource(r io.Reader) *Source {
	return &Source{
		Scanner: bufio.NewScanner(r),
	}
}

//Close closes the stream
func (source *Source) Close() error {
	//scanners don't need to be closed
	return nil
}

//DrawOne reads one event from the stream
func (source *Source) DrawOne() (*types.Event, error) {
	if ok := source.Scanner.Scan(); ok {
		evt := types.NewEventFromBytes(source.Scanner.Bytes())
		return &evt, nil
	}
	return nil, source.Scanner.Err()
}
