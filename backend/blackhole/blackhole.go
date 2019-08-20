package blackhole

import (
	"github.com/underscorenygren/metrics/pkg/types"
)

//Drain Blackhole drains without any processing
func Drain([]types.Event) []error {
	return nil
}
