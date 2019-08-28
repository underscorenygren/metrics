//Package json handles parsing of json using fastjson
// NB: fastjson works by side effect, so transformation on
// a given metric is not thread safe
package json

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
	"time"
)

//Event event in json format
type Event struct {
	//You can access V directly to read json data using fastjson
	V   *fastjson.Value
	Evt *types.Event
}

//TransformerFn mapping fn signature for json events
type TransformerFn func(*Event) *Event

//SetString sets a string value on event. Shorthand for accessing fastjson value
func (e *Event) SetString(key string, value string) *Event {
	e.V.Set(key, fastjson.MustParse(fmt.Sprintf(`"%s"`, value)))
	return e
}

//SetInt sets an int value on event. Shorthand for accessing fastjson value
func (e *Event) SetInt(key string, value int) *Event {
	e.V.Set(key, fastjson.MustParse(fmt.Sprintf(`%d`, value)))
	return e
}

//Mapper returns a mapping function that processes and transforms
//a json event
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

		e := fn(&Event{V: v, Evt: evt})
		bytes = e.V.MarshalTo(nil)
		logger.Debug("json.Mapper: post-transform", zap.ByteString("eventBytes", bytes))

		return evt.NewBytes(bytes), nil
	}
}

//AddElasticsearchTimestamp convenience function for adding timestamp to an event
//when shipping to elasticsearch.
//Should probably live in a different module, but this is will be good enough for now.
func AddElasticsearchTimestamp(evt *Event) *Event {
	return evt.SetString("@timestamp", time.Now().UTC().Format(time.RFC3339Nano))
}
