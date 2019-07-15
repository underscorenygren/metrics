package server

import (
	"github.com/underscorenygren/metrics/producer"
	"go.uber.org/zap"
	"time"
)

type state struct {
	logger *zap.Logger
	p      producer.Producer
}

//Config server configuration
type Config struct {
	Port              *int
	Host              *string
	ReadHeaderTimeout *time.Duration
	ReadTimeout       *time.Duration
	WriteTimeout      *time.Duration
	Producer          producer.Producer
	Logger            *zap.Logger
}
