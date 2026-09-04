package auth

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	original := []byte("secret-token-data-12345")

	encrypted, err := Encrypt(original)
	if err != nil {
		t.Fatalf("failed to encrypt data: %v", err)
	}

	if bytes.Equal(original, encrypted) {
		t.Error("expected ciphertext to differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt data: %v", err)
	}

	if !bytes.Equal(original, decrypted) {
		t.Errorf("decrypted data %q does not match original %q", string(decrypted), string(original))
	}
}

func TestDecryptInvalidData(t *testing.T) {
	invalidData := []byte("too-short")
	_, err := Decrypt(invalidData)
	if err == nil {
		t.Error("expected error when decrypting invalid short data, got nil")
	}
}
