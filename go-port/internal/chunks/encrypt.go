package chunks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

var (
	magicHeader = []byte("PAMPAE1")
	hkdfInfo    = []byte("pampa-chunk-v1")
)

const (
	saltLength = 16
	ivLength   = 12
	tagLength  = 16
)

// DeriveChunkKey derives a 32-byte AES key from a 32-byte master key and a 16-byte salt.
func DeriveChunkKey(masterKey, salt []byte) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("invalid master key length: got %d, want 32", len(masterKey))
	}

	if len(salt) != saltLength {
		return nil, fmt.Errorf("invalid salt length: got %d, want %d", len(salt), saltLength)
	}

	reader := hkdf.New(sha256.New, masterKey, salt, hkdfInfo)
	derived := make([]byte, 32)
	if _, err := io.ReadFull(reader, derived); err != nil {
		return nil, fmt.Errorf("derive hkdf key: %w", err)
	}

	return derived, nil
}

// Encrypt wraps gzipped bytes into the PAMPAE1 encrypted chunk format.
func Encrypt(gzipped, masterKey []byte) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("invalid master key length: got %d, want 32", len(masterKey))
	}

	salt := make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate iv: %w", err)
	}

	derivedKey, err := DeriveChunkKey(masterKey, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm cipher: %w", err)
	}

	sealed := gcm.Seal(nil, iv, gzipped, nil)
	payload := make([]byte, 0, len(magicHeader)+saltLength+ivLength+len(sealed))
	payload = append(payload, magicHeader...)
	payload = append(payload, salt...)
	payload = append(payload, iv...)
	payload = append(payload, sealed...)

	return payload, nil
}

// Decrypt unwraps a PAMPAE1 encrypted payload and returns the gzipped bytes.
func Decrypt(payload, masterKey []byte) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("invalid master key length: got %d, want 32", len(masterKey))
	}

	minLength := len(magicHeader) + saltLength + ivLength + tagLength + 1
	if len(payload) < minLength {
		return nil, errors.New("encrypted chunk payload is truncated")
	}

	header := payload[:len(magicHeader)]
	if string(header) != string(magicHeader) {
		return nil, errors.New("encrypted chunk payload has an unknown header")
	}

	saltStart := len(magicHeader)
	ivStart := saltStart + saltLength
	cipherStart := ivStart + ivLength

	salt := payload[saltStart:ivStart]
	iv := payload[ivStart:cipherStart]
	sealed := payload[cipherStart:]

	derivedKey, err := DeriveChunkKey(masterKey, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm cipher: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, sealed, nil)
	if err != nil {
		return nil, errors.New("authentication failed")
	}

	return plaintext, nil
}
