package unit

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alessandrojcm/pampax-go/internal/chunks"
)

func TestComputeSHAMatchesNodeIncludingCRLFAndBOM(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple", input: "hello", expected: "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{name: "crlf", input: "hello\r\nworld", expected: "d07cff009c449bfdf131d865e1dc4413256e5f52"},
		{name: "bom", input: "\ufeffbom\nline", expected: "84ab5499e7d6d83d8839cc1749dce7ac8d85ae9e"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chunks.ComputeSHA(tt.input)
			if got != tt.expected {
				t.Fatalf("ComputeSHA() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestGzipRoundtripPreservesExactBytes(t *testing.T) {
	original := []byte("\xef\xbb\xbfline1\r\nline2\nconst 日本語 = '値';")

	compressed, err := chunks.Compress(original)
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	restored, err := chunks.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress() error = %v", err)
	}

	if !bytes.Equal(restored, original) {
		t.Fatalf("decompressed bytes differ\n got: %q\nwant: %q", restored, original)
	}
}

func TestDeriveChunkKeyMatchesNodeVector(t *testing.T) {
	masterKey, err := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	if err != nil {
		t.Fatalf("DecodeString master key error = %v", err)
	}

	salt, err := hex.DecodeString("f0e0d0c0b0a090807060504030201000")
	if err != nil {
		t.Fatalf("DecodeString salt error = %v", err)
	}

	derived, err := chunks.DeriveChunkKey(masterKey, salt)
	if err != nil {
		t.Fatalf("DeriveChunkKey() error = %v", err)
	}

	got := hex.EncodeToString(derived)
	expected := "6eed612f20f4bcb23e0f5f3023a337c73647da8e626041dea455feafe5ba3b99"
	if got != expected {
		t.Fatalf("DeriveChunkKey() = %s, want %s", got, expected)
	}
}

func TestEncryptDecryptRoundtripAndHeader(t *testing.T) {
	masterKey := bytes.Repeat([]byte{7}, 32)
	gzipped, err := chunks.Compress([]byte("console.log('secret');\r\n"))
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	payload, err := chunks.Encrypt(gzipped, masterKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if !bytes.HasPrefix(payload, []byte("PAMPAE1")) {
		t.Fatalf("encrypted payload missing PAMPAE1 header: %x", payload)
	}

	decrypted, err := chunks.Decrypt(payload, masterKey)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, gzipped) {
		t.Fatalf("Decrypt() mismatch")
	}
}

func TestDecryptFailsWhenTagIsTampered(t *testing.T) {
	masterKey := bytes.Repeat([]byte{3}, 32)
	gzipped, err := chunks.Compress([]byte("tamper me"))
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	payload, err := chunks.Encrypt(gzipped, masterKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	payload[len(payload)-1] ^= 0xff
	_, err = chunks.Decrypt(payload, masterKey)
	if err == nil {
		t.Fatal("expected Decrypt() to fail for tampered payload")
	}

	if !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("expected authentication failure, got: %v", err)
	}
}

func TestWriteReadChunkRoundtripAndAtomicWrite(t *testing.T) {
	chunkDir := t.TempDir()
	sha := "5ea95a5a78779486d1fccdab927a7d64f5cf1599"
	code := "\ufefffunction foo() {\r\n  return 42;\r\n}"

	if err := chunks.WriteChunk(chunkDir, sha, code, false, nil); err != nil {
		t.Fatalf("WriteChunk() error = %v", err)
	}

	content, err := chunks.ReadChunk(chunkDir, sha, false, nil)
	if err != nil {
		t.Fatalf("ReadChunk() error = %v", err)
	}

	if content != code {
		t.Fatalf("ReadChunk() mismatch\n got: %q\nwant: %q", content, code)
	}

	entries, err := os.ReadDir(chunkDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	targetFile := sha + ".gz"
	foundTarget := false
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, ".tmp-") {
			t.Fatalf("unexpected temp file left behind: %s", name)
		}
		if name == targetFile {
			foundTarget = true
		}
	}

	if !foundTarget {
		t.Fatalf("expected %s to exist", targetFile)
	}
}

func TestWriteReadEncryptedChunkRoundtripAndRemove(t *testing.T) {
	chunkDir := t.TempDir()
	masterKey := bytes.Repeat([]byte{9}, 32)
	sha := chunks.ComputeSHA("export const token = 'abc';")
	code := "export const token = 'abc';\n"

	if err := chunks.WriteChunk(chunkDir, sha, code, true, masterKey); err != nil {
		t.Fatalf("WriteChunk() encrypted error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(chunkDir, sha+".gz.enc")); err != nil {
		t.Fatalf("expected encrypted file to exist: %v", err)
	}

	if _, err := os.Stat(filepath.Join(chunkDir, sha+".gz")); !os.IsNotExist(err) {
		t.Fatalf("expected plaintext file to not exist, got err: %v", err)
	}

	content, err := chunks.ReadChunk(chunkDir, sha, true, masterKey)
	if err != nil {
		t.Fatalf("ReadChunk() encrypted error = %v", err)
	}

	if content != code {
		t.Fatalf("encrypted roundtrip mismatch")
	}

	if err := chunks.RemoveChunk(chunkDir, sha); err != nil {
		t.Fatalf("RemoveChunk() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(chunkDir, sha+".gz.enc")); !os.IsNotExist(err) {
		t.Fatalf("expected encrypted file removed, got err: %v", err)
	}
}
