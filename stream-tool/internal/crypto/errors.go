package crypto

import (
	"errors"
)

var ErrInvalidHMAC = errors.New("hmac verification failed")

var ErrDecryptFailed = errors.New("decryption failed: data may be corrupted or password is wrong")

var ErrFileTooShort = errors.New("encrypted file too short: may not be a valid encrypted file")

var ErrHashMistmatch = errors.New("FAIL: hash mismatch")

var ErrHMACMistmatch = errors.New("FAIL: hmac mismatch")
