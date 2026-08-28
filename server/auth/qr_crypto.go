package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var qrSecretKey = sha256.Sum256([]byte("PCRemote-v3-Secure-QR-Envelope-Protocol-2026"))

// QRProtocolPrefix is the header prepended to all encrypted QR payloads.
const QRProtocolPrefix = "PCR3:"

// EncryptQRPayload encrypts raw JSON bytes into an encrypted string "PCR3:<base64>".
// This prevents generic phone cameras/lens scanners from reading server details in plaintext.
func EncryptQRPayload(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(qrSecretKey[:])
	if err != nil {
		return "", err
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	combined := append(iv, ciphertext...)
	encoded := base64.StdEncoding.EncodeToString(combined)

	return QRProtocolPrefix + encoded, nil
}

// DecryptQRPayload decrypts a "PCR3:<base64>" payload back to JSON bytes.
func DecryptQRPayload(input string) ([]byte, error) {
	if len(input) < len(QRProtocolPrefix) || input[:len(QRProtocolPrefix)] != QRProtocolPrefix {
		return nil, errors.New("invalid QR protocol prefix")
	}

	encoded := input[len(QRProtocolPrefix):]
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 payload: %w", err)
	}

	if len(combined) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := combined[:aes.BlockSize]
	ciphertext := combined[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}

	block, err := aes.NewCipher(qrSecretKey[:])
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	unpadded, err := pkcs7Unpad(decrypted)
	if err != nil {
		return nil, err
	}

	return unpadded, nil
}

func pkcs7Pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func pkcs7Unpad(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, errors.New("empty decrypted data")
	}
	unpadding := int(src[length-1])
	if unpadding > length || unpadding == 0 {
		return nil, errors.New("invalid PKCS#7 padding")
	}
	for _, b := range src[length-unpadding:] {
		if int(b) != unpadding {
			return nil, errors.New("invalid PKCS#7 padding byte")
		}
	}
	return src[:(length - unpadding)], nil
}
