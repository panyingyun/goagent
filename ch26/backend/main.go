// ch26 综合案例:AI 数学方程绘图智能体(后端)
//
// Go + SQLite + DeepSeek 实现设计稿 sw-design.md 的后端部分:
//   - 会话 CRUD(SQLite 持久化)
//   - /api/chat:自然语言 -> DeepSeek -> 结构化 JSON -> 更新 graphState
//   - 9 条验收用例见 api_test.go
//
// 运行:
//
//	go run .
//
// 前端(web/):npm install && npm run dev(代理 /api 到本服务)
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	store, err := OpenStore(envOr("DATABASE_PATH", "math_graph_ai.db"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "打开数据库失败:", err)
		os.Exit(1)
	}
	defer store.Close()

	llm := NewDeepSeekClient()
	if llm.APIKey == "" {
		fmt.Fprintln(os.Stderr, "警告:未设置 DEEPSEEK_API_KEY,/api/chat 将返回 503")
	}
	srv := &Server{store: store, llm: llm}
	port := envOr("PORT", "8080")
	fmt.Println("MathGraph AI 后端已启动: http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, srv.routes()); err != nil {
		fmt.Fprintln(os.Stderr, "服务启动失败:", err)
		os.Exit(1)
	}
}
