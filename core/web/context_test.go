package web_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/web"
)

func TestCtxValues(t *testing.T) {
	// Initialize ctx
	ctx := context.Background()

	// Check whether ctx values are present
	values, err := web.GetCtxValues(ctx)
	assert.Errorf(t, err, "context values are not set")
	assert.Nil(t, values)

	// Extend ctx with values
	ctx = web.SetCtxValues(ctx, &web.Values{
		TraceID: "traceID",
	})

	// Check whether ctx values are present
	values, err = web.GetCtxValues(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "traceID", values.TraceID)
}
