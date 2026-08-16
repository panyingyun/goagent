package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var (
	identifierPattern = regexp.MustCompile(`[A-Za-z]+`)
	colorPattern      = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	implicitMultiply  = regexp.MustCompile(`[0-9)]\s*([A-Za-z]|\()`)
	valueWhitespace   = regexp.MustCompile(`\b(x|pi|e)\s+(x|pi|e|[A-Za-z]+\()`)
	valueParenthesis  = regexp.MustCompile(`\b(x|pi|e)\s*\(`)
)

var allowedIdentifiers = map[string]bool{
	"x": true, "t": true, "sin": true, "cos": true, "tan": true,
	"exp": true, "log": true, "sqrt": true, "abs": true,
	"pow": true, "pi": true, "e": true,
}

var graphPalette = []string{"#2563EB", "#DC2626", "#059669", "#7C3AED", "#EA580C", "#0891B2"}

func DefaultGraphState() GraphState {
	return GraphState{
		Equations: []Equation{},
		Viewport:  Viewport{XMin: -10, XMax: 10, YMin: -10, YMax: 10},
		Settings:  GraphSettings{ShowGrid: true, ShowAxes: true, ShowLegend: true, SampleCount: 700},
	}
}

func NormalizeExpression(raw string) (string, error) {
	expr := strings.TrimSpace(strings.NewReplacer("×", "*", "÷", "/", "²", "^2", "−", "-").Replace(raw))
	if len(expr) > 120 {
		return "", errors.New("expression exceeds 120 characters")
	}
	if index := strings.Index(expr, "="); index >= 0 {
		left := strings.TrimSpace(expr[:index])
		if strings.ToLower(left) != "y" || strings.Contains(expr[index+1:], "=") {
			return "", errors.New("only equations in the form y = f(x) are supported")
		}
		expr = strings.TrimSpace(expr[index+1:])
	}
	if expr == "" {
		return "", errors.New("expression is required")
	}
	depth := 0
	for _, char := range expr {
		switch {
		case char == '(':
			depth++
		case char == ')':
			depth--
			if depth < 0 {
				return "", errors.New("parentheses are not balanced")
			}
		case char >= '0' && char <= '9':
		case char >= 'A' && char <= 'Z':
		case char >= 'a' && char <= 'z':
		case strings.ContainsRune("+-*/^., \t", char):
		default:
			return "", fmt.Errorf("unsupported character %q", char)
		}
	}
	if depth != 0 {
		return "", errors.New("parentheses are not balanced")
	}
	for _, identifier := range identifierPattern.FindAllString(strings.ToLower(expr), -1) {
		if !allowedIdentifiers[identifier] {
			return "", fmt.Errorf("unsupported identifier %q", identifier)
		}
	}
	lower := strings.ToLower(expr)
	if implicitMultiply.MatchString(lower) || valueWhitespace.MatchString(lower) || valueParenthesis.MatchString(lower) {
		return "", errors.New("multiplication must use an explicit *")
	}
	return lower, nil
}

func ApplyGraphPatch(state GraphState, patch GraphPatch) (GraphState, error) {
	next := cloneGraphState(state)
	switch patch.Operation {
	case "replace_equations":
		if len(patch.Equations) == 0 || len(patch.Equations) > 12 {
			return state, errors.New("replace_equations requires 1 to 12 equations")
		}
		next.Equations = make([]Equation, 0, len(patch.Equations))
		for _, input := range patch.Equations {
			equation, err := newEquation(input, len(next.Equations))
			if err != nil {
				return state, err
			}
			next.Equations = append(next.Equations, equation)
		}
	case "add_equation":
		if patch.Equation == nil || len(next.Equations) >= 12 {
			return state, errors.New("add_equation requires an equation and the graph supports at most 12")
		}
		equation, err := newEquation(*patch.Equation, len(next.Equations))
		if err != nil {
			return state, err
		}
		next.Equations = append(next.Equations, equation)
	case "update_equation":
		index := equationIndex(next.Equations, patch.EquationID)
		if index < 0 || patch.Updates == nil {
			return state, errors.New("update_equation requires an existing equationId and updates")
		}
		if err := updateEquation(&next.Equations[index], *patch.Updates); err != nil {
			return state, err
		}
	case "remove_equation":
		index := equationIndex(next.Equations, patch.EquationID)
		if index < 0 {
			return state, errors.New("equationId does not exist")
		}
		next.Equations = append(next.Equations[:index], next.Equations[index+1:]...)
	case "set_viewport":
		if patch.Viewport == nil {
			return state, errors.New("set_viewport requires viewport")
		}
		applyViewportUpdate(&next.Viewport, *patch.Viewport)
	case "set_settings":
		if patch.Settings == nil {
			return state, errors.New("set_settings requires settings")
		}
		applySettingsUpdate(&next.Settings, *patch.Settings)
	case "set_analysis":
		if patch.Analysis == nil || strings.TrimSpace(patch.Analysis.Summary) == "" {
			return state, errors.New("set_analysis requires a non-empty analysis")
		}
		if len(patch.Analysis.Summary) > 1000 || len(patch.Analysis.Highlights) > 12 {
			return state, errors.New("analysis is too large")
		}
		for _, highlight := range patch.Analysis.Highlights {
			if len(highlight) > 240 {
				return state, errors.New("analysis highlight exceeds 240 characters")
			}
		}
		analysis := *patch.Analysis
		analysis.Summary = strings.TrimSpace(analysis.Summary)
		next.Analysis = &analysis
	case "clear_graph":
		next.Equations = []Equation{}
		next.Analysis = nil
	default:
		return state, fmt.Errorf("unsupported graph operation %q", patch.Operation)
	}
	if err := validateGraphState(next); err != nil {
		return state, err
	}
	return next, nil
}

func newEquation(input EquationInput, paletteIndex int) (Equation, error) {
	expression, err := NormalizeExpression(input.Expression)
	if err != nil {
		return Equation{}, err
	}
	visible, lineWidth := true, 2.5
	if input.Visible != nil {
		visible = *input.Visible
	}
	if input.LineWidth != nil {
		lineWidth = *input.LineWidth
	}
	color := input.Color
	if color == "" {
		color = graphPalette[paletteIndex%len(graphPalette)]
	}
	label := strings.TrimSpace(input.Label)
	if label == "" {
		label = "y = " + expression
	}
	equation := Equation{ID: newID("eq"), Expression: expression, Label: label, Color: color, Visible: visible, LineWidth: lineWidth}
	return equation, validateEquation(equation)
}

func updateEquation(equation *Equation, update EquationUpdate) error {
	if update.Expression != nil {
		expression, err := NormalizeExpression(*update.Expression)
		if err != nil {
			return err
		}
		equation.Expression = expression
	}
	if update.Label != nil {
		equation.Label = strings.TrimSpace(*update.Label)
	}
	if update.Color != nil {
		equation.Color = *update.Color
	}
	if update.Visible != nil {
		equation.Visible = *update.Visible
	}
	if update.LineWidth != nil {
		equation.LineWidth = *update.LineWidth
	}
	return validateEquation(*equation)
}

func validateGraphState(state GraphState) error {
	if len(state.Equations) > 12 {
		return errors.New("graph supports at most 12 equations")
	}
	for _, equation := range state.Equations {
		if err := validateEquation(equation); err != nil {
			return err
		}
	}
	if !finite(state.Viewport.XMin, state.Viewport.XMax, state.Viewport.YMin, state.Viewport.YMax) || state.Viewport.XMin >= state.Viewport.XMax || state.Viewport.YMin >= state.Viewport.YMax {
		return errors.New("viewport bounds must be finite and increasing")
	}
	if state.Viewport.XMax-state.Viewport.XMin > 1e6 || state.Viewport.YMax-state.Viewport.YMin > 1e6 {
		return errors.New("viewport range is too large")
	}
	if state.Settings.SampleCount < 100 || state.Settings.SampleCount > 4000 {
		return errors.New("sampleCount must be between 100 and 4000")
	}
	return nil
}

func validateEquation(equation Equation) error {
	if equation.ID == "" || equation.Expression == "" {
		return errors.New("equation id and expression are required")
	}
	if len(equation.Label) > 100 {
		return errors.New("equation label exceeds 100 characters")
	}
	if !colorPattern.MatchString(equation.Color) {
		return errors.New("equation color must be a six-digit hex color")
	}
	if equation.LineWidth < 1 || equation.LineWidth > 8 || math.IsNaN(equation.LineWidth) {
		return errors.New("lineWidth must be between 1 and 8")
	}
	return nil
}

func applyViewportUpdate(viewport *Viewport, update ViewportUpdate) {
	if update.XMin != nil {
		viewport.XMin = *update.XMin
	}
	if update.XMax != nil {
		viewport.XMax = *update.XMax
	}
	if update.YMin != nil {
		viewport.YMin = *update.YMin
	}
	if update.YMax != nil {
		viewport.YMax = *update.YMax
	}
}

func applySettingsUpdate(settings *GraphSettings, update SettingsUpdate) {
	if update.ShowGrid != nil {
		settings.ShowGrid = *update.ShowGrid
	}
	if update.ShowAxes != nil {
		settings.ShowAxes = *update.ShowAxes
	}
	if update.ShowLegend != nil {
		settings.ShowLegend = *update.ShowLegend
	}
	if update.SampleCount != nil {
		settings.SampleCount = *update.SampleCount
	}
}

func equationIndex(equations []Equation, id string) int {
	for index := range equations {
		if equations[index].ID == id {
			return index
		}
	}
	return -1
}

func cloneGraphState(state GraphState) GraphState {
	state.Equations = append(make([]Equation, 0, len(state.Equations)), state.Equations...)
	if state.Analysis != nil {
		analysis := *state.Analysis
		analysis.Highlights = append(make([]string, 0, len(state.Analysis.Highlights)), state.Analysis.Highlights...)
		state.Analysis = &analysis
	}
	return state
}

func finite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func newID(prefix string) string {
	buffer := make([]byte, 10)
	if _, err := rand.Read(buffer); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return prefix + "_" + hex.EncodeToString(buffer)
}
