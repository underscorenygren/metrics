package internal

import (
	"github.com/underscorenygren/metrics/pkg/types"
)

//ToEvents converts a slice of byte sequences to a slice of events
func ToEvents(bytesSlice [][]byte) []types.Event {
	evts := []types.Event{}
	for _, bytes := range bytesSlice {
		evts = append(evts, types.NewEventFromBytes(bytes))
	}
	return evts
}
