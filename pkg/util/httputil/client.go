package httputil

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"e.coding.net/codingcorp/carctl/pkg/util/ioutils"
	"github.com/pkg/errors"
)

var (
	defaultHttpClient = &http.Client{
		Timeout: time.Minute * 5,
	}

	DefaultClient = &Client{client: defaultHttpClient}
)

type Client struct {
	client *http.Client
}

func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) GetWithAuth(url, username, password string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	return c.Do(req)
}

func (c *Client) Post(url, contentType string, body io.Reader, username, password string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	if contentType != "" {
		req.Header.Set(http.CanonicalHeaderKey("Content-Type"), contentType)
	}

	return c.Do(req)
}

func (c *Client) PostJson(url string, body io.Reader, username, password string) (resp *http.Response, err error) {
	return c.Post(url, "application/json;charset=UTF-8", body, username, password)
}

func (c *Client) PostForm(url string, data url.Values, username, password string) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded;charset=UTF-8", strings.NewReader(data.Encode()), username, password)
}

func (c *Client) Put(url, contentType string, body io.Reader, username, password string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	if contentType != "" {
		req.Header.Set(http.CanonicalHeaderKey("Content-Type"), contentType)
	}

	return c.Do(req)
}

func (c *Client) PutJson(url string, body io.Reader, username, password string) (resp *http.Response, err error) {
	return c.Put(url, "application/json;charset=UTF-8", body, username, password)
}

func (c *Client) PutFile(url, file, username, password string) (resp *http.Response, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer ioutils.QuiteClose(f)

	return c.Put(url, "", f, username, password)
}

func (c *Client) Patch(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set(http.CanonicalHeaderKey("Content-Type"), contentType)
	}

	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (resp *http.Response, err error) {
	if req == nil {
		return nil, errors.New("nil pointer exception: parameter 'req' is nil")
	}

	return c.client.Do(req)
}

func New() *Client {
	return &Client{
		client: defaultHttpClient,
	}
}
