package json_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/sink/buffer"
	"github.com/underscorenygren/metrics/pkg/source"
)

type testEvent struct {
	i int `json:"id"`
}

var _ = Describe("Json", func() {
	var testBytes [][]byte
	logging.ConfigureDevelopment(GinkgoWriter)

	BeforeEach(func() {
		testBytes = [][]byte{}
		for i := 0; i < 3; i++ {
			bytes, err := json.Marshal(testEvent{i: i})
			Expect(err).To(BeNil())
			testBytes = append(testBytes, bytes)
		}
	})

	It("parses a sequence of json events", func(done Done) {

		source := source.NewProgrammaticSource()
		sink := buffer.Sink()

		pipeline.NewPipeline(source, sink)

		close(done)
	})

})
