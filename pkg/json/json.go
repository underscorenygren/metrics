/*
Package json uses fastjson package to do allow for json transformation
on events.

fastjson stores mutable bytes internally. You can parallelize handling
of separate events, but processing of a given event is not thread safe.
*/
package json

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
	"time"
)

//Event is an event that has been parsed into fastjson.
type Event struct {
	V *fastjson.Value //fastjson.Value exposed so you can operate directly in it inside the mapper
	E *types.Event    //original event is available as well
}

//TransformerFn is the function signature for transforming json events.
type TransformerFn func(*Event) *Event

//SetString is a convenience function for adding a string value to the specified key.
func (e *Event) SetString(key string, value string) *Event {
	e.V.Set(key, fastjson.MustParse(fmt.Sprintf(`"%s"`, value)))
	return e
}

//SetInt is a convenience function for adding an int value to the specified key.
func (e *Event) SetInt(key string, value int) *Event {
	e.V.Set(key, fastjson.MustParse(fmt.Sprintf(`%d`, value)))
	return e
}

/*
AddElasticsearchTimestamp is a convenience function  that adds the current time
as an elasticsearch-compatible RFC3339 string at "@timestamp".
*/
func AddElasticsearchTimestamp(evt *Event) *Event {
	return evt.SetString("@timestamp", time.Now().UTC().Format(time.RFC3339Nano))
}

/*
Mapper takes a TransformerFn signature - which operates on json.Events -
and returns a function signature compatible with the transformer.MapperFn signature.
*/
func Mapper(fn TransformerFn) func(*types.Event) (*types.Event, error) {
	//The closure is a types.TransformerFn
	return func(evt *types.Event) (*types.Event, error) {
		logger := logging.Logger()
		bytes := evt.Bytes()
		logger.Debug("json.Mapper: pre-transform", zap.ByteString("eventBytes", bytes))

		v, err := fastjson.ParseBytes(bytes)
		if err != nil {
			logger.Debug("json.Mapper: ParseBytes error", zap.Error(err))
			return evt, err
		}

		e := fn(&Event{V: v, E: evt})
		bytes = e.V.MarshalTo(nil)
		logger.Debug("json.Mapper: post-transform", zap.ByteString("eventBytes", bytes))

		return evt.NewBytes(bytes), nil
	}
}
