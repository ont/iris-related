package logging

import (
	"github.com/kataras/iris/context"
	"github.com/ont/iris-related/requestid"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

// Returns logger from context.
func Get(ctx context.Context) *logrus.Entry {
	return ctx.Values().Get("logger").(*logrus.Entry)
}

func Fatalf(message string, args ...interface{}) {
	logger.Fatalf(message, args...)
}

// Prepares logger and generates middleware handler.
// Usage: app.Use(logging.Middleware(logrus.JSONFormatter{}))
func Middleware(formatter logrus.Formatter) context.Handler {
	installed := false
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

func init() {
	logger = logrus.New()
}
