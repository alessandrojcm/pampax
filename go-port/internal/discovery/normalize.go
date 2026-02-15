package discovery

import (
	"fmt"
	"path/filepath"
	"strings"
)

func normalizeRelativePath(rootPath string, fullPath string) (string, error) {
	relativePath, err := filepath.Rel(rootPath, fullPath)
	if err != nil {
		return "", fmt.Errorf("compute relative path: %w", err)
	}

	return normalizeFromRelative(relativePath), nil
}

func normalizeFromRelative(relativePath string) string {
	normalized := strings.ReplaceAll(relativePath, "\\", "/")
	normalized = filepath.ToSlash(normalized)
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	return normalized
}
