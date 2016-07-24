package retryhttp

import (
	"errors"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/pivotal-golang/lager"
)

func retry(logger lager.Logger, retryPolicy RetryPolicy, sleeper Sleeper, action func() bool) {
	retryLogger := logger.Session("retry")
	startTime := time.Now()

	var failedAttempts uint
	for {
		if action() {
			break
		}

		failedAttempts++

		delay, keepRetrying := retryPolicy.DelayFor(failedAttempts)
		if !keepRetrying {
			retryLogger.Error("giving-up", errors.New("giving up"), lager.Data{
				"total-failed-attempts": failedAttempts,
				"ran-for":               time.Now().Sub(startTime).String(),
			})

			break
		}

		retryLogger.Info("retrying", lager.Data{
			"failed-attempts": failedAttempts,
			"next-attempt-in": delay.String(),
			"ran-for":         time.Now().Sub(startTime).String(),
		})

		sleeper.Sleep(delay)
	}
}

func retryable(err error) bool {
	if neterr, ok := err.(net.Error); ok {
		if neterr.Temporary() {
			return true
		}
	}

	s := err.Error()
	for _, retryableError := range retryableErrors {
		if strings.HasSuffix(s, retryableError.Error()) {
			return true
		}
	}

	return false
}

var retryableErrors = []error{
	syscall.ECONNREFUSED,
	syscall.ECONNRESET,
	syscall.ETIMEDOUT,
	errors.New("i/o timeout"),
	errors.New("no such host"),
	errors.New("remote error: handshake failure"),
}
