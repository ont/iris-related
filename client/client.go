package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	urllib "net/url"
	"path"
	"time"

	gocontext "context"

	iriscontext "github.com/kataras/iris/v12/context"
	"github.com/ont/iris-related/logging"
	"github.com/ont/iris-related/requestid"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// Simple buffered client which returns strings and follows redirects.
// This client is mainly for adding X-Request-Id header to each http query.
type Client struct {
	BaseURL string // base url for doing request to (example "http://some-site.com")

	ctx       iriscontext.Context // current iris context
	requestId string              // request-id extracted from context
	log       *logrus.Entry       // logger extracted from context

	traceCtx gocontext.Context
}

func NewClient(baseUrl string) *Client {
	return &Client{
		BaseURL: baseUrl,
	}
}

func (c *Client) WithIris(ctx iriscontext.Context) *Client {
	c.ctx = ctx
	c.requestId = requestid.Get(ctx)
	c.log = logging.Get(ctx)

	return c
}

func (c *Client) WithTrace(traceCtx gocontext.Context) *Client {
	c.traceCtx = traceCtx
	return c
}

// simple version of GET which only returns response body for 2xx response codes (and follows redirects)
func (c *Client) GET(url string, params ...interface{}) (string, error) {
	url, err := c.joinBaseUrl(url)
	if err != nil {
		return "", err
	}

	vs := urllib.Values{}
	for i := 0; i < len(params); i += 2 {
		iname := params[i]
		ivalue := params[i+1]

		if name, ok := iname.(string); ok {
			vs.Add(name, fmt.Sprintf("%v", ivalue))
		} else {
			return "", fmt.Errorf("parameter name %v is not string", iname)
		}
	}

	if len(params) >= 2 {
		url += "?" + vs.Encode()
	}

	return c.doRequest("GET", url, "", nil)
}

// simple version of POST for sending ...
// it doesn't support sending array data
func (c *Client) POST(url string, data map[string]interface{}) (string, error) {
	url, err := c.joinBaseUrl(url)
	if err != nil {
		return "", err
	}

	form := urllib.Values{}
	for k, v := range data {
		form.Add(k, fmt.Sprintf("%v", v))
	}

	return c.doRequest("POST", url, form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
}

func (c *Client) doRequest(method string, url string, data string, headers map[string]string) (string, error) {
	var span opentracing.Span

	if c.traceCtx != nil {
		span, _ = opentracing.StartSpanFromContext(c.traceCtx, "doRequest")
		defer span.Finish()
	}

	c.log.WithField("http_method", method).Debugf("Request to %s", url)

	buffer := bytes.NewBufferString(data)

	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return "", err
	}

	// TODO: headers for json (separate method doJsonRequest?)
	if c.requestId != "" {
		req.Header.Set("X-Request-Id", c.requestId) // add Request-Id for each request
	}

	for name, value := range headers {
		req.Header.Add(name, value)
	}

	client := &http.Client{
		Timeout: 10 * time.Second, // NOTE: very important (default timeout is infinite)
	}

	// log info about request and inject span into HTTP headers
	if span != nil {
		span.SetTag("method", method).
			SetTag("url", url).
			LogKV(
				"event", "doing request",
				"method", method,
				"url", url,
				"data", data,
			)

		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header),
		)
	}

	resp, err := client.Do(req)
	if err != nil {

		if span != nil {
			span.SetTag("error", true).
				LogKV(
					"event", "error during http request",
					"error", err.Error(),
				)
		}
		return "", err
	}

	defer resp.Body.Close()

	if span != nil {
		span.LogKV(
			"event", "response from server",
			"status", resp.Status,
			"headers", resp.Header,
		)
	}
	c.log.Debug("response Status: ", resp.Status)
	c.log.Debug("response Headers: ", resp.Header)

	// check for http-code errors
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non 200 http code: %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if span != nil {
		span.LogKV(
			"event", "response body from server",
			"body", string(bytes),
		)
	}

	c.log.Debug("response Body: ", string(bytes))

	return string(bytes), nil
}

func (c *Client) joinBaseUrl(url string) (string, error) {
	u, err := urllib.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, url)
	return u.String(), nil
}
