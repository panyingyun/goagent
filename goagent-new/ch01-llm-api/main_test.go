package main

import (
	"context"
	"testing"

	"goagent-new/pkg/llm"
)

type fakeLLM struct {
	request llm.Request
}

func (f *fakeLLM) Complete(_ context.Context, request llm.Request) (llm.Response, error) {
	f.request = request
	return llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "answer"}, Usage: llm.Usage{TotalTokens: 3}}, nil
}

func TestAskBuildsMessageList(t *testing.T) {
	fake := &fakeLLM{}
	response, err := ask(context.Background(), fake, "question")
	if err != nil {
		t.Fatal(err)
	}
	if response.Message.Content != "answer" || len(fake.request.Messages) != 2 || fake.request.Messages[1].Content != "question" {
		t.Fatalf("unexpected request/response: %+v %+v", fake.request, response)
	}
}

