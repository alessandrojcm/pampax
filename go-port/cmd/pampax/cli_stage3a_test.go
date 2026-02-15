package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommandHasPersistentFlags(t *testing.T) {
	cmd := NewRootCommand()

	for _, flagName := range []string{"pretty", "config", "verbose"} {
		flag := cmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Fatalf("missing persistent flag %q", flagName)
		}
	}
}

func TestIndexCommandRunsWithExpectedFlags(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "main.ts"), []byte("export const ok = true\n"), 0o644); err != nil {
		t.Fatalf("seed temp project: %v", err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"index", projectDir, "--provider", "openai", "--encrypt", "off"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute index command: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "index scaffold") {
		t.Fatalf("expected scaffold output, got %q", output)
	}
	if !strings.Contains(output, `"provider":"openai"`) {
		t.Fatalf("expected provider output, got %q", output)
	}
	if !strings.Contains(output, `"encrypt":"off"`) {
		t.Fatalf("expected encrypt output, got %q", output)
	}
}

func TestSearchCommandRequiresQuery(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"search"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected search command to fail without query")
	}
}

func TestSearchCommandSupportsPathAndTopAlias(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"search", "auth flow", "./repo", "--top", "5", "--provider", "openai"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute search command: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "search scaffold") {
		t.Fatalf("expected scaffold output, got %q", output)
	}
	if !strings.Contains(output, `"path":"./repo"`) {
		t.Fatalf("expected path output, got %q", output)
	}
	if !strings.Contains(output, `"limit":5`) {
		t.Fatalf("expected limit output, got %q", output)
	}
	if !strings.Contains(output, `"result_count":5`) {
		t.Fatalf("expected result count output, got %q", output)
	}
}

func TestInfoCommandRuns(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"info"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute info command: %v", err)
	}

	if !strings.Contains(out.String(), `"message":"info scaffold"`) {
		t.Fatalf("expected scaffold output, got %q", out.String())
	}
}
