package logger

import (
	"github.com/underscorenygren/metrics/producer"
	"go.uber.org/zap"
)

type logProducer struct {
	logger *zap.Logger
}

//New outputs messages using zap
func New(logger *zap.Logger) producer.Producer {
	return &logProducer{logger: logger}
}

func (l *logProducer) PutRecords(records [][]byte) [][]byte {
	for _, rec := range records {
		l.logger.Debug("request", zap.ByteString("body", rec))
	}
	return records
}
