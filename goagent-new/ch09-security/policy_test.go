package main

import (
	"context"
	"net"
	"testing"
)

type fakeResolver map[string][]net.IP

func (f fakeResolver) LookupIP(_ context.Context, _, host string) ([]net.IP, error) {
	return f[host], nil
}

func TestURLPolicyRejectsPrivateAndUnlistedHosts(t *testing.T) {
	policy := URLPolicy{
		AllowedHosts: map[string]bool{"public.example": true, "internal.example": true},
		Resolver: fakeResolver{
			"public.example":   {net.ParseIP("93.184.216.34")},
			"internal.example": {net.ParseIP("127.0.0.1")},
		},
	}
	if _, err := policy.Validate(context.Background(), "https://public.example/data"); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		"http://public.example/data",
		"https://evil.example/data",
		"https://internal.example/admin",
		"https://user:pass@public.example/data",
	} {
		if _, err := policy.Validate(context.Background(), raw); err == nil {
			t.Fatalf("expected rejection: %s", raw)
		}
	}
}

func TestFixedCommandDoesNotAcceptShellText(t *testing.T) {
	if _, err := FixedCommand("go-version & whoami"); err == nil {
		t.Fatal("shell text should not be accepted")
	}
	cmd, err := FixedCommand("go-version")
	if err != nil || cmd.Path == "" || len(cmd.Args) != 2 {
		t.Fatalf("unexpected command: %+v %v", cmd, err)
	}
}

