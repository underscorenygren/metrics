package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/internal/timeutil"
	"github.com/underscorenygren/partaj/pkg/cloudwatch"
	"github.com/underscorenygren/partaj/pkg/errfilter"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/pipe"
	"github.com/underscorenygren/partaj/pkg/stream"
	"github.com/underscorenygren/partaj/pkg/transformer"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

var errMissingLogGroup = "ResourceNotFoundException"

//Used with compile flags
var version = "0.0.0"

//Stream one log stream
type Stream struct {
	name  string
	stage types.Stage
}

//Streamer struct for all streams
type Streamer struct {
	client       *cloudwatchlogs.CloudWatchLogs
	nextToken    *string
	logStreams   map[string]*Stream
	logger       *zap.Logger
	logGroupName string
}

//creates the top level streamer that holds many streamers to many log groups+streams
func newStreamer(logGroupName string) (*Streamer, error) {
	if logGroupName == "" {
		return nil, fmt.Errorf("logGroupName must be set")
	}

	streamer := &Streamer{
		logGroupName: logGroupName,
		client:       cloudwatch.NewClient(false),
		logStreams:   map[string]*Stream{},
		logger:       logging.Logger(),
	}

	streamer.logger.Debug("created Streamer", zap.String("logGroupName", logGroupName))

	return streamer, nil
}

//newStream creates a new source/sink for a log group name and stream
func newStream(logGroupName, logStreamName string) (*Stream, error) {
	cloudwatchSource, err := cloudwatch.NewSource(cloudwatch.SourceConfig{
		LogGroupName:  logGroupName,
		LogStreamName: logStreamName,
		StartTime:     aws.Int64(timeutil.UnixMillis()),
	})
	if err != nil {
		return nil, err
	}

	mapper := func(evt *types.Event) (*types.Event, error) {
		return evt.NewBytes([]byte(fmt.Sprintf("%s:%b", logStreamName, evt.Bytes()))), nil
	}

	transformer, err := transformer.NewSource(cloudwatchSource, mapper)
	if err != nil {
		return nil, err
	}
	//TODO Test
	if transformer != nil {
		transformer = nil
	}

	streamSink := stream.NewSink(os.Stdout)

	pipe, err := pipe.NewStage(cloudwatchSource, streamSink)
	if err != nil {
		return nil, err
	}

	stage, err := errfilter.NewStage(pipe, []error{errors.ErrCloudwatchEnd}, 1500*time.Millisecond)
	if err != nil {
		return nil, err
	}

	return &Stream{
		name:  logStreamName,
		stage: stage,
	}, nil
}

//runs one stream
func (s *Stream) run() {
	logger := logging.Logger()
	err := s.stage.Flow()
	if err != nil {
		logger.Error("Stream Errored",
			zap.String("name", s.name),
			zap.Error(err))
	} else {
		logger.Info("Stream Ended with nil",
			zap.String("name", s.name))
	}
}

//calls the cloudwatchlogs list command once
func (streamer *Streamer) list() ([]string, error) {
	res := []string{}

	streamer.logger.Debug("listing", zap.String("logGroupName", streamer.logGroupName))

	out, err := streamer.client.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(streamer.logGroupName),
		NextToken:    streamer.nextToken,
		OrderBy:      aws.String("LastEventTime"),
	})

	if err != nil {
		return nil, err
	}

	streamer.nextToken = out.NextToken
	for _, s := range out.LogStreams {
		res = append(res, *s.LogStreamName)
	}

	return res, nil
}

//isFatal true iff cloudwatch error isn't retryable
func isFatal(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == errMissingLogGroup
	}
	return false
}

//lists streams once
func (streamer *Streamer) doList() error {
	streamer.logger.Debug("listing streams")
	streams, err := streamer.list()
	if isFatal(err) {
		return err
	} else if err != nil {
		streamer.logger.Error("couldn't list streams", zap.Error(err))
	} else {
		for _, streamName := range streams {
			if _, ok := streamer.logStreams[streamName]; !ok {
				streamer.logger.Debug("creating stream", zap.String("streamName", streamName))
				s, err := newStream(streamer.logGroupName, streamName)
				if err != nil {
					streamer.logger.Error("couldn't create stream", zap.Error(err))
				} else {
					streamer.logStreams[streamName] = s
					go s.run()
				}
			} else {
				streamer.logger.Debug("stream pre-existing", zap.String("streamName", streamName))
			}
		}
	}

	return nil
}

//Run streams forever
func (streamer *Streamer) Run() error {
	//Checks for new streams every second
	err := streamer.doList()
	if err != nil {
		return err
	}
	for range time.Tick(3 * time.Second) {
		err = streamer.doList()
		if err != nil {
			return err
		}
	}

	return nil
}

//command entry point
func main() {
	var logGroupName string
	var debug bool
	var versionFlag bool

	flag.StringVar(&logGroupName, "name", "", "the  name of the log group to tail")
	flag.BoolVar(&debug, "debug", false, "sets logging to debug level")
	flag.BoolVar(&versionFlag, "version", false, "prints version and exits")

	flag.Parse()

	if versionFlag {
		fmt.Printf("%s\n", version)
		return
	}

	if debug {
		logging.SetLevel(zap.DebugLevel)
		logging.Logger().Debug("setting to debug level")
	}

	streamer, err := newStreamer(logGroupName)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(streamer.Run())
}
