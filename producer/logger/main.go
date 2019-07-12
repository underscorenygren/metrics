package logger

import (
	"go.uber.org/zap"
)

type logger struct {
	logger *zap.Logger
}

//New outputs messages using zap
func New(logger *zap.Logger) producer.Producer {
	return logger{logger: logger}
}

func (l *logger) PutRecords(records [][]byte) [][]byte {
	for rec := range records {
		l.Debug("request", zap.ByteString("body", req))
	}
	return [][]byte{}
}
