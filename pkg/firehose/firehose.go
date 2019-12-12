/*
Package firehose provides a sink that sends events to an AWS Firehose.
*/
package firehose

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/underscorenygren/partaj/internal/awsutil"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

const (
	//LocalEndpoint is the address of the firehose service when using localstack for testing.
	LocalEndpoint = "http://localhost:4573"
)

//Sink implements Sink interface for pushing events to a Firehose.
type Sink struct {
	Name     string //the name of the firehose
	firehose *firehose.Firehose
}

//Config is the input arguments to NewSink.
type Config struct {
	Name  string //name of the firehose as defined by AWS.
	Local bool   //when set to true, will configure firehose client to make requests to the local endpoint.
}

//NewSink constructs a firehose Sink.
func NewSink(cfg Config) (*Sink, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("No name provided")
	}
	endpoint := ""
	if cfg.Local {
		endpoint = LocalEndpoint
	}
	awsCfg := awsutil.GetDefaultConfig(endpoint)
	firehose := firehose.New(session.New(), awsCfg)

	return &Sink{
		Name:     cfg.Name,
		firehose: firehose,
	}, nil
}

//Client returns the underlying AWS Firehose Client
func (fh *Sink) Client() *firehose.Firehose {
	return fh.firehose
}

//Drain sends the supplied events to the firehose using
//PutRecordBatch.
func (fh *Sink) Drain(events []types.Event) []error {
	firehoseRecords := []*firehose.Record{}
	errs := []error{}
	logger := logging.Logger()

	//convert to firehose records
	for _, evt := range events {
		firehoseRecords = append(firehoseRecords, toRecord(evt.Bytes()))
	}

	//put batch to firehose
	logger.Debug("firehose.Drain: putting batch",
		zap.Int("n", len(firehoseRecords)),
		zap.String("name", fh.Name),
	)
	res, err := fh.firehose.PutRecordBatch(&firehose.PutRecordBatchInput{
		DeliveryStreamName: aws.String(fh.Name),
		Records:            firehoseRecords,
	})
	logger.Debug("firehose.Drain: finished batch")

	//Put error means all failed (permission error or whatnot)
	if err != nil {
		logger.Debug("firehose.Drain: put error", zap.Error(err))
		for range firehoseRecords {
			errs = append(errs, errors.ErrPutFailure)
		}
		return errs
	}

	//handle any failures
	nFailed := res.FailedPutCount
	failure := nFailed != nil && *nFailed > 0
	if failure {
		logger.Debug("firehose.Drain: failed some records", zap.Int64("n", *nFailed))
		for index, resp := range res.RequestResponses {
			code := *resp.ErrorCode
			msg := *resp.ErrorMessage
			if resp.ErrorCode != nil {
				logger.Debug("firehose.Drain: record error",
					zap.Int("index", index),
					zap.String("code", code),
					zap.String("msg", msg),
					zap.String("recordId", *resp.RecordId))
				errs = append(errs, fmt.Errorf("[%s]:%s", code, msg))
			} else {
				logger.Debug("firehose.Drain: record succeeded",
					zap.Int("index", index))
				errs = append(errs, nil)
			}
		}

		return errs
	}

	return nil
}

//toRecord creates a firehose firehose record from bytes
func toRecord(bytes []byte) *firehose.Record {
	return &firehose.Record{Data: bytes}
}
