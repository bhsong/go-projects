package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func GenerateHMAC(path, key string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("crypto.GenerateHMAC: open file: %w", err)
	}
	defer f.Close()

	h := hmac.New(sha256.New, []byte(key))

	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("crypto.GenerateHMAC: read file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func VerifyHMAC(path, key, expected string) (bool, error) {
	actual, err := GenerateHMAC(path, key)
	if err != nil {
		return false, fmt.Errorf("crypto.VerifyHMAC: %w", err)
	}

	actualBytes, err := hex.DecodeString(actual)
	if err != nil {
		return false, fmt.Errorf("crypto.VerifyHMAC: decode actual: %w", err)
	}

	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false, fmt.Errorf("crypto.VerifyHMAC: decode expected: %w", err)
	}

	return hmac.Equal(actualBytes, expectedBytes), nil
}
