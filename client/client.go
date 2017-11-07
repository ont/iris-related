package client

import (
	"bytes"
	"fmt"
	"github.com/kataras/iris/context"
	"github.com/ont/iris-related/middlewares/logging"
	"github.com/ont/iris-related/middlewares/requestid"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	urllib "net/url"
	"path"
	"time"
)

// Simple buffered client which returns strings and follows redirects.
// This client is mainly for adding X-Request-Id header to each http query.
type Client struct {
	BaseURL string // base url for doing request to (example "http://some-site.com")

	ctx       context.Context // current iris context
	requestId string          // request-id extracted from context
	log       *logrus.Entry   // logger extracted from context
}

func NewClient(ctx context.Context, baseUrl string) *Client {
	return &Client{
		BaseURL:   baseUrl,
		ctx:       ctx,
		requestId: requestid.Get(ctx),
		log:       logging.Get(ctx),
	}
}

// simple version of GET which only returns response body for 2xx response codes (and follows redirects)
func (c *Client) GET(url string) (string, error) {
	url, err := c.joinBaseUrl(url)
	if err != nil {
		return "", err
	}

	return c.doRequest("GET", url, nil)
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

	return c.doRequest("POST", url, bytes.NewBufferString(form.Encode()))
}

func (c *Client) doRequest(method string, url string, data io.Reader) (string, error) {
	c.log.WithField("http_method", method).Debugf("Request to %s", url)

	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return "", err
	}

	// TODO: headers for json (separate method doJsonRequest?)
	req.Header.Set("X-Request-Id", c.requestId) // add Request-Id for each request

	client := &http.Client{
		Timeout: 10 * time.Second, // NOTE: very important (default timeout is infinite)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	c.log.Debug("response Status: ", resp.Status)
	c.log.Debug("response Headers: ", resp.Header)

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
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
