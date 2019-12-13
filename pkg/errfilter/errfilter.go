package errfilter

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
	"time"
)

/*
Keeps a stage flowing through predefined errors
*/

//ErrFilter implements Stage interface
type ErrFilter struct {
	stage         types.Stage
	ignoredErrors []error
	waitDuration  time.Duration
}

//NewStage creates a new error filtering stage
func NewStage(stage types.Stage, ignoredErrors []error, waitDuration time.Duration) (*ErrFilter, error) {
	if stage == nil {
		return nil, fmt.Errorf("no stage provided")
	}
	return &ErrFilter{
		stage:         stage,
		ignoredErrors: ignoredErrors,
		waitDuration:  waitDuration,
	}, nil
}

func contains(errs []error, err error) bool {
	for _, e := range errs {
		if e == err {
			return true
		}
	}
	return false
}

/*
Flow fulfills Stage interface. Calls flow on Flow() on underlying
stage while error return is not in ignoredErrors
*/
func (ef *ErrFilter) Flow() error {
	var err error
	for err == nil || contains(ef.ignoredErrors, err) {
		err = ef.stage.Flow()
		if err != nil {
			logging.Logger().Debug("sleeping", zap.Duration("duration", ef.waitDuration))
			time.Sleep(ef.waitDuration)
		}
	}
	return err
}
