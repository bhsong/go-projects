package crypto

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("crypto.HashFile: open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()

	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("crypto.HashFile: read file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func VerifyFile(path, expected string) (bool, error) {
	actual, err := HashFile(path)
	if err != nil {
		return false, fmt.Errorf("crypto.VerifyFile: %w", err)
	}

	actualBytes, err := hex.DecodeString(actual)
	if err != nil {
		return false, fmt.Errorf("crypto.VerifyFile: decode actual: %w", err)
	}

	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false, fmt.Errorf("crypto.VerifyFile: decode expectged: %w", err)
	}

	return subtle.ConstantTimeCompare(actualBytes, expectedBytes) == 1, nil
}
