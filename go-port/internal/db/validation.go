package db

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type ValidatedJSONFields struct {
	PampaTags     *string
	VariablesUsed *string
	ContextInfo   *string
}

var jsonFieldValidator = newJSONFieldValidator()

// ValidateChunkJSONFields validates optional JSON metadata fields and skips invalid values.
func ValidateChunkJSONFields(pampaTags, variablesUsed, contextInfo *string) ValidatedJSONFields {
	return ValidatedJSONFields{
		PampaTags:     validateJSONField("pampa_tags", pampaTags, "json_array"),
		VariablesUsed: validateJSONField("variables_used", variablesUsed, "json_array"),
		ContextInfo:   validateJSONField("context_info", contextInfo, "json_object"),
	}
}

func newJSONFieldValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("json_array", func(fl validator.FieldLevel) bool {
		return validateJSONKind(fl.Field().String(), "array") == nil
	})
	v.RegisterValidation("json_object", func(fl validator.FieldLevel) bool {
		return validateJSONKind(fl.Field().String(), "object") == nil
	})
	return v
}

func validateJSONField(fieldName string, value *string, validationTag string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		warnInvalidJSONField(fieldName, fmt.Errorf("empty value"))
		return nil
	}

	if err := jsonFieldValidator.Var(trimmed, validationTag); err != nil {
		warnInvalidJSONField(fieldName, err)
		return nil
	}

	validated := trimmed
	return &validated
}

func validateJSONKind(raw string, expectedKind string) error {
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return err
	}

	switch expectedKind {
	case "array":
		if _, ok := payload.([]any); !ok {
			return fmt.Errorf("expected JSON array")
		}
	case "object":
		if _, ok := payload.(map[string]any); !ok {
			return fmt.Errorf("expected JSON object")
		}
	default:
		return fmt.Errorf("unsupported expected JSON kind: %s", expectedKind)
	}

	return nil
}

func warnInvalidJSONField(fieldName string, err error) {
	log.Warn().Str("field", fieldName).Err(err).Msg("invalid JSON field, skipping")
}
