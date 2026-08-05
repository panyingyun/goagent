package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Logger 输出一行一个 JSON 的结构化日志,方便机器解析和排查问题。
type Logger struct {
	w io.Writer
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

func (l *Logger) Log(level, event string, fields map[string]any) {
	entry := map[string]any{
		"ts":    time.Now().Format(time.RFC3339),
		"level": level,
		"event": event,
	}
	for k, v := range fields {
		entry[k] = v
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.w, string(data))
}
