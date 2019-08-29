/*
Package transformer is a source that applies a transformation function to events
passing through it.
*/
package transformer

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//Source fulfills the Source interface. Events read from
//this source are transformed with the supplied MapperFn.
type Source struct {
	source   types.Source
	mapperFn MapperFn
}

/*
MapperFn is the function signature for transforming a single event.

Errors returned by the mapper function will be returned in calls to Draw.
*/
type MapperFn func(*types.Event) (*types.Event, error)

//NewSource makes a new transformer Source, that applies a transformation
//function on events drawn from the supplied source.
func NewSource(source types.Source, mapperFn MapperFn) (*Source, error) {
	if source == nil {
		return nil, fmt.Errorf("source cannot be nil")
	}
	if mapperFn == nil {
		return nil, fmt.Errorf("transformer fn cannot be nil")
	}

	return &Source{
		source:   source,
		mapperFn: mapperFn,
	}, nil
}

//DrawOne draws one event from the underlying source and transforms it.
func (t *Source) DrawOne() (*types.Event, error) {

	logger := logging.Logger()
	e, err := t.source.DrawOne()
	if err != nil {
		logger.Debug("transformer.DrawOne: error on draw", zap.Error(err))
		return nil, err
	}

	logger.Debug("transformer.DrawOne: transforming", zap.ByteString("event", e.Bytes()))
	e, err = t.mapperFn(e)
	if err != nil {
		logger.Debug("transformer.DrawOne: mapping error", zap.Error(err))
		return nil, err
	}
	if e != nil {
		logger.Debug("transformer.DrawOne: transformed", zap.ByteString("event", e.Bytes()))
	} else {
		logger.Debug("transormer.DrawOne: nulled event")
	}

	return e, nil
}

//Close closes the transformer and it's underlying source.
func (t *Source) Close() error {
	return t.source.Close()
}
