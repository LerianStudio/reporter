package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

func RandString(n int) string {
	if n <= 0 {
		n = 8
	}

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	out := make([]byte, n)
	tmp := make([]byte, n)

	_, _ = rand.Read(tmp)
	for i, b := range tmp {
		out[i] = letters[int(b)%len(letters)]
	}

	return string(out)
}

func RandHex(n int) string {
	if n <= 0 {
		n = 8
	}

	b := make([]byte, n)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
