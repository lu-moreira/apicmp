package diff

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"runtime"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
)

type httpClient interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

func newRetriableHTTPClient(retry map[int]struct{}) httpClient {
	// default pooled client
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	var defaultRetryWaitMin = 1 * time.Second
	var defaultRetryWaitMax = 30 * time.Second

	c := &retryablehttp.Client{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Minute,
		},
		Logger:       nil,
		RetryWaitMin: defaultRetryWaitMin,
		RetryWaitMax: defaultRetryWaitMax,
		RetryMax:     1,
		CheckRetry:   newRetryPolicy(retry),
		Backoff:      retryablehttp.DefaultBackoff,
	}
	return c
}

func newRetryPolicy(retry map[int]struct{}) retryablehttp.CheckRetry {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// do not retry on context.Canceled or context.DeadlineExceeded
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if resp != nil {
			if _, ok := retry[resp.StatusCode]; ok {
				log.Info("Retrying")
				return true, nil
			}
		}

		return false, nil
	}
}

func httpTraceReq(req *retryablehttp.Request) {
	if bs, err := req.BodyBytes(); err == nil {
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
	}

	reqStr, _ := httputil.DumpRequestOut(req.Request, true)
	log.Tracef("---TRACE REQUEST---\n%s\n--- END ---\n\n", reqStr)
}
func httpTraceResp(resp *http.Response) {
	respStr, _ := httputil.DumpResponse(resp, true)
	log.Tracef("---TRACE RESPONSE---\n%s\n--- END ---\n\n", respStr)
}
