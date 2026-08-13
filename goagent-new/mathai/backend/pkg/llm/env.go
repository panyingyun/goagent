package llm

import "os"

// 截至 2026-08,DeepSeek 的模型名是 deepseek-v4-flash / deepseek-v4-pro
// (旧的 deepseek-chat / deepseek-reasoner 已于 2026-07-24 退役)。
const (
	DefaultBaseURL = "https://api.deepseek.com"
	DefaultModel   = "deepseek-v4-flash"
)

// NewHTTPClientFromEnv 从环境变量构造客户端。
// 优先读 DEEPSEEK_* ,兼容旧的 LLM_* ;URL 与模型有 DeepSeek 默认值。
func NewHTTPClientFromEnv() *HTTPClient {
	return NewHTTPClient(
		firstEnv("DEEPSEEK_API_KEY", "LLM_API_KEY"),
		envOr("DEEPSEEK_BASE_URL", envOr("LLM_BASE_URL", DefaultBaseURL)),
		envOr("DEEPSEEK_MODEL", envOr("LLM_MODEL", DefaultModel)),
	)
}

// Configured 表示至少具备发起请求所需的 API Key(URL/模型有默认值)。
func (c *HTTPClient) Configured() bool {
	return c != nil && c.APIKey != "" && c.BaseURL != "" && c.Model != ""
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
