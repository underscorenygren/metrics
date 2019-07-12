package kinesis

import (
	"github.com/aws-sdk-go/aws/session"
	"github.com/aws-sdk-go/service/firehose"
	"github.com/underscorenygren/producer"
	"go.uber.org/zap"
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
	session := aws.Session()
	firehose := firehose.Firehose(session)

	return &kinesisProducer{
		name:     name,
		logger:   logger,
		firehose: firehose,
	}
}

func (kp *kinesisProducer) PutRecords(records [][]byte) []byte {
	firehoseRecords := []*firehose.Record{}
	failed := [][]byte{}

	for bytes := range records {
		firehoseRecords = append(firehoseRecords, toRecord(bytes))
	}

	kp.logger.Debug("kinesis putting records", zap.Integer("len", len(toProc)))
	res := firehose.PutRecordBatch(&firehose.PutRecordBatchInput{
		DeliveryStreamName: kp.name,
		Records:            firehoseRecords,
	})
	kp.logger.Debug("finished putting records")

	nFailed = res.FailedPutCount
	if nFailed > 0 {
		kp.Logger.Info("kinesis failed records", zap.Integer("len", nFailed))
		for index, resp := range res.RequestResponses {
			if resp.ErrorCode != nil {
				kp.logger.Error("kinesis producer error",
					zap.String("code", *resp.Code),
					zap.String("err", *resp.ErrorMessage),
					zap.String("recordId", *resp.RecordId))
				failed = append(failed, records[index])
			}
		}
	}

	return failed
}

func toRecord(bytes []byte) *kinesis.Record {
	return kinesis.Record{Data: bytes}
}
