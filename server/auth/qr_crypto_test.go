package auth

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncryptDecryptQRPayload(t *testing.T) {
	original := []byte(`{"host":"192.168.0.100","port":"8000","pair_token":"abcdef123456"}`)

	encrypted, err := EncryptQRPayload(original)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if !strings.HasPrefix(encrypted, "PCR3:") {
		t.Errorf("expected PCR3: prefix, got: %s", encrypted)
	}

	if strings.Contains(encrypted, "192.168.0.100") || strings.Contains(encrypted, "pair_token") {
		t.Errorf("encrypted string leaked plaintext data: %s", encrypted)
	}

	decrypted, err := DecryptQRPayload(encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, original) {
		t.Errorf("expected %s, got %s", string(original), string(decrypted))
	}
}

func TestDecryptQRPayload_InvalidPrefix(t *testing.T) {
	_, err := DecryptQRPayload("INVALID:somebase64string")
	if err == nil {
		t.Errorf("expected error for invalid prefix, got nil")
	}
}

func TestDecryptQRPayload_CorruptedData(t *testing.T) {
	_, err := DecryptQRPayload("PCR3:corrupted===data")
	if err == nil {
		t.Errorf("expected error for corrupted base64 data, got nil")
	}
}
