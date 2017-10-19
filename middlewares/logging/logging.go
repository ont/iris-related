package logging

import (
	"github.com/kataras/iris/context"
	"github.com/sirupsen/logrus"
	"server/middlewares/requestid"
)

// Returns logger from context.
func Get(ctx context.Context) *logrus.Entry {
	return ctx.Values().Get("logger").(*logrus.Entry)
}

// Serve middleware and prepare logger for each request's context.
// Usage: app.Use(logging.Middleware)
func Middleware(ctx context.Context) {
	logger := logrus.WithField("request-id", requestid.Get(ctx))
	ctx.Values().Set("logger", logger)
	ctx.Next() // all ok, call other middlewares
}
