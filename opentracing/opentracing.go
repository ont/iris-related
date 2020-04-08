package opentracing

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"runtime/debug"

	gocontext "context"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var (
	tracer opentracing.Tracer
	closer io.Closer

	reNum  *regexp.Regexp
	reUuid *regexp.Regexp
	reHash *regexp.Regexp
)

// func StartRootSpan(spanName string) (opentracing.Span, gocontext.Context) {
// 	span = tracer.StartSpan(spanName, opentracing.ChildOf(spanCtx))
// 	traceCtx := opentracing.ContextWithSpan(gocontext.Background(), span)
// }

func GetContextFrom(ctx context.Context) gocontext.Context {
	// TODO: fix crash when Middleware wasn't added to app.UseGlobal(...)
	return ctx.Values().Get("opentrace-ctx").(gocontext.Context)
}

func StartSpanFromContext(ctx context.Context, spanName string) (opentracing.Span, gocontext.Context) {
	tctx := GetContextFrom(ctx)
	return opentracing.StartSpanFromContext(tctx, spanName)
}

func Middleware(ctx context.Context) {
	carrier := opentracing.HTTPHeadersCarrier(ctx.Request().Header)
	spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, carrier)

	var span opentracing.Span

	path := reNum.ReplaceAllString(ctx.Path(), "{num}")
	path = reUuid.ReplaceAllString(path, "{uuid}")
	path = reHash.ReplaceAllString(path, "{hash}")

	spanName := fmt.Sprintf("HTTP request (%s: %s)", ctx.Method(), path)

	if err != nil {
		span = tracer.StartSpan(spanName)

		if err != opentracing.ErrSpanContextNotFound {
			span.SetTag("error", true)
			span.LogKV(
				"event", "error",
				"error", err.Error(),
			)
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.StopExecution()
		}
	} else {
		span = tracer.StartSpan(spanName, opentracing.ChildOf(spanCtx))
	}

	traceCtx := opentracing.ContextWithSpan(gocontext.Background(), span)

	ctx.Values().Set("opentrace-ctx", traceCtx)

	span.SetTag("path", ctx.Path()).
		SetTag("method", ctx.Method())

	defer func() {
		if r := recover(); r != nil {
			span.SetTag("panic", true).
				SetTag("error", true).
				SetTag("panic-message", r).
				LogKV("trace", string(debug.Stack()))
		}

		span.Finish()
	}()

	ctx.Next()
}

func NewTracerFromEnv() (opentracing.Tracer, io.Closer) {
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		log.Fatalf("Can't parse jaeger config from env vars: %s", err)
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatalf("Can't init jaeger tracing: %s", err)
	}

	return tracer, closer
}

func Jsonify(value interface{}) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	return string(bytes)
}

// TODO: rewrite whole module:
//	- remove init() function
//  - create middleware object instead of Middleware function
//  - move global tracer into middleware object
//  - delete this function
func SetTracer(t opentracing.Tracer, c io.Closer) {
	tracer = t
	closer = c
	opentracing.SetGlobalTracer(tracer)
}

func init() {
	tracer, closer = NewTracerFromEnv()
	opentracing.SetGlobalTracer(tracer)

	reNum = regexp.MustCompile(`\d+`)
	reUuid = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	reHash = regexp.MustCompile(`[0-9a-f]{32}`)
}
