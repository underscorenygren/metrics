package file_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bufio"
	"github.com/underscorenygren/metrics/internal"
	"github.com/underscorenygren/metrics/pkg/buffer"
	"github.com/underscorenygren/metrics/pkg/file"
	"github.com/underscorenygren/metrics/pkg/pipeline"
	"github.com/underscorenygren/metrics/pkg/types"
	"io/ioutil"
	"os"
	"strings"
)

var _ = Describe("File", func() {

	It("reads a sequence of events from a file", func(done Done) {

		tmp, err := ioutil.TempFile("", "sequence")
		Expect(err).To(BeNil())
		defer os.Remove(tmp.Name())

		ref := []string{"one", "two", "three"}

		w := bufio.NewWriter(tmp)
		w.WriteString(strings.Join(ref, "\n"))
		w.Flush()
		tmp.Seek(0, 0)

		refEvents := internal.StringsToEvents(ref)

		buf := buffer.Sink()
		source, err := file.NewSource(tmp.Name())
		Expect(err).To(BeNil())

		//Runs until file end
		pipeline.NewPipeline(source, buf).Flow()
		Expect(buf.Events).To(Equal(refEvents))
		close(done)
	})

	It("reads a file with one element ending in newline", func(done Done) {

		tmp, err := ioutil.TempFile("", "one")
		Expect(err).To(BeNil())
		defer os.Remove(tmp.Name())

		w := bufio.NewWriter(tmp)
		w.WriteString("one\n")
		w.Flush()
		tmp.Seek(0, 0)

		refEvents := internal.StringsToEvents([]string{"one"})

		buf := buffer.Sink()
		source, err := file.NewSource(tmp.Name())
		Expect(err).To(BeNil())

		//Runs until file end
		pipeline.NewPipeline(source, buf).Flow()
		Expect(buf.Events).To(Equal(refEvents))
		close(done)
	})

	It("reads an empty file correctly", func(done Done) {

		tmp, err := ioutil.TempFile("", "empty")
		Expect(err).To(BeNil())
		defer os.Remove(tmp.Name())

		refEvents := []types.Event{}

		buf := buffer.Sink()
		source, err := file.NewSource(tmp.Name())
		Expect(err).To(BeNil())

		//Runs until file end
		pipeline.NewPipeline(source, buf).Flow()

		Expect(buf.Events).To(Equal(refEvents))
		close(done)
	})

	It("fulfills Source interface", func() {
		var source types.Source
		source = &file.Source{}
		Expect(source).ToNot(BeNil())
	})
})
