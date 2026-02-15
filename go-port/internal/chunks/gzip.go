package chunks

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// Compress applies gzip compression to data using default compression.
func Compress(data []byte) ([]byte, error) {
	var out bytes.Buffer
	writer := gzip.NewWriter(&out)

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("write gzip payload: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close gzip writer: %w", err)
	}

	return out.Bytes(), nil
}

// Decompress expands a gzip payload back into raw bytes.
func Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open gzip reader: %w", err)
	}
	defer reader.Close()

	out, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read gzip payload: %w", err)
	}

	return out, nil
}
