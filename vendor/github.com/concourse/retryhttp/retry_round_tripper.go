package retryhttp

import (
	"io"
	"net/http"
	"time"

	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . Sleeper

type Sleeper interface {
	Sleep(time.Duration)
}

//go:generate counterfeiter . RoundTripper

type RoundTripper interface {
	RoundTrip(request *http.Request) (*http.Response, error)
}

type RetryRoundTripper struct {
	Logger       lager.Logger
	Sleeper      Sleeper
	RetryPolicy  RetryPolicy
	RoundTripper RoundTripper
}

type RetryReadCloser struct {
	io.ReadCloser
	IsRead bool
}

func (rrc *RetryReadCloser) Read(p []byte) (n int, err error) {
	rrc.IsRead = true
	return rrc.ReadCloser.Read(p)
}

func (d *RetryRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	retryReadCloser := &RetryReadCloser{request.Body, false}

	if request.Body != nil {
		request.Body = retryReadCloser
	}

	var response *http.Response
	var err error
	retry(d.Logger, d.RetryPolicy, d.Sleeper, func() bool {
		response, err = d.RoundTripper.RoundTrip(request)
		return err == nil || retryReadCloser.IsRead || !retryable(err)
	})

	return response, err
}
