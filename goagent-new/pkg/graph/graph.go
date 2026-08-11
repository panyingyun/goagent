package graph

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

// Status 表示图运行的生命周期状态。
type Status string

const (
	Pending   Status = "pending"
	Running   Status = "running"
	Succeeded Status = "succeeded"
	Failed    Status = "failed"
)

// State 是图执行的可检查点状态(数据、输出、路径与幂等键)。
type State struct {
	GraphVersion  int               `json:"graph_version"`
	RunID         string            `json:"run_id"`
	Current       string            `json:"current,omitempty"`
	Status        Status            `json:"status"`
	Step          int               `json:"step"`
	Data          map[string]string `json:"data"`
	Outputs       map[string]string `json:"outputs"`
	CompletedKeys map[string]string `json:"completed_keys"`
	Path          []string          `json:"path"`
	Error         string            `json:"error,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// Handler 执行某个图节点并返回输出。
type Handler func(context.Context, *State) (string, error)

// Predicate 判断某条边在当前状态下是否可走。
type Predicate func(State) bool

// KeyFunc 为节点计算幂等键,用于恢复时跳过已完成副作用。
type KeyFunc func(State) string

// Node 是图中的一个可执行节点。
type Node struct {
	ID             string
	Run            Handler
	IdempotencyKey KeyFunc
}

// Edge 是带可选条件的有向边。
type Edge struct {
	From string
	To   string
	Name string
	When Predicate
}

// Definition 描述一张可验证、可执行的图(入口、节点与边)。
type Definition struct {
	Name    string
	Version int
	Entry   string
	Nodes   map[string]Node
	Edges   []Edge
}

// Event 是图执行产生的事件日志条目。
type Event struct {
	RunID  string    `json:"run_id"`
	Step   int       `json:"step"`
	Node   string    `json:"node"`
	Kind   string    `json:"kind"`
	Output string    `json:"output,omitempty"`
	Error  string    `json:"error,omitempty"`
	At     time.Time `json:"at"`
}

// Checkpointer 持久化图状态。
type Checkpointer interface {
	Save(State) error
}

// EventSink 追加图事件日志。
type EventSink interface {
	Append(Event) error
}

// Engine 按定义逐步执行图,并可选写入检查点与事件。
type Engine struct {
	Definition   Definition
	MaxSteps     int
	Checkpointer Checkpointer
	Events       EventSink
	Now          func() time.Time
}

func NewState(definition Definition, runID string, data map[string]string) *State {
	cloned := make(map[string]string, len(data))
	for key, value := range data {
		cloned[key] = value
	}
	return &State{
		GraphVersion: definition.Version, RunID: runID, Current: definition.Entry, Status: Pending,
		Data: cloned, Outputs: make(map[string]string), CompletedKeys: make(map[string]string), Path: []string{},
	}
}

func (d Definition) Validate() error {
	if d.Name == "" || d.Version <= 0 || d.Entry == "" {
		return errors.New("graph name, positive version and entry are required")
	}
	if len(d.Nodes) == 0 {
		return errors.New("graph requires at least one node")
	}
	if _, ok := d.Nodes[d.Entry]; !ok {
		return fmt.Errorf("entry node %q does not exist", d.Entry)
	}
	for id, node := range d.Nodes {
		if id == "" || node.ID != id || node.Run == nil {
			return fmt.Errorf("node %q must have a matching id and handler", id)
		}
	}
	adjacency := make(map[string][]string)
	terminal := false
	for _, edge := range d.Edges {
		if _, ok := d.Nodes[edge.From]; !ok {
			return fmt.Errorf("edge source %q does not exist", edge.From)
		}
		if _, ok := d.Nodes[edge.To]; !ok {
			return fmt.Errorf("edge target %q does not exist", edge.To)
		}
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	for id := range d.Nodes {
		if len(adjacency[id]) == 0 {
			terminal = true
		}
	}
	if !terminal {
		return errors.New("graph requires at least one terminal node")
	}
	visited := map[string]bool{d.Entry: true}
	queue := []string{d.Entry}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range adjacency[current] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	if len(visited) != len(d.Nodes) {
		unreachable := make([]string, 0)
		for id := range d.Nodes {
			if !visited[id] {
				unreachable = append(unreachable, id)
			}
		}
		sort.Strings(unreachable)
		return fmt.Errorf("unreachable nodes: %v", unreachable)
	}
	return nil
}

func (e Engine) Run(ctx context.Context, state *State) error {
	if err := e.prepare(state); err != nil {
		return err
	}
	for state.Status != Succeeded {
		done, err := e.Step(ctx, state)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
	return nil
}

func (e Engine) Step(ctx context.Context, state *State) (bool, error) {
	if err := e.prepare(state); err != nil {
		return false, err
	}
	if state.Status == Succeeded {
		return true, nil
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if state.Step >= e.maxSteps() {
		return false, e.fail(state, "graph step budget exhausted")
	}
	state.Status = Running
	state.Error = ""
	node := e.Definition.Nodes[state.Current]
	if err := e.emit(Event{RunID: state.RunID, Step: state.Step, Node: node.ID, Kind: "node_started", At: e.now()}); err != nil {
		return false, err
	}
	output, err := e.runNode(ctx, node, state)
	if err != nil {
		state.Status = Failed
		state.Error = err.Error()
		state.UpdatedAt = e.now()
		_ = e.emit(Event{RunID: state.RunID, Step: state.Step, Node: node.ID, Kind: "node_failed", Error: err.Error(), At: e.now()})
		_ = e.save(*state)
		return false, err
	}
	state.Outputs[node.ID] = output
	state.Path = append(state.Path, node.ID)
	state.Step++
	if err := e.emit(Event{RunID: state.RunID, Step: state.Step, Node: node.ID, Kind: "node_succeeded", Output: output, At: e.now()}); err != nil {
		return false, err
	}
	next, err := e.next(*state, node.ID)
	if err != nil {
		return false, e.fail(state, err.Error())
	}
	state.Current = next
	if next == "" {
		state.Status = Succeeded
	}
	state.UpdatedAt = e.now()
	if err := e.save(*state); err != nil {
		return false, err
	}
	return state.Status == Succeeded, nil
}

func (e Engine) prepare(state *State) error {
	if state == nil {
		return errors.New("graph state is required")
	}
	if err := e.Definition.Validate(); err != nil {
		return err
	}
	if state.GraphVersion != e.Definition.Version {
		return fmt.Errorf("state graph version %d does not match definition version %d", state.GraphVersion, e.Definition.Version)
	}
	if state.Current == "" && state.Status != Succeeded {
		return errors.New("non-terminal graph state has no current node")
	}
	if _, ok := e.Definition.Nodes[state.Current]; !ok && state.Status != Succeeded {
		return fmt.Errorf("current node %q does not exist", state.Current)
	}
	if state.Data == nil {
		state.Data = make(map[string]string)
	}
	if state.Outputs == nil {
		state.Outputs = make(map[string]string)
	}
	if state.CompletedKeys == nil {
		state.CompletedKeys = make(map[string]string)
	}
	return nil
}

func (e Engine) runNode(ctx context.Context, node Node, state *State) (string, error) {
	key := ""
	if node.IdempotencyKey != nil {
		key = node.IdempotencyKey(*state)
		if output, ok := state.CompletedKeys[key]; key != "" && ok {
			if err := e.emit(Event{RunID: state.RunID, Step: state.Step, Node: node.ID, Kind: "node_reused", Output: output, At: e.now()}); err != nil {
				return "", err
			}
			return output, nil
		}
	}
	output, err := node.Run(ctx, state)
	if err == nil && key != "" {
		state.CompletedKeys[key] = output
	}
	return output, err
}

func (e Engine) next(state State, from string) (string, error) {
	matches := make([]Edge, 0, 1)
	outgoing := 0
	for _, edge := range e.Definition.Edges {
		if edge.From != from {
			continue
		}
		outgoing++
		if edge.When == nil || edge.When(state) {
			matches = append(matches, edge)
		}
	}
	if len(matches) == 0 {
		if outgoing > 0 {
			return "", fmt.Errorf("no route matched from %q", from)
		}
		return "", nil
	}
	if len(matches) > 1 {
		names := make([]string, len(matches))
		for index, edge := range matches {
			names[index] = edge.Name
		}
		return "", fmt.Errorf("ambiguous route from %q: %v", from, names)
	}
	return matches[0].To, nil
}

func (e Engine) fail(state *State, message string) error {
	state.Status = Failed
	state.Error = message
	state.UpdatedAt = e.now()
	_ = e.save(*state)
	return errors.New(message)
}

func (e Engine) save(state State) error {
	if e.Checkpointer == nil {
		return nil
	}
	return e.Checkpointer.Save(state)
}

func (e Engine) emit(event Event) error {
	if e.Events == nil {
		return nil
	}
	return e.Events.Append(event)
}

func (e Engine) maxSteps() int {
	if e.MaxSteps <= 0 {
		return 32
	}
	return e.MaxSteps
}

func (e Engine) now() time.Time {
	if e.Now == nil {
		return time.Now().UTC()
	}
	return e.Now().UTC()
}
