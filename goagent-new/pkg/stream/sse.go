package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Event 是推送给 SSE 客户端的单条事件。
type Event struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
}

// Relay 把事件通道转为 SSE。客户端断开时立即返回 context 错误。
// 事件 data 使用 JSON 编码，避免换行破坏 SSE 帧。
func Relay(ctx context.Context, w http.ResponseWriter, events <-chan Event) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("response writer does not support flushing")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-events:
			if !ok {
				return nil
			}
			payload, err := json.Marshal(event)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload); err != nil {
				return err
			}
			flusher.Flush()
		}
	}
}

