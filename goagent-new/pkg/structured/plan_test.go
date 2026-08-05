package structured

import "testing"

func TestDecodeTaskPlan(t *testing.T) {
	valid := `{"title":"demo","steps":[{"id":"s1","task":"do it"}]}`
	if _, err := DecodeTaskPlan(valid); err != nil {
		t.Fatal(err)
	}
	invalid := []string{
		`{"title":"","steps":[{"id":"s1","task":"x"}]}`,
		`{"title":"demo","steps":[]}`,
		`{"title":"demo","steps":[{"id":"s1","task":"x"},{"id":"s1","task":"y"}]}`,
		`{"title":"demo","unknown":true,"steps":[{"id":"s1","task":"x"}]}`,
	}
	for _, input := range invalid {
		if _, err := DecodeTaskPlan(input); err == nil {
			t.Fatalf("expected rejection: %s", input)
		}
	}
}

