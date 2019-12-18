/*
Package math provides functionality for counting and doing stats on events
*/
package math

import (
	"encoding/json"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/internal/stage"
	"github.com/underscorenygren/partaj/pkg/types"
	gomath "math"
)

//State statistics state for a given Stage
type State struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Sum float64 `json:"sum"`
	N   int64   `json:"n"`
}

//ValueFn is the function signature for turning an event into a number
type ValueFn func(*types.Event) float64

//Stage struct for Stage data, fulfills Stage interface
type Stage struct {
	Interval int64
	fn       ValueFn
	source   types.Source
	sink     types.Sink
	state    State
}

//Average returns average for state
func (stat *State) Average() float64 {
	return stat.Sum / float64(stat.N)
}

//NewStage creates a Stage that emits stats about events
func NewStage(source types.Source, sink types.Sink, interval int64, fn ValueFn) (*Stage, error) {

	s := &Stage{
		Interval: interval,
		fn:       fn,
		source:   source,
		sink:     sink,
		state:    State{},
	}

	return s, nil
}

//Unmarshal parses stat struct from event. Returns error if invalid struct
func Unmarshal(e *types.Event) (*State, error) {
	state := State{}
	if err := json.Unmarshal(e.Bytes(), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

//update updates internal state
func (s *Stage) update(e *types.Event) {
	val := s.fn(e)
	s.state.Min = gomath.Min(s.state.Min, val)
	s.state.Max = gomath.Max(s.state.Max, val)
	s.state.Sum += val
	s.state.N++
}

//True iff Stage should emit an event
func (s *Stage) isAtEmitInterval() bool {
	return s.Interval > 0 && s.state.N%s.Interval == 0
}

//emits an event to the sink
func (s *Stage) emit() error {
	logger := logging.Logger()

	bytes, err := json.Marshal(s.state)
	if err != nil {
		return err
	}

	e := types.NewEventFromBytes(bytes)
	events := []types.Event{e}
	errs := s.sink.Drain(events)
	//End flow iff errors from drain
	if err = stage.FlattenErrors(errs, logger); err != nil {
		return err
	}

	return nil
}

/*Flow fulfills Stage interface.
* draws events from source and emits events at interval,
* and one final one at source end
 */
func (s *Stage) Flow() error {
	logger := logging.Logger()

	for {
		logger.Debug("math.Flow: drawing")
		e, err := s.source.DrawOne()
		logger.Debug("math.Flow: drew")

		if e != nil {
			s.update(e)
		}

		//always emit stats when
		if err != nil || e == nil {
			logger.Debug("math.Flow: emitting at end")
			//if emit fails, returns that error first
			if e2 := s.emit(); e2 != nil {
				return e2
			}
			return err
		} else if s.isAtEmitInterval() {
			logger.Debug("math.Flow: emitting at interval")
			//exit if emit fails
			if err := s.emit(); err != nil {
				return err
			}
		}
	}
}
