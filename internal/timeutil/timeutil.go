package timeutil

import "time"

//UnixMillis gets milliseconds from Unix Epoch
func UnixMillis() int64 {
	return time.Now().UnixNano() / 1000000
}
