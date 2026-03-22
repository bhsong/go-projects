package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"os"

	"golang.org/x/crypto/scrypt"
)

func EncryptFile(srcPath, dstPath, password string) error {
	plaintext, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("crypto.EncryptFile: read source: %w", err)
	}

	salt := make([]byte, 32)
	_, err = rand.Read(salt)
	if err != nil {
		return fmt.Errorf("crypto.EncryptFile: generate salt: %w", err)
	}

	key, err := deriveKey(password, salt)
	if err != nil {
		return fmt.Errorf("crypto.EncryptFile: derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("crypto.EncryptFile: create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("crypto.EncryptFile: create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return fmt.Errorf("crypto.EncryptFile: generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	out := append(salt, ciphertext...)

	err = os.WriteFile(dstPath, out, 0600)
	if err != nil {
		return fmt.Errorf("crypto.EncryptFile: write output: %w", err)
	}

	return nil
}

func DecryptFile(srcPath, dstPath, password string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("crypto.DecryptFile: read source: %w", err)
	}

	if len(data) < 44 {
		return fmt.Errorf("crypto.DecryptFile: %w", ErrFileTooShort)
	}

	salt := data[:32]
	rest := data[32:]

	key, err := deriveKey(password, salt)
	if err != nil {
		return fmt.Errorf("crypto.DecryptFile: derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("crypto.DecryptFile: create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("crypto.DecryptFile: create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()

	if len(rest) < nonceSize {
		return fmt.Errorf("crypto.DecryptFile: %w", ErrFileTooShort)
	}

	nonce := rest[:nonceSize]
	ciphertext := rest[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("crypto.DecryptFile: %w", ErrDecryptFailed)
	}
	err = os.WriteFile(dstPath, plaintext, 0600)
	if err != nil {
		return fmt.Errorf("crypto.DecryptFile: write output: %w", err)
	}

	return nil
}

func deriveKey(password string, salt []byte) ([]byte, error) {
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("deriveKey: scrypt failed: %w", err)
	}

	return key, nil
}
