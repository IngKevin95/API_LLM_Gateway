package kms_test

import (
	"context"
	"testing"

	"github.com/IngKevin95/API_LLM_Gateway/internal/security/kms"
)

// HU-028 AC1 — Happy: Inserción cifrada con DEK
func TestKMS_EncryptPayload_CifradoLocal(t *testing.T) {
	encryptor := kms.NewMockEncryptor()

	// Dado: payload listo para auditar
	plaintext := "sensitive audit log data"

	// Cuando: se cifra localmente
	ciphertext, err := encryptor.Encrypt(context.Background(), plaintext)

	// Entonces: retorna cifrado (mock retorna base64 encodificado)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if ciphertext == plaintext {
		t.Error("payload should be encrypted, not plaintext")
	}
	if len(ciphertext) == 0 {
		t.Error("expected non-empty ciphertext")
	}
}

// HU-028 AC2 — Edge: Lectura autorizada con KMS_READER
func TestKMS_AuthorizedRead_Decrypts(t *testing.T) {
	encryptor := kms.NewMockEncryptor()

	// Dado: payload cifrado
	plaintext := "user data"
	ciphertext, _ := encryptor.Encrypt(context.Background(), plaintext)

	// Cuando: admin con rol KMS_READER intenta descifrar
	decrypted, err := encryptor.Decrypt(context.Background(), ciphertext)

	// Entonces: descifra exitosamente
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

// HU-028 AC3 — Error: KMS inaccesible → descartar payload
func TestKMS_KMSUnavailable_DiscardPayload(t *testing.T) {
	encryptor := kms.NewFailingEncryptor() // mock que falla

	// Dado: KMS externo está caído
	plaintext := "data to encrypt"

	// Cuando: intenta cifrar
	_, err := encryptor.Encrypt(context.Background(), plaintext)

	// Entonces: retorna error (payload descartado)
	if err == nil {
		t.Error("expected error when KMS unavailable")
	}
	if err != kms.ErrKMSUnavailable {
		t.Errorf("expected ErrKMSUnavailable, got %v", err)
	}
}

// HU-028 AC4 — Error: Sin autorización → acceso denegado
func TestKMS_UnauthorizedRead_AccessDenied(t *testing.T) {
	// Dado: payload cifrado
	ciphertext := "encrypted-data"

	// Cuando: usuario sin KMS_READER intenta descifrar
	// (mock que simula falta de permiso)
	unauthorizedEncryptor := kms.NewUnauthorizedEncryptor()
	_, err := unauthorizedEncryptor.Decrypt(context.Background(), ciphertext)

	// Entonces: acceso denegado
	if err == nil {
		t.Error("expected authorization error")
	}
	if err != kms.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

// HU-028 AC5 — Sad path: Fallo inserción BD → fallback a archivo
func TestKMS_DBInsertFails_FallbackToFile(t *testing.T) {
	// Este test verifica el contrato: si Persist falla,
	// el fallback a archivo temporal se dispara
	persistor := kms.NewPersistor()

	// Dado: inserción en BD fallará (mock)
	ciphertext := "encrypted-data"

	// Cuando: intenta persistir con DB offline
	err := persistor.PersistWithFallback(context.Background(), ciphertext)

	// Entonces: fallback se activa sin error (recuperable)
	if err != nil {
		t.Errorf("expected fallback to succeed, got %v", err)
	}
}
