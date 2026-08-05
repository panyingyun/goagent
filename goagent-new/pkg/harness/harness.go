package harness

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

type Status string

const (
	Pending   Status = "pending"
	Running   Status = "running"
	Succeeded Status = "succeeded"
	Failed    Status = "failed"
	Canceled  Status = "canceled"
)

type Budget struct {
	MaxSteps    int           `json:"max_steps"`
	MaxFailures int           `json:"max_failures"`
	MaxCost     int           `json:"max_cost"`
	MaxDuration time.Duration `json:"max_duration"`
}

type Task struct {
	ID                   string   `json:"id"`
	Goal                 string   `json:"goal"`
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`
	Budget               Budget   `json:"budget"`
}

type Manifest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

type Environment struct {
	Manifest  Manifest          `json:"manifest"`
	Workspace string            `json:"workspace"`
	Values    map[string]string `json:"values,omitempty"`
}

type Action struct {
	Name     string `json:"name"`
	Summary  string `json:"summary"`
	Progress bool   `json:"progress"`
	Done     bool   `json:"done"`
	Cost     int    `json:"cost"`
	Error    string `json:"error,omitempty"`
}

type State struct {
	TaskID    string            `json:"task_id"`
	Status    Status            `json:"status"`
	Step      int               `json:"step"`
	Failures  int               `json:"failures"`
	Cost      int               `json:"cost"`
	Values    map[string]string `json:"values"`
	Actions   []Action          `json:"actions"`
	StartedAt time.Time         `json:"started_at"`
	EndedAt   time.Time         `json:"ended_at,omitempty"`
}

type Worker interface {
	Next(context.Context, Task, *State, Environment) (Action, error)
}

type WorkerFunc func(context.Context, Task, *State, Environment) (Action, error)

func (f WorkerFunc) Next(ctx context.Context, task Task, state *State, environment Environment) (Action, error) {
	return f(ctx, task, state, environment)
}

type Verifier interface {
	Verify(context.Context, Task, State) error
}

type VerifierFunc func(context.Context, Task, State) error

func (f VerifierFunc) Verify(ctx context.Context, task Task, state State) error {
	return f(ctx, task, state)
}

type Event struct {
	Kind   string `json:"kind"`
	Step   int    `json:"step"`
	Status Status `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type Observer interface {
	Record(context.Context, Event)
}

type ObserverFunc func(context.Context, Event)

func (f ObserverFunc) Record(ctx context.Context, event Event) { f(ctx, event) }

type Runner struct {
	Environment Environment
	Worker      Worker
	Verifiers   []Verifier
	Observer    Observer
	Now         func() time.Time
}

func (r Runner) Run(ctx context.Context, task Task) (State, error) {
	if err := validateTask(task, r.Environment, r.Worker); err != nil {
		return State{}, err
	}
	now := r.Now
	if now == nil {
		now = time.Now
	}
	state := State{TaskID: task.ID, Status: Running, Values: make(map[string]string), Actions: []Action{}, StartedAt: now().UTC()}
	r.record(ctx, Event{Kind: "run_started", Status: Running})

	runContext := ctx
	cancel := func() {}
	if task.Budget.MaxDuration > 0 {
		runContext, cancel = context.WithTimeout(ctx, task.Budget.MaxDuration)
	}
	defer cancel()

	stalled := 0
	for state.Step < task.Budget.MaxSteps {
		if err := runContext.Err(); err != nil {
			state.Status = Canceled
			state.EndedAt = now().UTC()
			r.record(ctx, Event{Kind: "run_canceled", Step: state.Step, Status: Canceled, Detail: err.Error()})
			return state, err
		}
		action, err := r.Worker.Next(runContext, task, &state, r.Environment)
		state.Step++
		state.Cost += action.Cost
		if err != nil {
			action.Error = err.Error()
			state.Failures++
		}
		state.Actions = append(state.Actions, action)
		r.record(runContext, Event{Kind: "action", Step: state.Step, Status: state.Status, Detail: action.Name})
		if task.Budget.MaxCost > 0 && state.Cost > task.Budget.MaxCost {
			return r.fail(ctx, state, now, "cost budget exceeded")
		}
		if err != nil {
			if state.Failures >= task.Budget.MaxFailures {
				return r.fail(ctx, state, now, fmt.Sprintf("failure budget exhausted: %v", err))
			}
			continue
		}
		if action.Progress {
			stalled = 0
		} else {
			stalled++
			if stalled >= 3 {
				return r.fail(ctx, state, now, "worker made no progress for three consecutive steps")
			}
		}
		if !action.Done {
			continue
		}
		for _, verifier := range r.Verifiers {
			if err := verifier.Verify(runContext, task, state); err != nil {
				return r.fail(ctx, state, now, "verification failed: "+err.Error())
			}
		}
		state.Status = Succeeded
		state.EndedAt = now().UTC()
		r.record(ctx, Event{Kind: "run_succeeded", Step: state.Step, Status: Succeeded})
		return state, nil
	}
	return r.fail(ctx, state, now, "step budget exhausted")
}

func validateTask(task Task, environment Environment, worker Worker) error {
	if task.ID == "" || task.Goal == "" {
		return errors.New("task id and goal are required")
	}
	if worker == nil {
		return errors.New("worker is required")
	}
	if task.Budget.MaxSteps <= 0 {
		return errors.New("max steps must be positive")
	}
	if task.Budget.MaxFailures <= 0 {
		return errors.New("max failures must be positive")
	}
	for _, required := range task.RequiredCapabilities {
		if !slices.Contains(environment.Manifest.Capabilities, required) {
			return fmt.Errorf("required capability %q is unavailable", required)
		}
	}
	return nil
}

func (r Runner) fail(ctx context.Context, state State, now func() time.Time, detail string) (State, error) {
	state.Status = Failed
	state.EndedAt = now().UTC()
	r.record(ctx, Event{Kind: "run_failed", Step: state.Step, Status: Failed, Detail: detail})
	return state, errors.New(detail)
}

func (r Runner) record(ctx context.Context, event Event) {
	if r.Observer != nil {
		r.Observer.Record(ctx, event)
	}
}
