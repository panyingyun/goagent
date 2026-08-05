package memory

import (
	"path/filepath"
	"testing"
)

func TestStoreLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set("drink", "coffee", "user explicitly asked to remember"); err != nil {
		t.Fatal(err)
	}
	loaded, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := loaded.Get("drink"); !ok || got.Value != "coffee" || got.Source == "" {
		t.Fatalf("unexpected preference: %+v, %v", got, ok)
	}
	if err := loaded.DeleteAll(); err != nil {
		t.Fatal(err)
	}
	if len(loaded.List()) != 0 {
		t.Fatal("preferences should be deletable")
	}
}

