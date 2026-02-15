package utils

import (
	"bytes"
	"strings"
	"testing"
)

func TestSetupLoggerJSONByDefault(t *testing.T) {
	var out bytes.Buffer
	logger := SetupLogger(LoggingOptions{Writer: &out})
	logger.Info().Msg("json check")

	logged := out.String()
	if !strings.Contains(logged, "\"message\":\"json check\"") {
		t.Fatalf("expected JSON log output, got %q", logged)
	}
}

func TestSetupLoggerPrettyOutput(t *testing.T) {
	var out bytes.Buffer
	logger := SetupLogger(LoggingOptions{Writer: &out, Pretty: true})
	logger.Info().Msg("pretty check")

	logged := out.String()
	if !strings.Contains(logged, "pretty check") {
		t.Fatalf("expected pretty message output, got %q", logged)
	}
	if strings.Contains(logged, "\"message\":\"pretty check\"") {
		t.Fatalf("expected non-JSON pretty output, got %q", logged)
	}
}

func TestSetupLoggerVerboseEnablesDebug(t *testing.T) {
	var out bytes.Buffer
	logger := SetupLogger(LoggingOptions{Writer: &out, Verbose: true})
	logger.Debug().Msg("debug check")

	logged := out.String()
	if !strings.Contains(logged, "\"level\":\"debug\"") {
		t.Fatalf("expected debug level log, got %q", logged)
	}
}
