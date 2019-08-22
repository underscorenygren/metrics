package json

import (
	"github.com/underscorenygren/metrics/pkg/types"
	"github.com/valyala/fastjson"
)

type TransformerFn func(JSONEvent) *JSONEvent

type Source struct {
	parser       fastjson.Parser
	transformers []TransformerFn
}

type JSONEvent struct {
	V   *fastjson.Value
	Evt *types.Event
}

func NewSource(transformers []TransformerFn) *Source {
	return &Source{
		parser:       fastjson.Parser{},
		transformers: transformers,
	}
}
