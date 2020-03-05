package filter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/filter"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/programmatic"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

var _ = Describe("Filter", func() {

	logger := logging.ConfigureDevelopment(GinkgoWriter)

	filtered := types.NewEventFromBytes([]byte("filtered"))
	unfiltered := types.NewEventFromBytes([]byte("unfiltered"))

	filterFn := func(e *types.Event) (*types.Event, error) {
		if filtered.IsEqual(e) {
			return nil, nil
		}
		logger.Debug("filter_test.filterFn: not equal", zap.ByteString("e", e.Bytes()), zap.ByteString("filtered", filtered.Bytes()))
		return e, nil
	}

	var src *programmatic.Source
	var sink *buffer.Sink
	var events = []types.Event{
		unfiltered, filtered,
		unfiltered, unfiltered,
		filtered, filtered, unfiltered}
	var ref = []types.Event{unfiltered, unfiltered, unfiltered, unfiltered}

	BeforeEach(func() {
		src = programmatic.NewSource()
		sink = buffer.NewSink()

		for _, e := range events {
			src.Put(e)
		}
		src.Close()
	})

	It("Filters events from a source", func(done Done) {
		filtered, err := filter.NewSource(src, filterFn)
		Expect(err).To(BeNil())

		p, err := pipe.NewStage(filtered, sink)
		Expect(err).To(BeNil())

		err = p.Flow()
		Expect(err).To(Equal(errors.ErrSourceClosed))
		Expect(sink.Events).To(Equal(ref))

		close(done)
	})

	It("Filters events to a sink", func(done Done) {
		filtered, err := filter.NewSink(sink, filterFn)
		Expect(err).To(BeNil())

		p, err := pipe.NewStage(src, filtered)
		Expect(err).To(BeNil())

		err = p.Flow()
		Expect(err).To(Equal(errors.ErrSourceClosed))
		Expect(sink.Events).To(Equal(ref))

		close(done)
	})
})
