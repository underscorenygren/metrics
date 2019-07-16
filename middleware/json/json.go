package json

import (
	"context"
	"fmt"
	"github.com/valyala/fastjson"
	"time"
)

//JSON parses and writes json
type JSON struct {
	chain []Processor
}

//Processor function signature for processor function
type Processor func(v *fastjson.Value, ctx context.Context) error

//New makes a new json parsing middleware
func New(chain []Processor) *JSON {
	json := &JSON{}
	for _, p := range chain {
		json.chain = append(json.chain, p)
	}
	return json
}

//AddFromContext reads field key as string and puts it as key on object
func AddFromContext(key interface{}, field string) Processor {
	return func(v *fastjson.Value, ctx context.Context) error {
		if val, ok := ctx.Value(key).(string); ok {
			return setString(field, val, v)
		}
		return fmt.Errorf("%s field is not a string", field)
	}
}

//ElasticsearchTimestamp adds the current time as elasticsearch's formatted string
func ElasticsearchTimestamp() Processor {
	key := "@timestamp"
	transformer := func(v *fastjson.Value, ctx context.Context) error {
		return setString(key,
			time.Now().UTC().Format(time.RFC3339Nano),
			v)
	}
	return transformer
}

//Transform executes json transformation
func (json *JSON) Transform(ctx context.Context, bytes []byte) ([]byte, error) {
	parser := fastjson.Parser{}
	v, err := parser.ParseBytes(bytes)
	if err != nil {
		return nil, err
	}

	for _, transformer := range json.chain {
		err := transformer(v, ctx)
		if err != nil {
			return nil, err
		}
	}

	return v.MarshalTo(nil), nil
}

//write string to fastjson object
func setString(key, val string, dst *fastjson.Value) error {
	//fastjson uses strict json format here, so quotes needed:
	//https://godoc.org/github.com/valyala/fastjson#Value.Set
	produced, err := fastjson.Parse(fmt.Sprintf(`"%s"`, val))
	if err != nil {
		return err
	}
	dst.Set(key, produced)
	return nil
}
