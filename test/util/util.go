package util

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func GenerateRandomAlphanum(t *testing.T, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			t.Fatalf("error generating random number: %v", err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
