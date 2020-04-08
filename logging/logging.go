package logging

import (
	"crypto/rand"
	"fmt"

	"github.com/kataras/iris/v12/context"
	"github.com/ont/iris-related/requestid"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

// Returns logger with random request-id pregenerated
func Generate() *logrus.Entry {
	return logger.WithField("request_id", randToken())
}

// SEE: https://stackoverflow.com/a/25431798
func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Returns logger from context.
func Get(ctx context.Context) *logrus.Entry {
	return ctx.Values().Get("logger").(*logrus.Entry)
}

// Logs fatal error and stops program
func Fatalf(message string, args ...interface{}) {
	logger.Fatalf(message, args...)
}

func Infof(message string, args ...interface{}) {
	logger.Infof(message, args...)
}

func Errorf(message string, args ...interface{}) {
	logger.Errorf(message, args...)
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

		// The main purpose of this middleware: setup logger entry with preconfigured request-id output
		entry := logger.WithField("request_id", requestid.Get(ctx))
		ctx.Values().Set("logger", entry)

		ctx.Next() // all ok, call other middlewares
	}
}

func init() {
	logger = logrus.New()
}
