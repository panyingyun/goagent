package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func main() {
	var request struct {
		Message string `json:"message"`
	}
	if err := DecodeJSONStrict(strings.NewReader(`{"message":"hello agent"}`), 1024, &request); err != nil {
		panic(err)
	}
	fmt.Println("严格 JSON:", request.Message)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	results, err := RunWorkers(ctx, 3, []int{1, 2, 3, 4, 5}, func(ctx context.Context, value int) (int, error) {
		select {
		case <-time.After(30 * time.Millisecond):
			return value * value, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("有限并发且保持顺序:", results)
	fmt.Println("运行 go test ./ch00-go-foundations -v 查看 HTTP 超时、取消和输入校验实验。")
}
