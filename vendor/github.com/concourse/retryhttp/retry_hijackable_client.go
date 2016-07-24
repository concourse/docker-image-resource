package retryhttp

import (
	"net/http"

	"github.com/pivotal-golang/lager"
)

type RetryHijackableClient struct {
	Logger           lager.Logger
	Sleeper          Sleeper
	RetryPolicy      RetryPolicy
	HijackableClient HijackableClient
}

func (d *RetryHijackableClient) Do(request *http.Request) (*http.Response, HijackCloser, error) {
	var response *http.Response
	var hijackCloser HijackCloser
	var err error
	retry(d.Logger, d.RetryPolicy, d.Sleeper, func() bool {
		response, hijackCloser, err = d.HijackableClient.Do(request)
		return err == nil || !retryable(err)
	})

	return response, hijackCloser, err
}
