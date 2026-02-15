package chunks

import (
	"crypto/sha1"
	"encoding/hex"
)

// ComputeSHA computes SHA-1 for the raw UTF-8 bytes of code.
func ComputeSHA(code string) string {
	sum := sha1.Sum([]byte(code))
	return hex.EncodeToString(sum[:])
}
