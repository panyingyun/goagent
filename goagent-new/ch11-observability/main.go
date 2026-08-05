package main

import (
	"context"
	"errors"
	"os"
	"time"

	"goagent-new/pkg/agent"
	"goagent-new/pkg/trace"
)

func recordDemo(ctx context.Context, recorder *trace.Recorder) {
	ctx, endRun := recorder.Start(ctx, "agent.run", map[string]any{"prompt_version": "v1"})
	llmCtx, endLLM := recorder.Start(ctx, "llm.complete", map[string]any{"model": "configured-model"})
	time.Sleep(time.Millisecond)
	recorder.Record(llmCtx, agent.Event{Kind: "llm", Status: "ok", InputTokens: 20, OutputTokens: 5, Duration: time.Millisecond})
	endLLM(nil)
	toolCtx, endTool := recorder.Start(ctx, "tool.calculator", map[string]any{"side_effect": "none"})
	recorder.Record(toolCtx, agent.Event{Kind: "tool", Tool: "calculator", Status: "ok", Duration: time.Millisecond})
	endTool(nil)
	_, endFailed := recorder.Start(ctx, "tool.example_failure", nil)
	endFailed(errors.New("demonstration error"))
	endRun(nil)
}

func main() {
	recorder := trace.NewRecorder()
	recordDemo(context.Background(), recorder)
	if err := recorder.Export(os.Stdout); err != nil {
		panic(err)
	}
}

