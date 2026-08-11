package main

import (
	"bytes"
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
	entry := orderedObject{}.
		Set("ts", time.Now().Format(time.DateTime)).
		Set("level", level).
		Set("event", event)
	for k, v := range fields {
		entry = entry.Set(k, v)
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.w, string(data))
}

// orderedObject 按 Set 的先后顺序序列化 JSON 对象。
// Go 的 map 本身无序;标准库也没有 ordered map,用这个小类型即可。
type orderedObject []orderedField

type orderedField struct {
	Key   string
	Value any
}

func (o orderedObject) Set(key string, value any) orderedObject {
	return append(o, orderedField{Key: key, Value: value})
}

func (o orderedObject) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, f := range o {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(f.Key)
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(f.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
