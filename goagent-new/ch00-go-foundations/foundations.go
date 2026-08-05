package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// DecodeJSONStrict 对请求体做三层约束：大小、未知字段、单一 JSON 值。
// 后续所有 Agent HTTP 接口都应采用类似边界，而不是直接 json.Decode。
func DecodeJSONStrict(r io.Reader, maxBytes int64, dst any) error {
	if maxBytes <= 0 {
		return errors.New("maxBytes must be positive")
	}
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return fmt.Errorf("read JSON: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return fmt.Errorf("JSON exceeds %d-byte limit", maxBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("body must contain exactly one JSON value")
	}
	return nil
}

// FetchText 展示一个受调用方 context 控制、限制响应体大小的 HTTP 调用。
func FetchText(ctx context.Context, client *http.Client, url string, maxBytes int64) (string, error) {
	if client == nil {
		return "", errors.New("http client is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return "", fmt.Errorf("response exceeds %d-byte limit", maxBytes)
	}
	return string(data), nil
}

type indexedJob struct {
	index int
	value int
}

type indexedResult struct {
	index int
	value int
	err   error
}

type WorkFunc func(context.Context, int) (int, error)

// RunWorkers 用固定数量的 worker 执行任务，结果仍按输入顺序返回。
// 任一任务失败后会取消其他任务；调用方取消也会传播到所有 worker。
func RunWorkers(ctx context.Context, workerCount int, values []int, work WorkFunc) ([]int, error) {
	if workerCount <= 0 {
		return nil, errors.New("workerCount must be positive")
	}
	if work == nil {
		return nil, errors.New("work function is required")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan indexedJob)
	results := make(chan indexedResult)

	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobs {
				value, err := work(ctx, job.value)
				select {
				case results <- indexedResult{index: job.index, value: value, err: err}:
				case <-ctx.Done():
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for index, value := range values {
			select {
			case jobs <- indexedJob{index: index, value: value}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(results)
	}()

	ordered := make([]int, len(values))
	var firstErr error
	for result := range results {
		if result.err != nil && firstErr == nil {
			firstErr = result.err
			cancel()
			continue
		}
		ordered[result.index] = result.value
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ordered, nil
}

