package diff

import (
	"context"
	"net/http"
	"net/http/httputil"

	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
)

type httpClient interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

func newRetriableHTTPClient(retry map[int]struct{}) httpClient {
	c := retryablehttp.NewClient()
	c.Logger = nil
	c.CheckRetry = newRetryPolicy(retry)
	return c
}

func newRetryPolicy(retry map[int]struct{}) retryablehttp.CheckRetry {
	return func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// do not retry on context.Canceled or context.DeadlineExceeded
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if _, ok := retry[resp.StatusCode]; ok {
			log.Info("Retrying")
			return true, nil
		}

		return false, nil
	}
}

func httpTraceReq(req *http.Request) {
	reqStr, _ := httputil.DumpRequestOut(req, true)
	log.Tracef("---TRACE REQUEST---\n%s\n--- END ---\n\n", reqStr)
}
func httpTraceResp(resp *http.Response) {
	respStr, _ := httputil.DumpResponse(resp, true)
	log.Tracef("---TRACE RESPONSE---\n%s\n--- END ---\n\n", respStr)
}
