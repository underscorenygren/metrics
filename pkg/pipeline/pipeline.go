//Package pipeline links many stages together
package pipeline

import (
	"fmt"
	"github.com/underscorenygren/metrics/internal/logging"
	"github.com/underscorenygren/metrics/pkg/types"
	"go.uber.org/zap"
)

//ParalellFailFirst executes stages in paralell, and returns a channel
//that will return any errors as soon as they supplied
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

/*AsyncFlow calls flow in a goroutine and retuns a channel
 * with the result of the flow when it ends
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
