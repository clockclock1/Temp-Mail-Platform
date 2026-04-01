package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func SplitEmailAddress(address string) (string, string, error) {
	addr := strings.ToLower(strings.TrimSpace(address))
	parts := strings.Split(addr, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid address")
	}
	return parts[0], parts[1], nil
}

func RandomAlphaNum(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	if n <= 0 {
		return ""
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
