package math_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"encoding/binary"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/math"
	"github.com/underscorenygren/partaj/pkg/programmatic"
	"github.com/underscorenygren/partaj/pkg/types"
)

//examples on how to use json
func Example() {}

var _ = Describe("Math", func() {

	var testBytes [][]byte
	var sink = buffer.NewSink()
	var source = programmatic.NewSource()
	var stage types.Stage
	var err error

	logger := logging.ConfigureDevelopment(GinkgoWriter)

	nEvents := int64(4)

	BeforeEach(func() {
		testBytes = [][]byte{}
		source = programmatic.NewSource()
		for i := int64(0); i < nEvents; i++ {
			bytes := []byte{0}
			binary.PutVarint(bytes, i)
			testBytes = append(testBytes, bytes)
			source.PutBytes(bytes)
		}
		sink = buffer.NewSink()
		source.Close()
	})

	It("Emits a single event at end of stream", func(done Done) {
		logger.Debug("single stat")

		add := func(e *types.Event) float64 {
			i, err := binary.ReadVarint(bytes.NewReader(e.Bytes()))
			Expect(err).To(BeNil())
			return float64(i)
		}

		stage, err = math.NewStage(source, sink, 0, add)
		Expect(err).To(BeNil())

		err = stage.Flow()
		Expect(err).To(Equal(errors.ErrSourceClosed))

		Expect(len(sink.Events)).To(Equal(1))
		stat := sink.Events[0]

		state, err := math.Unmarshal(&stat)
		Expect(err).To(BeNil())
		Expect(state).ToNot(BeNil())

		Expect(state.Min).To(Equal(float64(0)))
		Expect(state.Max).To(Equal(float64(nEvents - 1)))
		Expect(state.Sum).To(Equal(float64(0 + 1 + 2 + 3)))
		Expect(state.N).To(Equal(nEvents))
		Expect(state.Average()).To(Equal(float64(0+1+2+3) / float64(nEvents)))

		close(done)
	})

	It("Emits events at interval", func(done Done) {
		logger.Debug("single stat")

		add := func(e *types.Event) float64 {
			i, err := binary.ReadVarint(bytes.NewReader(e.Bytes()))
			Expect(err).To(BeNil())
			return float64(i)
		}

		stage, err = math.NewStage(source, sink, nEvents/2, add)
		Expect(err).To(BeNil())

		err = stage.Flow()
		Expect(err).To(Equal(errors.ErrSourceClosed))

		Expect(len(sink.Events)).To(Equal(3))
		stat := sink.Events[0]

		state, err := math.Unmarshal(&stat)
		Expect(err).To(BeNil())
		Expect(state).ToNot(BeNil())

		Expect(state.Min).To(Equal(float64(0)))
		Expect(state.Max).To(Equal(float64(1)))
		Expect(state.Sum).To(Equal(float64(0 + 1)))
		Expect(state.N).To(Equal(int64(2)))

		stat = sink.Events[1]

		state, err = math.Unmarshal(&stat)
		Expect(err).To(BeNil())
		Expect(state).ToNot(BeNil())

		Expect(state.Min).To(Equal(float64(0)))
		Expect(state.Max).To(Equal(float64(nEvents - 1)))
		Expect(state.Sum).To(Equal(float64(0 + 1 + 2 + 3)))
		Expect(state.N).To(Equal(nEvents))

		stat = sink.Events[2]

		state, err = math.Unmarshal(&stat)
		Expect(err).To(BeNil())
		Expect(state).ToNot(BeNil())

		Expect(state.Min).To(Equal(float64(0)))
		Expect(state.Max).To(Equal(float64(nEvents - 1)))
		Expect(state.Sum).To(Equal(float64(0 + 1 + 2 + 3)))
		Expect(state.N).To(Equal(nEvents))

		close(done)
	})
})
