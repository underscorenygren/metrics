package errors

import (
	"fmt"
)

//ErrSourceClosed returned if you try to put on a closed source
var ErrSourceClosed = fmt.Errorf("SourceClosed")

//ErrChannelBroken returned when a source using a channel cannot send to it
var ErrChannelBroken = fmt.Errorf("ChannelBroken")

//ErrNoRoute returned when a source using a channel cannot send to it
var ErrNoRoute = fmt.Errorf("NoRoute")

//ErrDrainFailures returned when drain fails in a stage
var ErrDrainFailures = fmt.Errorf("DrainFailures")

//ErrStreamEnd returned when a stream source is closed
var ErrStreamEnd = fmt.Errorf("StreamEnd")
