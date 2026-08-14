package main

import (
	"context"
	"errors"

	"mathai/pkg/llm"
)

type scriptedClient struct {
	responses []llm.Response
	requests  []llm.Request
}

func (c *scriptedClient) Complete(_ context.Context, request llm.Request) (llm.Response, error) {
	c.requests = append(c.requests, request)
	if len(c.responses) == 0 {
		return llm.Response{}, errors.New("no scripted response")
	}
	response := c.responses[0]
	c.responses = c.responses[1:]
	return response, nil
}
