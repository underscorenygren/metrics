package cloudwatch

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/underscorenygren/partaj/internal/awsutil"
	"github.com/underscorenygren/partaj/internal/timeutil"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/types"
)

const (
	//LocalEndpoint is the address of the cloudwatch service when using localstack for testing.
	LocalEndpoint = "http://localhost:4582"
)

//Source implements Source interface for cloudwatch logs
type Source struct {
	GetLogEventsInput cloudwatchlogs.GetLogEventsInput //Configures LogEvents input/filter
	cloudwatchlogs    *cloudwatchlogs.CloudWatchLogs   //client for interacting with API
	nextToken         *string
	initial           bool
	bufferedEvents    []*cloudwatchlogs.OutputLogEvent
}

//Sink implements Sink interface for cloudwatch logs
type Sink struct {
	LogGroupName   string
	LogStreamName  string
	cloudwatchlogs *cloudwatchlogs.CloudWatchLogs //client for interacting with API
}

//SourceConfig the input arguments for a new Source
type SourceConfig struct {
	LogGroupName  string
	LogStreamName string
	Local         bool //when set to true, will configure client to make requests to the local endpoint.
}

//SinkConfig the input arguments for a new Sink
type SinkConfig struct {
	LogGroupName  string
	LogStreamName string
	Local         bool //when set to true, will configure client to make requests to the local endpoint.
}

// ** Constructors ** //

//NewSource constructs a cloudwatch Source
func NewSource(cfg SourceConfig) (*Source, error) {
	if cfg.LogGroupName == "" {
		return nil, fmt.Errorf("No log group name provided")
	}
	if cfg.LogGroupName == "" {
		return nil, fmt.Errorf("No log stream name provided")
	}

	endpoint := ""
	if cfg.Local {
		endpoint = LocalEndpoint
	}
	awsCfg := awsutil.GetDefaultConfig(endpoint)

	client := cloudwatchlogs.New(session.New(), awsCfg)

	return &Source{
		cloudwatchlogs: client,
		initial:        true,
		GetLogEventsInput: cloudwatchlogs.GetLogEventsInput{
			EndTime:       nil,
			Limit:         nil,
			LogGroupName:  aws.String(cfg.LogGroupName),
			LogStreamName: aws.String(cfg.LogStreamName),
			StartFromHead: aws.Bool(true),
			StartTime:     nil,
		},
	}, nil
}

//NewSink creates a new sink for sending events to cloudwatch
func NewSink(cfg SinkConfig) (*Sink, error) {
	if cfg.LogGroupName == "" {
		return nil, fmt.Errorf("log group name cannot be empty")
	}
	if cfg.LogStreamName == "" {
		return nil, fmt.Errorf("log stream name cannot be empty")
	}
	endpoint := ""
	if cfg.Local {
		endpoint = LocalEndpoint
	}
	awsCfg := awsutil.GetDefaultConfig(endpoint)
	cloudwatchlogs := cloudwatchlogs.New(session.New(), awsCfg)

	return &Sink{
		LogGroupName:   cfg.LogGroupName,
		LogStreamName:  cfg.LogStreamName,
		cloudwatchlogs: cloudwatchlogs,
	}, nil
}

// ** SOURCE ** //

//calls internal cloudwatch logs api to buffer log events
func (source *Source) fetchFromClient() error {
	//super obnoxious there's not a "clone" on this command
	input := cloudwatchlogs.GetLogEventsInput{}
	input.EndTime = source.GetLogEventsInput.EndTime
	input.Limit = source.GetLogEventsInput.Limit
	input.StartFromHead = source.GetLogEventsInput.StartFromHead
	input.StartTime = source.GetLogEventsInput.StartTime
	input.LogGroupName = source.GetLogEventsInput.LogGroupName
	input.LogStreamName = source.GetLogEventsInput.LogStreamName

	if source.initial {
		source.initial = false
	} else {
		input.NextToken = source.nextToken
	}

	output, err := source.cloudwatchlogs.GetLogEvents(&input)
	if err != nil {
		return err
	}
	source.nextToken = output.NextForwardToken
	source.bufferedEvents = output.Events

	return nil
}

//advances internal buffer once and returns array
func (source *Source) advance() *cloudwatchlogs.OutputLogEvent {
	evt := source.bufferedEvents[0]
	source.bufferedEvents = source.bufferedEvents[1:]
	return evt
}

func (source *Source) hasBufferedEvents() bool {
	return source.bufferedEvents != nil && len(source.bufferedEvents) > 0
}

//DrawOne draws one event from the source
func (source *Source) DrawOne() (*types.Event, error) {
	if source.hasBufferedEvents() {
		if err := source.fetchFromClient(); err != nil {
			return nil, err
		}
	}

	if !source.hasBufferedEvents() {
		return nil, errors.ErrCloudwatchEnd
	}

	awsEvt := source.advance()
	evt := types.NewEventFromBytes([]byte(*awsEvt.Message))
	return &evt, nil
}

//Close does nothing but implements interface
func (source *Source) Close() error {
	return nil
}

//Client returns underlying cloudwatchlogs client
func (source *Source) Client() *cloudwatchlogs.CloudWatchLogs {
	return source.cloudwatchlogs
}

//* SINK **/

//Drain sends events to sink
func (sink *Sink) Drain(events []types.Event) []error {
	inputLogEvents := []*cloudwatchlogs.InputLogEvent{}
	nEvents := 0
	for _, evt := range events {
		nEvents++
		inputLogEvents = append(inputLogEvents, &cloudwatchlogs.InputLogEvent{
			Message:   aws.String(evt.String()),
			Timestamp: aws.Int64(timeutil.UnixMillis()),
		})
	}
	input := cloudwatchlogs.PutLogEventsInput{
		LogEvents:     inputLogEvents,
		LogGroupName:  aws.String(sink.LogGroupName),
		LogStreamName: aws.String(sink.LogStreamName),
	}

	_, err := sink.cloudwatchlogs.PutLogEvents(&input)
	errs := []error{}
	if err != nil {
		for i := 0; i < nEvents; i++ {
			errs = append(errs, err)
		}
	} else {
		errs = nil
	}

	return errs
}
