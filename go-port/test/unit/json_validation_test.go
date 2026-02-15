package unit

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alessandrojcm/pampax-go/internal/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestValidateChunkJSONFieldsAcceptsValidPayloads(t *testing.T) {
	pampaTags := ` ["auth","login"] `
	variablesUsed := `[{"type":"string","name":"token"}]`
	contextInfo := ` {"scope":"search"} `

	validated := db.ValidateChunkJSONFields(&pampaTags, &variablesUsed, &contextInfo)

	if validated.PampaTags == nil || *validated.PampaTags != `["auth","login"]` {
		t.Fatalf("expected trimmed valid pampa_tags, got %#v", validated.PampaTags)
	}
	if validated.VariablesUsed == nil || *validated.VariablesUsed != variablesUsed {
		t.Fatalf("expected valid variables_used, got %#v", validated.VariablesUsed)
	}
	if validated.ContextInfo == nil || *validated.ContextInfo != `{"scope":"search"}` {
		t.Fatalf("expected trimmed valid context_info, got %#v", validated.ContextInfo)
	}
}

func TestValidateChunkJSONFieldsSkipsInvalidPayloadsWithWarning(t *testing.T) {
	originalLogger := log.Logger
	var logBuffer bytes.Buffer
	log.Logger = zerolog.New(&logBuffer)
	t.Cleanup(func() {
		log.Logger = originalLogger
	})

	invalidTags := `{"not":"array"}`
	invalidVariables := `oops`
	emptyContext := "   "

	validated := db.ValidateChunkJSONFields(&invalidTags, &invalidVariables, &emptyContext)

	if validated.PampaTags != nil {
		t.Fatal("expected invalid pampa_tags to be skipped")
	}
	if validated.VariablesUsed != nil {
		t.Fatal("expected invalid variables_used to be skipped")
	}
	if validated.ContextInfo != nil {
		t.Fatal("expected empty context_info to be skipped")
	}

	logs := logBuffer.String()
	for _, field := range []string{"pampa_tags", "variables_used", "context_info"} {
		if !strings.Contains(logs, `"field":"`+field+`"`) {
			t.Fatalf("expected warning for %s, got logs: %s", field, logs)
		}
	}
	if !strings.Contains(logs, "invalid JSON field, skipping") {
		t.Fatalf("expected invalid JSON warning message, got logs: %s", logs)
	}
}
