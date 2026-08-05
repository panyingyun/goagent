package main

import "testing"

func TestNormalizeExpression(t *testing.T) {
	t.Parallel()
	tests := []struct{ input, want string }{
		{"y = x² + 2*x", "x^2 + 2*x"},
		{"SIN(x) + pi", "sin(x) + pi"},
	}
	for _, test := range tests {
		got, err := NormalizeExpression(test.input)
		if err != nil {
			t.Fatalf("NormalizeExpression(%q): %v", test.input, err)
		}
		if got != test.want {
			t.Fatalf("NormalizeExpression(%q) = %q, want %q", test.input, got, test.want)
		}
	}
	for _, invalid := range []string{"y = os.exit()", "2x+1", "x sin(x)", "pi(x)", "x;alert(1)", "x = y = 1", "sqrt(x"} {
		if _, err := NormalizeExpression(invalid); err == nil {
			t.Errorf("NormalizeExpression(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestApplyGraphPatchIsAtomic(t *testing.T) {
	t.Parallel()
	state := DefaultGraphState()
	state.Equations = []Equation{{ID: "eq_original", Expression: "x", Label: "y = x", Color: "#2563EB", Visible: true, LineWidth: 2}}
	_, err := ApplyGraphPatch(state, GraphPatch{Operation: "replace_equations", Equations: []EquationInput{{Expression: "x^2"}, {Expression: "bad(x)"}}})
	if err == nil {
		t.Fatal("invalid patch unexpectedly succeeded")
	}
	if len(state.Equations) != 1 || state.Equations[0].ID != "eq_original" {
		t.Fatal("failed patch mutated the original state")
	}
}

func TestApplyGraphPatchLifecycle(t *testing.T) {
	t.Parallel()
	state, err := ApplyGraphPatch(DefaultGraphState(), GraphPatch{Operation: "replace_equations", Equations: []EquationInput{{Expression: "x"}, {Expression: "x^2"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Equations) != 2 {
		t.Fatalf("got %d equations", len(state.Equations))
	}
	id := state.Equations[0].ID
	hidden := false
	state, err = ApplyGraphPatch(state, GraphPatch{Operation: "update_equation", EquationID: id, Updates: &EquationUpdate{Visible: &hidden}})
	if err != nil {
		t.Fatal(err)
	}
	if state.Equations[0].Visible {
		t.Fatal("equation should be hidden")
	}
	state, err = ApplyGraphPatch(state, GraphPatch{Operation: "remove_equation", EquationID: id})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Equations) != 1 {
		t.Fatalf("got %d equations after removal", len(state.Equations))
	}
}
