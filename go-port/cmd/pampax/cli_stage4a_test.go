package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestIndexCommandResolvesProviderDetails(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"index", "./project", "--provider", "openai"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute index command: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"provider_name":"OpenAI"`) {
		t.Fatalf("expected resolved provider name, got %q", output)
	}
	if !strings.Contains(output, `"provider_dimensions":1536`) {
		t.Fatalf("expected resolved provider dimensions from config, got %q", output)
	}
}

func TestSearchCommandRejectsUnknownProvider(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"search", "query", "--provider", "unknown"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected command error")
	}
	if !strings.Contains(err.Error(), "resolve embedding provider") {
		t.Fatalf("expected provider resolution error, got %v", err)
	}
}
