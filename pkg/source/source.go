package source

import (
	"errors"
)

//ErrSourceClosed returned if you try to put on a closed source
var ErrSourceClosed = errors.New("Source is closed")
