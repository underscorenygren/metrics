package firehose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	awsfirehose "github.com/aws/aws-sdk-go/service/firehose"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/failsink"
	"github.com/underscorenygren/partaj/pkg/firehose"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/programmatic"
	"github.com/underscorenygren/partaj/pkg/types"
	"net/http"
)

//examples on how to use firehose
func Example() {}

func localStackRunning() bool {
	resp, err := http.Get(firehose.LocalEndpoint)
	return err == nil && resp.StatusCode == http.StatusMethodNotAllowed
}

var _ = Describe("Firehose", func() {

	logger := logging.ConfigureDevelopment(GinkgoWriter)
	streamName := "test"
	var sink *firehose.Sink

	BeforeEach(func() {
		if !localStackRunning() {
			logger.Debug("localstack not running")
			Skip("localstack isn't running")
		} else {
			sink, _ = firehose.NewSink(firehose.Config{
				Name:  streamName,
				Local: true,
			})

			cli := sink.Client()
			cli.CreateDeliveryStream(&awsfirehose.CreateDeliveryStreamInput{
				DeliveryStreamName: aws.String(streamName),
			})
		}
	})

	AfterEach(func() {
		if localStackRunning() {
			sink.Client().DeleteDeliveryStream(
				&awsfirehose.DeleteDeliveryStreamInput{
					DeliveryStreamName: aws.String(streamName),
				})
		}
	})

	It("pushes events to firehose", func() {

		//setup pipeline
		src := programmatic.NewSource()
		buf := buffer.NewSink()
		withFailures, err := failsink.NewSink(sink, buf)
		Expect(err).To(BeNil())

		p, err := pipe.NewStage(src, withFailures)
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
