package codemap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func MarshalCodemap(codemap *OrderedMap) ([]byte, error) {
	raw, err := json.Marshal(codemap)
	if err != nil {
		return nil, fmt.Errorf("marshal codemap: %w", err)
	}

	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		return nil, fmt.Errorf("format codemap json: %w", err)
	}

	out.WriteByte('\n')
	return out.Bytes(), nil
}

func WriteCodemap(path string, codemap *OrderedMap) error {
	payload, err := MarshalCodemap(codemap)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create codemap directory: %w", err)
	}

	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write codemap file: %w", err)
	}

	return nil
}
