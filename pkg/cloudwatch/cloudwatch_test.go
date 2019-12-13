package cloudwatch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/cloudwatch"
	"github.com/underscorenygren/partaj/pkg/types"
	"net/http"
)

//examples on how to use cloudwatch
func Example() {}

func localStackRunning() bool {
	resp, err := http.Get(cloudwatch.LocalEndpoint)
	return err == nil && resp.StatusCode == http.StatusNotFound
}

var _ = Describe("Cloudwatch", func() {

	logger := logging.ConfigureDevelopment(GinkgoWriter)

	logGroupName := "test-log-group"
	logStreamName := "test-stream"
	nEvents := int64(3)
	//We would like to set limit to n-1 to test refresh, but it breaks localstack so we'll have to wait on it
	//limit := nEvents - 1

	var sink *cloudwatch.Sink
	var source *cloudwatch.Source

	BeforeEach(func() {
		if !localStackRunning() {
			logger.Debug("localstack not running")
			Skip("localstack isn't running")
		} else {
			var err error

			logger.Debug("creating sink")
			sink, err = cloudwatch.NewSink(cloudwatch.SinkConfig{
				LogGroupName:  logGroupName,
				LogStreamName: logStreamName,
				Local:         true,
			})

			Expect(err).To(BeNil())

			logger.Debug("creating source")
			source, err = cloudwatch.NewSource(cloudwatch.SourceConfig{
				LogGroupName:  logGroupName,
				LogStreamName: logStreamName,
				Limit:         nil,
				Local:         true,
			})

			client := source.Client()

			logger.Debug("creating log group")
			_, err = client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
				LogGroupName: aws.String(logGroupName),
			})
			Expect(err).To(BeNil())

			logger.Debug("creating log stream")
			_, err = client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
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

			_, err = client.DeleteLogStream(&cloudwatchlogs.DeleteLogStreamInput{
				LogGroupName:  aws.String(logGroupName),
				LogStreamName: aws.String(logStreamName),
			})
			Expect(err).To(BeNil())

			_, err = client.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
				LogGroupName: aws.String(logGroupName),
			})
			Expect(err).To(BeNil())
		}
	})

	It("reads and writes to cloudwatch", func() {
		one := types.NewEventFromBytes([]byte("one"))
		two := types.NewEventFromBytes([]byte("two"))
		three := types.NewEventFromBytes([]byte("three"))

		logger.Debug("testing local cloudwatch")
		//setup pipeline

		events := []types.Event{
			one,
			two,
			three,
		}
		//not strictly necessary, but checked to make sure adding events don't break test assumptions
		Expect(len(events)).To(Equal(int(nEvents)))

		errs := sink.Drain(events)

		Expect(errs).To(BeNil())

		evt, err := source.DrawOne()
		Expect(err).To(BeNil())
		Expect(evt).ToNot(BeNil())

		Expect(evt.IsEqual(&one)).To(BeTrue())

		evt, err = source.DrawOne()
		Expect(err).To(BeNil())
		Expect(evt).ToNot(BeNil())
		Expect(evt.IsEqual(&two)).To(BeTrue())

		evt, err = source.DrawOne()
		Expect(err).To(BeNil())
		Expect(evt).ToNot(BeNil())
		Expect(evt.IsEqual(&three)).To(BeTrue())

		//it will restart without "limit", which is currently broken in localstack
		//evt, err = source.DrawOne()
		//Expect(err).ToNot(BeNil())
	})
})
