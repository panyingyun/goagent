package llm

import (
	"os"
	"testing"
)

func TestNewHTTPClientFromEnvDefaults(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "sk-test")
	os.Unsetenv("LLM_API_KEY")
	os.Unsetenv("DEEPSEEK_BASE_URL")
	os.Unsetenv("LLM_BASE_URL")
	os.Unsetenv("DEEPSEEK_MODEL")
	os.Unsetenv("LLM_MODEL")

	c := NewHTTPClientFromEnv()
	if c.APIKey != "sk-test" {
		t.Fatalf("api key = %q", c.APIKey)
	}
	if c.BaseURL != DefaultBaseURL {
		t.Fatalf("base url = %q", c.BaseURL)
	}
	if c.Model != DefaultModel {
		t.Fatalf("model = %q", c.Model)
	}
	if !c.Configured() {
		t.Fatal("expected configured")
	}
}

func TestNewHTTPClientFromEnvFallsBackToLLMVars(t *testing.T) {
	os.Unsetenv("DEEPSEEK_API_KEY")
	t.Setenv("LLM_API_KEY", "legacy-key")
	t.Setenv("LLM_BASE_URL", "https://example.test")
	t.Setenv("LLM_MODEL", "custom-model")

	c := NewHTTPClientFromEnv()
	if c.APIKey != "legacy-key" || c.BaseURL != "https://example.test" || c.Model != "custom-model" {
		t.Fatalf("unexpected client: %+v", c)
	}
}
