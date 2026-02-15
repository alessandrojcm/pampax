package chunks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// WriteChunk writes a chunk to disk as {sha}.gz or {sha}.gz.enc using atomic rename.
func WriteChunk(chunkDir, sha, code string, encrypted bool, masterKey []byte) error {
	if sha == "" {
		return errors.New("sha is required")
	}

	if err := os.MkdirAll(chunkDir, 0o755); err != nil {
		return fmt.Errorf("create chunk directory: %w", err)
	}

	compressed, err := Compress([]byte(code))
	if err != nil {
		return fmt.Errorf("compress chunk: %w", err)
	}

	plainPath := filepath.Join(chunkDir, sha+".gz")
	encryptedPath := filepath.Join(chunkDir, sha+".gz.enc")

	if encrypted {
		payload, err := Encrypt(compressed, masterKey)
		if err != nil {
			return fmt.Errorf("encrypt chunk: %w", err)
		}

		if err := writeFileAtomically(encryptedPath, payload, 0o644); err != nil {
			return fmt.Errorf("write encrypted chunk: %w", err)
		}

		if err := removeIfExists(plainPath); err != nil {
			return fmt.Errorf("remove plaintext chunk: %w", err)
		}

		return nil
	}

	if err := writeFileAtomically(plainPath, compressed, 0o644); err != nil {
		return fmt.Errorf("write chunk: %w", err)
	}

	if err := removeIfExists(encryptedPath); err != nil {
		return fmt.Errorf("remove encrypted chunk: %w", err)
	}

	return nil
}

// ReadChunk loads a chunk from disk, preferring encrypted chunks when present.
func ReadChunk(chunkDir, sha string, encrypted bool, masterKey []byte) (string, error) {
	if sha == "" {
		return "", errors.New("sha is required")
	}

	plainPath := filepath.Join(chunkDir, sha+".gz")
	encryptedPath := filepath.Join(chunkDir, sha+".gz.enc")

	payloadPath := plainPath
	needsDecrypt := false

	if _, err := os.Stat(encryptedPath); err == nil {
		payloadPath = encryptedPath
		needsDecrypt = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat encrypted chunk: %w", err)
	}

	if !needsDecrypt {
		if _, err := os.Stat(plainPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("chunk %s not found", sha)
			}
			return "", fmt.Errorf("stat chunk: %w", err)
		}
	}

	raw, err := os.ReadFile(payloadPath)
	if err != nil {
		return "", fmt.Errorf("read chunk %s: %w", sha, err)
	}

	if needsDecrypt {
		if len(masterKey) == 0 {
			return "", fmt.Errorf("chunk %s is encrypted and no key was provided", sha)
		}

		raw, err = Decrypt(raw, masterKey)
		if err != nil {
			return "", fmt.Errorf("decrypt chunk %s: %w", sha, err)
		}
	} else if encrypted {
		return "", fmt.Errorf("chunk %s is not encrypted", sha)
	}

	decompressed, err := Decompress(raw)
	if err != nil {
		return "", fmt.Errorf("decompress chunk %s: %w", sha, err)
	}

	return string(decompressed), nil
}

// RemoveChunk deletes both plaintext and encrypted variants for a chunk SHA.
func RemoveChunk(chunkDir, sha string) error {
	if sha == "" {
		return errors.New("sha is required")
	}

	plainPath := filepath.Join(chunkDir, sha+".gz")
	encryptedPath := filepath.Join(chunkDir, sha+".gz.enc")

	if err := removeIfExists(plainPath); err != nil {
		return fmt.Errorf("remove plaintext chunk: %w", err)
	}

	if err := removeIfExists(encryptedPath); err != nil {
		return fmt.Errorf("remove encrypted chunk: %w", err)
	}

	return nil
}

func writeFileAtomically(path string, data []byte, perm os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpPath := file.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := file.Chmod(perm); err != nil {
		_ = file.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	cleanup = false
	return nil
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
