package cloudwatch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/services/cloudwatchlogs"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/cloudwatch"
	"github.com/underscorenygren/partaj/pkg/failsink"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/types"
	"net/http"
)

//examples on how to use cloudwatch
func Example() {}

func localStackRunning() bool {
	resp, err := http.Get(cloudwatch.LocalEndpoint)
	return err == nil && resp.StatusCode == http.StatusMethodNotAllowed
}

var _ = Describe("Cloudwatch", func() {

	logger := logging.ConfigureDevelopment(GinkgoWriter)

	logGroupName := "test-log-group"
	logStreamName := "test-stream"

	var sink *cloudwatch.Sink
	var source *cloudwatch.Source

	BeforeEach(func() {
		if !localStackRunning() {
			logger.Debug("localstack not running")
			Skip("localstack isn't running")
		} else {
			var err error

			sink, err = cloudwatch.NewSink(cloudwatch.SinkConfig{
				LogGroupName:  logGroupName,
				LogStreamName: logStreamName,
				Local:         true,
			})

			Expect(err).To(BeNil())

			source, err = cloudwatch.NewSource(cloudwatch.SourceConfig{
				LogGroupName:  logGroupName,
				LogStreamName: logStreamName,
				Local:         true,
			})

			client := source.Client()

			_, err = client.CreateLogGroup(cloudwatchlogs.CreateLogGroupInput{
				LogGroupName: logGroupName,
			})
			Expect(err).To(BeNil())

			_, err = client.CreateLogStream(cloudwatchlogs.CreateLogStreamInput{
				LogGroupName:  aws.String(logGroupName),
				LogStreamName: aws.String(logStreamName),
			})
			Expect(err).To(BeNil())
		}
	})

	AfterEach(func() {
		if localStackRunning() {
			var err error
			client := source.Client()

			err = client.DeleteLogStream(cloudwatchlogs.DeleteLogStreamInput{
				LogGroupName:  aws.String(logGroupName),
				LogStreamName: aws.String(logStreamName),
			})
			Expect(err).To(BeNil())

			err = client.DeleteLogGroup(cloudwatchlogs.DeleteLogGroupInput{
				LogGroupName: aws.String(logGroupName),
			})
			Expect(err).To(BeNil())
		}
	})

	It("pushes pushes and reads events to local cloudwatch", func() {

		//setup pipeline

		buf := buffer.NewSink()
		withFailures, err := failsink.NewSink(sink, buf)
		Expect(err).To(BeNil())

		p, err := pipe.Stage(source, withFailures)
		Expect(err).To(BeNil())

		//Run the pipeline
		errs := source.Drain([]events{
			types.EventFromBytes([]byte("one")),
			types.EventFromBytes([]byte("two")),
		})
		p.Flow()

		//Nothing should have failed
		Expect(buf.Events).To(Equal([]types.Event{}))
	})
})
