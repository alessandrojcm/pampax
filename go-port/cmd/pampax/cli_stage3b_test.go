package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestInfoCommandPrettyFlagUsesConsoleOutput(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--pretty", "info"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute info command: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "info scaffold") {
		t.Fatalf("expected info scaffold output, got %q", output)
	}
	if strings.Contains(output, `"message":"info scaffold"`) {
		t.Fatalf("expected pretty (console) log output, got %q", output)
	}
}
