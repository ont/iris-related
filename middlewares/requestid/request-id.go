package requestid

import (
	"encoding/hex"
	"github.com/kataras/iris/context"
	"math/rand"
)

// Returns request-id from context.
func Get(ctx context.Context) string {
	return ctx.Values().Get("request-id").(string)
}

// Use this function as middleware for iris.
// For example: app.Use(requestid.Middleware)
func Middleware(ctx context.Context) {
	requestId := getRequestId(ctx)
	ctx.Values().Set("request-id", requestId)
	ctx.Next() // all ok, call other middlewares
}

func getRequestId(ctx context.Context) string {
	value := ctx.GetHeader("X-Request-Id")

	if value == "" {
		value = genRand()
	}

	return value
}

// Generates random md5-like string.
func genRand() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
