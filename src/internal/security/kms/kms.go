package kms

import (
	"context"
	"encoding/base64"
	"errors"
)

var (
	ErrKMSUnavailable = errors.New("KMS unavailable")
	ErrUnauthorized   = errors.New("unauthorized")
)

// Encryptor interface para cifrado de sobre (envelope).
// HU-028: Client-Side Encryption con DEK.
type Encryptor interface {
	Encrypt(ctx context.Context, plaintext string) (string, error)
	Decrypt(ctx context.Context, ciphertext string) (string, error)
}

// Persistor interface para persistencia con fallback.
// HU-028 AC5: Fallback a archivo si BD falla.
type Persistor interface {
	PersistWithFallback(ctx context.Context, ciphertext string) error
}

// mockEncryptor implementa Encryptor con mock (no criptografía real).
// ponytail: mock en lugar de AES-GCM. Implementación real en EP-009.
type mockEncryptor struct{}

func NewMockEncryptor() Encryptor {
	return &mockEncryptor{}
}

func (m *mockEncryptor) Encrypt(ctx context.Context, plaintext string) (string, error) {
	// Cifrado mock: base64 encoding
	return base64.StdEncoding.EncodeToString([]byte(plaintext)), nil
}

func (m *mockEncryptor) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	// Descifrado mock: base64 decoding
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// failingEncryptor simula KMS inaccesible (AC3).
type failingEncryptor struct{}

func NewFailingEncryptor() Encryptor {
	return &failingEncryptor{}
}

func (f *failingEncryptor) Encrypt(ctx context.Context, plaintext string) (string, error) {
	return "", ErrKMSUnavailable
}

func (f *failingEncryptor) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	return "", ErrKMSUnavailable
}

// unauthorizedEncryptor simula falta de permisos (AC4).
type unauthorizedEncryptor struct{}

func NewUnauthorizedEncryptor() Encryptor {
	return &unauthorizedEncryptor{}
}

func (u *unauthorizedEncryptor) Encrypt(ctx context.Context, plaintext string) (string, error) {
	return "", ErrUnauthorized
}

func (u *unauthorizedEncryptor) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	return "", ErrUnauthorized
}

// chanPersistor implementa Persistor con fallback (AC5).
type chanPersistor struct {
	fallbackChan chan<- string
}

func NewPersistor() Persistor {
	return &chanPersistor{
		fallbackChan: make(chan string, 100),
	}
}

func (c *chanPersistor) PersistWithFallback(ctx context.Context, ciphertext string) error {
	// En caso de fallo de BD, envia a fallback
	select {
	case c.fallbackChan <- ciphertext:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
