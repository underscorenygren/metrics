package json_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"github.com/underscorenygren/metrics/internal/logging"
	pkgjson "github.com/underscorenygren/metrics/pkg/json"
	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/sink/buffer"
	"github.com/underscorenygren/metrics/pkg/source"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
	"strings"
)

type testEvent struct {
	ID int `json:"id"`
}

type testResult struct {
	ID    int    `json:"id"`
	Index int    `json:"index"`
	A     string `json:"a"`
}

var _ = Describe("Json", func() {
	var testBytes [][]byte
	var pipe *pipeline.Pipeline
	var sink *buffer.Buffer
	logger := logging.ConfigureDevelopment(GinkgoWriter)

	//We use Ginkgo's suggested method to put common setup
	//in BeforeEach and assign to vars
	BeforeEach(func() {
		testBytes = [][]byte{}
		for i := 0; i < 3; i++ {
			bytes, err := json.Marshal(testEvent{ID: i})
			Expect(err).To(BeNil())
			testBytes = append(testBytes, bytes)
		}
		//make pipeline for processing
		src := source.NewProgrammaticSource()
		sink = buffer.Sink()
		pipe = pipeline.NewPipeline(src, sink)

		//fill pipeline and close source
		for _, bytes := range testBytes {
			src.PutBytes(bytes)
		}
		src.Close()
	})

	It("parses and maps sequence of json events", func(done Done) {

		//some fake dynamic data to add to object in mapper
		i := 0
		enumeration := func() int {
			pre := i
			i++
			return pre
		}

		//mapper functoin adds some fields
		pipe.MapFn = pkgjson.Mapper(
			func(jsonEvent *pkgjson.Event) *pkgjson.Event {
				return jsonEvent.
					SetInt("index", enumeration()).
					SetString("a", strings.Repeat("a", i))
			})

		//execute pipeline
		err := pipe.Flow()
		Expect(err).To(Equal(source.ErrSourceClosed))

		//build expected result
		expected := []testResult{
			testResult{ID: 0, Index: 0, A: "a"},
			testResult{ID: 1, Index: 1, A: "aa"},
			testResult{ID: 2, Index: 2, A: "aaa"},
		}
		ref := []types.Event{}
		for _, res := range expected {
			dumped, err := json.Marshal(res)
			Expect(err).To(BeNil())
			ref = append(ref, types.NewEventFromBytes(dumped))
		}

		//ensure marshalling worked
		Expect(sink.Events).To(Equal(ref))

		close(done)
	})

	It("parses json data, reads contents", func(done Done) {

		ref := 3 // we just add all the IDs together
		total := 0

		pipe.MapFn = pkgjson.Mapper(
			func(jsonEvent *pkgjson.Event) *pkgjson.Event {
				//Uses underlying V to get value
				v := jsonEvent.V.GetInt("id")
				logger.Debug("id is", zap.Int("id", v))
				total += v
				return jsonEvent
			})

		//execute pipeline
		err := pipe.Flow()
		Expect(err).To(Equal(source.ErrSourceClosed))

		//ids should have been summed
		Expect(total).To(Equal(ref))

		close(done)
	})
})
