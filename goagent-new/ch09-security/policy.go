package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"strings"
)

type Resolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

type URLPolicy struct {
	AllowedHosts map[string]bool
	Resolver     Resolver
}

// Validate 检查 scheme、凭据、主机 allowlist 和解析后的全部 IP。
// 真正发送请求时还必须固定已验证 IP，并对每次重定向重新执行策略，防 DNS rebinding。
func (p URLPolicy) Validate(ctx context.Context, raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Hostname() == "" {
		return nil, errors.New("only absolute https URLs are allowed")
	}
	if parsed.User != nil {
		return nil, errors.New("URL credentials are not allowed")
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	if !p.AllowedHosts[host] {
		return nil, fmt.Errorf("host %q is not in the allowlist", host)
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return nil, errors.New("only port 443 is allowed")
	}
	resolver := p.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addresses, err := resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve host: %w", err)
	}
	if len(addresses) == 0 {
		return nil, errors.New("host resolved to no addresses")
	}
	for _, address := range addresses {
		if !publicIP(address) {
			return nil, fmt.Errorf("host resolves to non-public address %s", address)
		}
	}
	return parsed, nil
}

func publicIP(ip net.IP) bool {
	return ip != nil && !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() && !ip.IsUnspecified() && !ip.IsMulticast()
}

// FixedCommand 把用户选择映射为固定 executable + argv，不经过 cmd /C 或 sh -c。
func FixedCommand(name string) (*exec.Cmd, error) {
	switch name {
	case "go-version":
		return exec.Command("go", "version"), nil
	case "git-version":
		return exec.Command("git", "--version"), nil
	default:
		return nil, fmt.Errorf("command %q is not allowed", name)
	}
}

