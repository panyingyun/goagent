package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const CurrentVersion = 1

// Status 表示工作流步骤的执行状态。
type Status string

const (
	Pending         Status = "pending"
	Running         Status = "running"
	WaitingApproval Status = "waiting_approval"
	Done            Status = "done"
	Failed          Status = "failed"
)

// Step 是工作流中的单个可检查点步骤。
type Step struct {
	ID               string `json:"id"`
	Status           Status `json:"status"`
	RequiresApproval bool   `json:"requires_approval,omitempty"`
	IdempotencyKey   string `json:"idempotency_key,omitempty"`
	Output           string `json:"output,omitempty"`
	Error            string `json:"error,omitempty"`
}

// State 是可持久化的工作流运行状态。
type State struct {
	Version       int             `json:"version"`
	ID            string          `json:"id"`
	Goal          string          `json:"goal"`
	Current       int             `json:"current"`
	Steps         []Step          `json:"steps"`
	CompletedKeys map[string]bool `json:"completed_keys"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Handler 执行某个步骤并返回输出文本。
type Handler func(context.Context, *State, *Step) (string, error)

// Approver 在 RequiresApproval 步骤上决定是否继续。
type Approver interface {
	Approve(context.Context, State, Step) (bool, error)
}

// Engine 按步骤推进工作流,并在关键节点写入检查点。
type Engine struct {
	Checkpoint string
	Handlers   map[string]Handler
	Approver   Approver
}

func NewState(id, goal string, steps []Step) *State {
	for i := range steps {
		if steps[i].Status == "" {
			steps[i].Status = Pending
		}
	}
	return &State{Version: CurrentVersion, ID: id, Goal: goal, Steps: steps, CompletedKeys: make(map[string]bool)}
}

func (e *Engine) Run(ctx context.Context, state *State) error {
	if state.Version != CurrentVersion {
		return fmt.Errorf("unsupported workflow version %d", state.Version)
	}
	for state.Current < len(state.Steps) {
		if err := ctx.Err(); err != nil {
			return err
		}
		step := &state.Steps[state.Current]
		if step.Status == Done || (step.IdempotencyKey != "" && state.CompletedKeys[step.IdempotencyKey]) {
			step.Status = Done
			state.Current++
			continue
		}
		if step.RequiresApproval {
			step.Status = WaitingApproval
			if err := e.Save(state); err != nil {
				return err
			}
			if e.Approver == nil {
				return errors.New("workflow is waiting for approval")
			}
			approved, err := e.Approver.Approve(ctx, *state, *step)
			if err != nil {
				return err
			}
			if !approved {
				return errors.New("workflow step denied")
			}
		}
		handler, ok := e.Handlers[step.ID]
		if !ok {
			return fmt.Errorf("no handler for step %q", step.ID)
		}
		step.Status = Running
		if err := e.Save(state); err != nil {
			return err
		}
		out, err := handler(ctx, state, step)
		if err != nil {
			step.Status = Failed
			step.Error = err.Error()
			_ = e.Save(state)
			return err
		}
		step.Output = out
		step.Status = Done
		step.Error = ""
		if step.IdempotencyKey != "" {
			state.CompletedKeys[step.IdempotencyKey] = true
		}
		state.Current++
		if err := e.Save(state); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) Save(state *State) error {
	if e.Checkpoint == "" {
		return errors.New("checkpoint path is required")
	}
	state.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return replaceFile(e.Checkpoint, data)
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	if state.CompletedKeys == nil {
		state.CompletedKeys = make(map[string]bool)
	}
	return &state, nil
}

func replaceFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".checkpoint-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

