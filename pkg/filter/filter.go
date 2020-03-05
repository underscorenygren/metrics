/*
Package filter wraps sinks or source in a filter, that
can be used to filter out certain events
*/
package filter

import (
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//Source filters events as they are drawn
type Source struct {
	source types.Source
	fn     EventFilterFn
}

//Sink filters events as they are drained
type Sink struct {
	sink types.Sink
	fn   EventFilterFn
}

//EventFilterFn function signature for filtering function
type EventFilterFn func(*types.Event) (*types.Event, error)

//implements interfaces
var _ types.Sink = &Sink{}
var _ types.Source = &Source{}

//NewSource creates a filter on a source
func NewSource(source types.Source, fn EventFilterFn) (*Source, error) {
	if source == nil {
		return nil, errors.ErrNilSource
	}

	if fn == nil {
		return nil, errors.ErrNilFn
	}

	return &Source{
		source: source,
		fn:     fn,
	}, nil
}

//NewSink creates a filter on a sink
func NewSink(sink types.Sink, fn EventFilterFn) (*Sink, error) {
	if sink == nil {
		return nil, errors.ErrNilSink
	}

	if fn == nil {
		return nil, errors.ErrNilFn
	}

	return &Sink{
		sink: sink,
		fn:   fn,
	}, nil
}

//DrawOne draws from source and filters by EventFilterFn
func (source *Source) DrawOne() (*types.Event, error) {
	e, err := source.source.DrawOne()
	if err == nil && e != nil {
		var newE *types.Event
		newE, err = source.fn(e)
		if newE == nil {
			logging.Logger().Debug("filter.DrawOne: filtered", zap.ByteString("event", e.Bytes()))
		}
		e = newE
	}
	return e, err
}

//Close fulfills close on source interface
func (source *Source) Close() error {
	return source.source.Close()
}

//Drain filters events by EventFilterFn before draining
func (sink *Sink) Drain(events []types.Event) []error {
	newEvents := []types.Event{}
	errs := []error{}
	for _, e := range events {
		ne, err := sink.fn(&e)
		if err != nil {
			errs = append(errs, err)
		} else {
			if ne != nil {
				newEvents = append(newEvents, *ne)
			} else {
				logging.Logger().Debug("filter.Drain: filtered", zap.ByteString("event", e.Bytes()))
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}

	return sink.sink.Drain(newEvents)
}
