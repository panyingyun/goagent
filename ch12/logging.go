package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
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
	// map + json.Marshal 会按 key 字典序输出,ts 往往排到最后。
	// 这里按固定顺序手写 JSON:ts → level → event → 其余字段(字典序)。
	type kv struct {
		k string
		v any
	}
	items := []kv{
		{"ts", time.Now().Format(time.RFC3339)},
		{"level", level},
		{"event", event},
	}
	extra := make([]string, 0, len(fields))
	for k := range fields {
		extra = append(extra, k)
	}
	sort.Strings(extra)
	for _, k := range extra {
		items = append(items, kv{k, fields[k]})
	}

	buf := make([]byte, 0, 128)
	buf = append(buf, '{')
	for i, it := range items {
		if i > 0 {
			buf = append(buf, ',')
		}
		kb, _ := json.Marshal(it.k)
		vb, _ := json.Marshal(it.v)
		buf = append(buf, kb...)
		buf = append(buf, ':')
		buf = append(buf, vb...)
	}
	buf = append(buf, '}')
	fmt.Fprintln(l.w, string(buf))
}
