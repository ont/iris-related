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

// Prepares logger and generates middleware handler.
// Usage: app.Use(logging.Middleware(logrus.JSONFormatter{}))
func Middleware(formatter logrus.Formatter) context.Handler {
	installed := false

	logger := logrus.New()
	logger.Formatter = formatter

	return func(ctx context.Context) {
		// TODO: refactor "installed" check
		if !installed {
			ctx.Application().Logger().Install(logger)
			installed = true
		}

		entry := logger.WithField("request-id", requestid.Get(ctx))

		ctx.Values().Set("logger", entry)
		ctx.Next() // all ok, call other middlewares
	}
}
