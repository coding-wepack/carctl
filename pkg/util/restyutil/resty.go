package restyutil

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultTimeout = time.Second * 15
)

var DefaultClient = resty.New()

func Get(url string) (*resty.Response, error) {
	return DefaultClient.R().Get(url)
}

func GetWithCtx(ctx context.Context, url string) (*resty.Response, error) {
	return DefaultClient.R().SetContext(ctx).Get(url)
}

func GetWithTimeout(url string, timeout time.Duration) (*resty.Response, error) {
	return resty.New().SetTimeout(timeout).R().Get(url)
}

func GetWithRetry(url string, retry int) (*resty.Response, error) {
	return resty.New().SetRetryCount(retry).R().Get(url)
}
