package kinesis_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/failsink"
	"github.com/underscorenygren/partaj/pkg/kinesis"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/programmatic"
	"github.com/underscorenygren/partaj/pkg/types"
	"net/http"
)

func localStackRunning() bool {
	resp, err := http.Get(kinesis.LocalEndpoint)
	return err == nil && resp.StatusCode == http.StatusMethodNotAllowed
}

var _ = Describe("Kinesis", func() {

	logger := logging.ConfigureDevelopment(GinkgoWriter)
	streamName := "test"
	var sink *kinesis.Firehose

	BeforeEach(func() {
		if !localStackRunning() {
			logger.Debug("localstack not running")
			Skip("localstack isn't running")
		} else {
			sink, _ = kinesis.Sink(kinesis.SinkConfig{
				Name:  streamName,
				Local: true,
			})

			cli := sink.Client()
			cli.CreateDeliveryStream(&firehose.CreateDeliveryStreamInput{
				DeliveryStreamName: aws.String(streamName),
			})
		}
	})

	AfterEach(func() {
		if localStackRunning() {
			sink.Client().DeleteDeliveryStream(
				&firehose.DeleteDeliveryStreamInput{
					DeliveryStreamName: aws.String(streamName),
				})
		}
	})

	It("pushes events to kinesis", func() {

		//setup kinesis pipeline
		src := programmatic.NewSource()
		buf := buffer.NewSink()
		withFailures, err := failsink.NewSink(sink, buf)
		Expect(err).To(BeNil())

		p, err := pipe.Stage(src, withFailures)
		Expect(err).To(BeNil())

		//Run the pipeline
		src.PutString("a")
		src.PutString("a")
		src.Close()
		p.Flow()

		//Nothing should have failed
		Expect(buf.Events).To(Equal([]types.Event{}))
	})
})
