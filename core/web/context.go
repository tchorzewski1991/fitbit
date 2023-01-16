package web

import (
	"context"
	"errors"
	"time"
)

type ctxKey int

const ctxValuesKey ctxKey = 1

type Values struct {
	TraceID   string
	StartedAt time.Time
}

func GetCtxValues(ctx context.Context) (*Values, error) {
	v, ok := ctx.Value(ctxValuesKey).(*Values)
	if !ok {
		return nil, errors.New("context values are not set")
	}
	return v, nil
}

func SetCtxValues(ctx context.Context, values *Values) context.Context {
	return context.WithValue(ctx, ctxValuesKey, values)
}
