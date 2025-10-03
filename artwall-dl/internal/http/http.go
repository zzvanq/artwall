package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func GetBodyReader(ctx context.Context, url string) (io.ReadCloser, error) {
	return getBodyWithRetries(ctx, url, 3, 500*time.Millisecond)
}

func GetBody(ctx context.Context, url string) ([]byte, error) {
	r, err := GetBodyReader(ctx, url)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getBodyWithRetries(ctx context.Context, url string, retries int, backoff time.Duration) (io.ReadCloser, error) {
	var lastErr error

	for attempt := range retries {
		body, statusCode, err := fetchBody(ctx, url)
		if err == nil {
			return body, nil
		}

		lastErr = err

		if !isRetryable(err, statusCode) {
			return nil, err
		}

		jitter := time.Duration(rand.Intn(int(backoff / 2)))
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("tried %d times: %w", attempt+1, ctx.Err())
		case <-time.After(backoff + jitter):
			backoff *= 2
		}
	}

	return nil, fmt.Errorf("tried %d times: %w", retries, lastErr)
}

func fetchBody(ctx context.Context, url string) (io.ReadCloser, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(fmt.Errorf("failed to build request: %w", err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != 200 {
		return nil, resp.StatusCode, fmt.Errorf("status code %d", resp.StatusCode)
	}

	return resp.Body, resp.StatusCode, nil
}

func isRetryable(err error, statusCode int) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	if statusCode >= 500 {
		return true
	}

	var netErr *url.Error
	if errors.As(err, &netErr) {
		return true
	}

	return false
}
