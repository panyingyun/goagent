package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"goagent-new/pkg/memory"
)

func execute(store *memory.Store, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("commands: set <key> <value> | get <key> | list | delete <key> | clear")
	}
	switch args[0] {
	case "set":
		if len(args) != 3 {
			return errors.New("usage: set <key> <value>")
		}
		return store.Set(args[1], args[2], "user explicitly requested persistence")
	case "get":
		if len(args) != 2 {
			return errors.New("usage: get <key>")
		}
		preference, ok := store.Get(args[1])
		if !ok {
			return errors.New("preference not found")
		}
		return json.NewEncoder(out).Encode(preference)
	case "list":
		return json.NewEncoder(out).Encode(store.List())
	case "delete":
		if len(args) != 2 {
			return errors.New("usage: delete <key>")
		}
		return store.Delete(args[1])
	case "clear":
		return store.DeleteAll()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func main() {
	path := filepath.Join(envOr("DATA_DIR", "data"), "preferences.json")
	store, err := memory.Open(path)
	if err != nil {
		fatal(err)
	}
	if err := execute(store, os.Args[1:], os.Stdout); err != nil {
		fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

