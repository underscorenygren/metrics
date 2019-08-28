package stream_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/stream"
	"github.com/underscorenygren/partaj/pkg/types"
	"os"
)

var _ = Describe("Stream", func() {

	It("reads an event from test stream, and closes correctly", func() {

		str := "test event"
		buf := bytes.NewBufferString(str)
		source := stream.NewSource(buf)

		//reads the event successfully
		read, err := source.DrawOne()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(read).ToNot(BeNil())
		Expect(read.Bytes()).To(Equal([]byte(str)))

		//stream end results in error
		read, err = source.DrawOne()
		Expect(err).To(Equal(errors.ErrStreamEnd))

		//no error on close
		Expect(source.Close()).ShouldNot(HaveOccurred())
	})

	It("fulfills source interface", func() {
		var source types.Source
		source = stream.NewSource(os.Stdin)
		Expect(source).ToNot(BeNil())
	})
})
