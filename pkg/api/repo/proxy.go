package repo

import (
	"net/http"
	"time"

	"github.com/coding-wepack/carctl/pkg/constants"
	repotypes "github.com/coding-wepack/carctl/pkg/types/repo"
	"github.com/coding-wepack/carctl/pkg/util/httputil"
	"github.com/coding-wepack/carctl/pkg/util/jsonutil"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

const (
	defaultTimeout    = time.Second * 15
	defaultRetryCount = 3
)

func GetRepoProxySourceList(proxyUrl, cookie string) ([]repotypes.ProxySource, error) {
	xsrfToken := httputil.GetXsrfToken(cookie)

	resp, err := resty.New().
		SetTimeout(defaultTimeout).
		SetRetryCount(defaultRetryCount).
		SetHeader("Cookie", cookie).
		SetHeader("X-Xsrf-Token", xsrfToken).
		R().
		Get(proxyUrl)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("unexpected response status: GET %s: %s", proxyUrl, resp.Status())
	}

	var proxySourceResp repotypes.ProxySourceResponse
	if err = jsonutil.Unmarshal(resp.Body(), &proxySourceResp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if proxySourceResp.Code != constants.BizCodeSucceeded {
		return nil, errors.New(string(resp.Body()))
	}

	return proxySourceResp.Data, nil
}

func AddProxySource(proxyUrl, cookie string, payload *repotypes.ProxySourcePayload) error {
	xsrfToken := httputil.GetXsrfToken(cookie)

	resp, err := resty.New().
		SetTimeout(defaultTimeout).
		SetRetryCount(defaultRetryCount).
		SetHeader("Cookie", cookie).
		SetHeader("X-Xsrf-Token", xsrfToken).
		R().
		SetBody(payload).
		Post(proxyUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.Errorf("unexpected response status: POST %s: %s", proxyUrl, resp.Status())
	}

	var proxySourceResp repotypes.PostProxySourceResponse
	if err = jsonutil.Unmarshal(resp.Body(), &proxySourceResp); err != nil {
		return errors.Wrap(err, "failed to unmarshal response body")
	}

	if proxySourceResp.Code != constants.BizCodeSucceeded {
		return errors.New(string(resp.Body()))
	}

	return nil
}
