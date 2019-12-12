/*
Package errors defines all typed errors used by all other modules.
*/
package errors

import (
	"fmt"
)

//ErrSourceClosed is returned when you try to put to a closed programmatic Source.
var ErrSourceClosed = fmt.Errorf("SourceClosed")

//ErrChannelBroken is returned when a programmatic Source can't send to its internal channel.
var ErrChannelBroken = fmt.Errorf("ChannelBroken")

//ErrStreamEnd is returned when trying to Draw from a closed stream Source.
var ErrStreamEnd = fmt.Errorf("StreamEnd")

//ErrPutFailure is returned when firehose Sink fails to put all events, e.g. for IAM errors.
var ErrPutFailure = fmt.Errorf("ErrPutFailure")

//ErrCloudwatchEnd is returned when no more entries are available from cloudwatch source
var ErrCloudwatchEnd = fmt.Errorf("ErrCloudwatchEnd")
