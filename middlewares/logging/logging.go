package logging

import (
	"github.com/kataras/iris/context"
	"github.com/ont/iris-related/middlewares/requestid"
	"github.com/sirupsen/logrus"
)

// Returns logger from context.
func Get(ctx context.Context) *logrus.Entry {
	return ctx.Values().Get("logger").(*logrus.Entry)
}

// Serve middleware and prepare logger for each request's context.
// Usage: app.Use(logging.Middleware)
func Middleware(ctx context.Context) {
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}

	entry := logger.WithField("request-id", requestid.Get(ctx))

	ctx.Values().Set("logger", entry)
	ctx.Next() // all ok, call other middlewares
}
