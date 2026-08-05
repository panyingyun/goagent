package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	host := flag.String("host", "example.com", "允许访问的精确主机名")
	command := flag.String("command", "go-version", "固定命令名称")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: go run ./ch09-security -host example.com https://example.com")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	policy := URLPolicy{AllowedHosts: map[string]bool{*host: true}}
	validated, err := policy.Validate(ctx, flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "URL rejected:", err)
		os.Exit(1)
	}
	fmt.Println("URL accepted:", validated.String())
	cmd, err := FixedCommand(*command)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(string(output))
}

