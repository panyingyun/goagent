package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepDone    StepStatus = "done"
	StepFailed  StepStatus = "failed"
)

// Step 是工作流里的一个步骤:一个子问题 + 状态 + 结果。
type Step struct {
	ID       int        `json:"id"`
	Desc     string     `json:"desc"`
	Question string     `json:"question"`
	Status   StepStatus `json:"status"`
	Output   string     `json:"output,omitempty"`
}

// Workflow 是"任务 -> 步骤 -> 执行 -> 断点续跑"的完整状态机。
// Checkpoint 字段保存到文件,崩溃后可以从上次位置继续。
type Workflow struct {
	ID         string `json:"id"`
	Goal       string `json:"goal"`
	Steps      []Step `json:"steps"`
	Current    int    `json:"current"`
	Checkpoint string `json:"-"`
}

// Plan 让 LLM 把目标拆成 3~5 个步骤(JSON 输出),并初始化状态。
func (w *Workflow) Plan(ctx context.Context, llm ChatLLM) error {
	system := `你是任务规划器。把用户目标拆成 3~5 个可执行步骤。
只输出 JSON 数组,每个元素: {"desc": "步骤说明", "question": "该步骤要解决的子问题"}` +
		"\n目标: " + w.Goal
	reply, err := llm.Chat(ctx, []Message{{Role: "system", Content: system}})
	if err != nil {
		return err
	}
	var steps []struct {
		Desc     string `json:"desc"`
		Question string `json:"question"`
	}
	if err := json.Unmarshal([]byte(extractJSON(reply.Content)), &steps); err != nil {
		return fmt.Errorf("解析步骤失败: %w", err)
	}
	if len(steps) == 0 {
		return fmt.Errorf("规划结果为空")
	}
	w.Steps = make([]Step, len(steps))
	for i, s := range steps {
		w.Steps[i] = Step{ID: i, Desc: s.Desc, Question: s.Question, Status: StepPending}
	}
	w.Current = 0
	return w.Save()
}

// NextStep 返回下一个未完成的步骤;全部完成返回 nil。
func (w *Workflow) NextStep() *Step {
	for i := w.Current; i < len(w.Steps); i++ {
		if w.Steps[i].Status == StepPending {
			return &w.Steps[i]
		}
	}
	return nil
}

// RunStep 执行一个步骤并保存检查点。
func (w *Workflow) RunStep(ctx context.Context, llm ChatLLM, step *Step) error {
	// 把已完成步骤的产出作为上下文带给当前步骤,保证步骤之间能衔接
	contextMsg := "当前子任务: " + step.Question
	var b strings.Builder
	for _, s := range w.Steps {
		if s.Status == StepDone && s.Output != "" {
			b.WriteString(fmt.Sprintf("- %s: %s\n", s.Desc, s.Output))
		}
	}
	if b.Len() > 0 {
		contextMsg += "\n已完成步骤的产出:\n" + b.String()
	}
	reply, err := llm.Chat(ctx, []Message{
		{Role: "system", Content: "你是执行器,参考已完成步骤的产出,专注完成当前子任务,回答简洁。"},
		{Role: "user", Content: contextMsg},
	})
	if err != nil {
		step.Status = StepFailed
		_ = w.Save()
		return err
	}
	step.Output = strings.TrimSpace(reply.Content)
	step.Status = StepDone
	w.Current = step.ID + 1
	return w.Save()
}

// Resume 从检查点继续执行剩余步骤。
func (w *Workflow) Resume(ctx context.Context, llm ChatLLM) error {
	for step := w.NextStep(); step != nil; step = w.NextStep() {
		fmt.Fprintf(os.Stderr, "[步骤 %d/%d] %s\n", step.ID+1, len(w.Steps), step.Desc)
		if err := w.RunStep(ctx, llm, step); err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Save() error {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(w.Checkpoint, data, 0o644)
}

func LoadWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var w Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	w.Checkpoint = path
	return &w, nil
}

// extractJSON 抽出回复里第一个 JSON 对象或数组。
func extractJSON(s string) string {
	start := strings.IndexAny(s, "{[")
	if start < 0 {
		return ""
	}
	open, close := byte('{'), byte('}')
	if s[start] == '[' {
		open, close = '[', ']'
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}
