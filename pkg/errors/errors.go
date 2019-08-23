package errors

import (
	"fmt"
)

//ErrSourceClosed returned if you try to put on a closed source
var ErrSourceClosed = fmt.Errorf("SourceClosed")

//ErrChannelBroken returned when a source using a channel cannot send to it
var ErrChannelBroken = fmt.Errorf("ChannelBroken")
