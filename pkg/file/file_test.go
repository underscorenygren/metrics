package file_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bufio"
	"github.com/underscorenygren/partaj/internal"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/file"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/types"
	"io/ioutil"
	"os"
	"strings"
)

//Used to display test code in godoc
func Example() {}

var _ = Describe("File", func() {

	logging.ConfigureDevelopment(GinkgoWriter)

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

		buf := buffer.NewSink()
		source, err := file.NewSource(tmp.Name())
		Expect(err).To(BeNil())

		//Runs until file end
		p, err := pipe.NewStage(source, buf)
		Expect(err).To(BeNil())
		p.Flow()
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

		buf := buffer.NewSink()
		source, err := file.NewSource(tmp.Name())
		Expect(err).To(BeNil())

		//Runs until file end
		p, err := pipe.NewStage(source, buf)
		Expect(err).To(BeNil())
		p.Flow()
		Expect(buf.Events).To(Equal(refEvents))
		close(done)
	})

	It("reads an empty file correctly", func(done Done) {

		tmp, err := ioutil.TempFile("", "empty")
		Expect(err).To(BeNil())
		defer os.Remove(tmp.Name())

		refEvents := []types.Event{}

		buf := buffer.NewSink()
		source, err := file.NewSource(tmp.Name())
		Expect(err).To(BeNil())

		//Runs until file end
		p, err := pipe.NewStage(source, buf)
		Expect(err).To(BeNil())
		p.Flow()

		Expect(buf.Events).To(Equal(refEvents))
		close(done)
	})

	It("fulfills Source interface", func() {
		var source types.Source
		source = &file.Source{}
		Expect(source).ToNot(BeNil())
	})
})
