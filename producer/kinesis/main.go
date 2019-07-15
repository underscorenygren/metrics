package kinesis

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/underscorenygren/metrics/producer"
	"go.uber.org/zap"
	"os"
)

const (
	//MaxSize max records for a kinesis put
	MaxSize = 500
)

type kinesisProducer struct {
	logger   *zap.Logger
	name     string
	firehose *firehose.Firehose
}

//New constructs kinesis producer
func New(name string, logger *zap.Logger) producer.Producer {

	region := "us-east-1"
	envRegion := os.Getenv("AWS_DEFAULT_REGION")
	if envRegion != "" {
		region = envRegion
	}
	awsCfg := aws.NewConfig().WithRegion(region)
	firehose := firehose.New(session.New(), awsCfg)

	return &kinesisProducer{
		name:     name,
		logger:   logger,
		firehose: firehose,
	}
}

func (kp *kinesisProducer) PutRecords(records [][]byte) [][]byte {
	firehoseRecords := []*firehose.Record{}
	failed := [][]byte{}

	for _, bytes := range records {
		firehoseRecords = append(firehoseRecords, toRecord(bytes))
	}

	kp.logger.Debug("kinesis putting records", zap.Int("len", len(records)))
	res, err := kp.firehose.PutRecordBatch(&firehose.PutRecordBatchInput{
		DeliveryStreamName: &kp.name,
		Records:            firehoseRecords,
	})
	kp.logger.Debug("finished putting records")
	if err != nil {
		kp.logger.Error("kinesis exception", zap.Error(err))
		//All failed
		return records
	}

	nFailed := res.FailedPutCount
	if nFailed != nil && *nFailed > 0 {
		kp.logger.Info("kinesis failed records", zap.Int64("len", *nFailed))
		for index, resp := range res.RequestResponses {
			if resp.ErrorCode != nil {
				kp.logger.Error("kinesis producer error",
					zap.String("code", *resp.ErrorCode),
					zap.String("err", *resp.ErrorMessage),
					zap.String("recordId", *resp.RecordId))
				failed = append(failed, records[index])
			}
		}
	}

	return failed
}

func toRecord(bytes []byte) *firehose.Record {
	return &firehose.Record{Data: bytes}
}
