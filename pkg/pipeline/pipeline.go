/*
Package pipeline provides functionality for linking many stages together,
and other convenience functions.
*/
package pipeline

import (
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

/*
ParalellFailFirst executes supplied stages in paralell goroutines.

Returns a read-only channel for any errors returned by failed Flows.
*/
func ParalellFailFirst(stages []types.Stage, logger *zap.Logger) <-chan error {
	errChan := make(chan error)

	if stages == nil {
		errChan <- fmt.Errorf("no stages supplied")
		return errChan
	}

	for i, s := range stages {
		if s == nil {
			errChan <- fmt.Errorf("nil stage supplied")
		} else {
			logger.Debug("starting flow for stage", zap.Int("i", i))
			go func(s types.Stage) {
				errChan <- s.Flow()
			}(s)
		}
	}

	return errChan
}

/*
AsyncFlow calls flow in a goroutine and retuns a channel
which will return the Flow's error, at which point the
goroutine will also end.
*/
func AsyncFlow(s types.Stage) <-chan error {
	drained := make(chan error)
	logger := logging.Logger()
	go func() {
		logger.Debug("stage.AsyncFlow: starting")
		drained <- s.Flow()
		logger.Debug("stage.AsyncFlow: ended")
	}()
	return drained
}
