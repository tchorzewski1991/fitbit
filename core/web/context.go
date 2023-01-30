package web

import (
	"context"
	"errors"
	"time"
)

type ctxKey int

const ctxValuesKey ctxKey = 1

// Values encapsulates the details of the request which will be shared across program
// boundaries within the context.
type Values struct {
	TraceID   string
	StartedAt time.Time
}

// GetCtxValues retrieves instance of context Values out of given context.
func GetCtxValues(ctx context.Context) (*Values, error) {
	v, ok := ctx.Value(ctxValuesKey).(*Values)
	if !ok {
		return nil, errors.New("context values are not set")
	}
	return v, nil
}

// SetCtxValues extends given context with new instance of context Values.
func SetCtxValues(ctx context.Context, values *Values) context.Context {
	return context.WithValue(ctx, ctxValuesKey, values)
}
