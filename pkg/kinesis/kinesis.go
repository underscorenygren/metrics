package kinesis

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
	"os"
)

const (
	//MaxSize max records for a kinesis put
	MaxSize = 500
	//LocalEndpoint address of firehose in localstack
	LocalEndpoint = "http://localhost:4573"
)

//ErrPutFailure error when all records fail to put, such as for IAM errors
var ErrPutFailure = fmt.Errorf("ErrPutFailure")

//Firehose Sink for pushing to firehose
type Firehose struct {
	Name     string
	firehose *firehose.Firehose
}

//SinkConfig for constructing Fireshose Sink
type SinkConfig struct {
	//Name name of firehose on aws
	Name string
	//Local points to local testing endpoint using localstack
	Local bool
}

//Sink constructs a firehose sink
func Sink(cfg SinkConfig) (*Firehose, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("No name provided")
	}
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = endpoints.UsEast1RegionID
	}
	awsCfg := aws.NewConfig().WithRegion(region)
	if cfg.Local {
		awsCfg.Endpoint = aws.String(LocalEndpoint)
	}
	firehose := firehose.New(session.New(), awsCfg)

	return &Firehose{
		Name:     cfg.Name,
		firehose: firehose,
	}, nil
}

//Client access the underlying aws Firehose client
func (fh *Firehose) Client() *firehose.Firehose {
	return fh.firehose
}

//Drain sends events to kinesis firehose
func (fh *Firehose) Drain(events []types.Event) []error {
	firehoseRecords := []*firehose.Record{}
	errs := []error{}
	logger := logging.Logger()

	//convert to firehose records
	for _, evt := range events {
		firehoseRecords = append(firehoseRecords, toRecord(evt.Bytes()))
	}

	//put batch to firehose
	logger.Debug("kinesis.Drain: putting batch",
		zap.Int("n", len(firehoseRecords)),
		zap.String("name", fh.Name),
	)
	res, err := fh.firehose.PutRecordBatch(&firehose.PutRecordBatchInput{
		DeliveryStreamName: aws.String(fh.Name),
		Records:            firehoseRecords,
	})
	logger.Debug("kinesis.Drain: finished batch")
	//Put error means all failed (permission error or whatnot)
	if err != nil {
		logger.Debug("kinesis.Drain: put error", zap.Error(err))
		for range firehoseRecords {
			errs = append(errs, ErrPutFailure)
		}
		return errs
	}

	//handle any failures
	nFailed := res.FailedPutCount
	failure := nFailed != nil && *nFailed > 0
	if failure {
		logger.Debug("kinesis.Drain: failed some records", zap.Int64("n", *nFailed))
		for index, resp := range res.RequestResponses {
			code := *resp.ErrorCode
			msg := *resp.ErrorMessage
			if resp.ErrorCode != nil {
				logger.Debug("kinesis.Drain: record error",
					zap.Int("index", index),
					zap.String("code", code),
					zap.String("msg", msg),
					zap.String("recordId", *resp.RecordId))
				errs = append(errs, fmt.Errorf("[%s]:%s", code, msg))
			} else {
				logger.Debug("kinesis.Drain: record succeeded",
					zap.Int("index", index))
				errs = append(errs, nil)
			}
		}

		return errs
	}

	return nil
}

//toRecord creates a kinesis firehose record from bytes
func toRecord(bytes []byte) *firehose.Record {
	return &firehose.Record{Data: bytes}
}
