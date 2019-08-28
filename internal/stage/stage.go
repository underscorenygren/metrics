//Package stage helper functions stages have in common
package stage

import (
	"go.uber.org/zap"
)

//FlattenErrors flattens error array to just one error for use in Flow
//logs errors, etc
//returns nil if there's no error
//logs by side-effect
func FlattenErrors(errs []error, logger *zap.Logger) error {
	var err error

	if errs == nil {
		return nil
	}

	for _, _err := range errs {
		if _err != nil {
			logger.Debug("stage.DrainFaulures: error", zap.Error(err))
			err = _err
		}
	}

	return err
}
