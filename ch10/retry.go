package main

import (
	"context"
	"errors"
	"time"
)

// RetryableError 包装"值得重试"的错误(网络抖动、限流等)。
type RetryableError struct{ Err error }

func (e *RetryableError) Error() string { return "retryable: " + e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

// Retry 对可重试错误做指数退避重试;非可重试错误立即返回。
func Retry(ctx context.Context, attempts int, baseDelay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		var re *RetryableError
		if !errors.As(err, &re) {
			return err
		}
		delay := baseDelay * time.Duration(1<<i)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}
