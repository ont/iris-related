package etag

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/kataras/iris/v12"
)

func Record(ctx iris.Context) {
	ctx.Record()
	ctx.Next() // all ok, call other middlewares
}

func Emit(ctx iris.Context) {
	body := ctx.Recorder().Body()

	hasher := sha1.New()
	if _, err := hasher.Write(body); err != nil {
		return
	}

	hex := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	value := fmt.Sprintf("%d-%s", len(body), hex)
	ctx.Header("ETag", value)
}
