package transformer

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

type transformer struct {
	source   types.Source
	mapperFn MapperFn
}

//MapperFn fn signature for transforming one event
type MapperFn func(*types.Event) (*types.Event, error)

//Source makes a new transformer, that applies a transformation
//function on events drawn from the source
func Source(source types.Source, mapperFn MapperFn) (types.Source, error) {
	if source == nil {
		return nil, fmt.Errorf("source cannot be nil")
	}
	if mapperFn == nil {
		return nil, fmt.Errorf("transformer fn cannot be nil")
	}

	return &transformer{
		source:   source,
		mapperFn: mapperFn,
	}, nil
}

//DrawOne draws one event from the underlying source and transforms it
func (t *transformer) DrawOne() (*types.Event, error) {

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

//Close closing transformer closes it's containing source
func (t *transformer) Close() error {
	return t.source.Close()
}
