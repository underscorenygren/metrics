package middleware

import (
	"context"
)

//Transformer Performs transformation on the body
type Transformer interface {
	Transform(ctx context.Context, bytes []byte) ([]byte, error)
}
