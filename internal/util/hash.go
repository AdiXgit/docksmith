package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
