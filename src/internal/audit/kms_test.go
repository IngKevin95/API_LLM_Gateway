package audit

import (
	"context"
	"testing"
)

// AC1: Cifra y descifra correctamente con la llave maestra
func TestKMS_EncryptDecrypt(t *testing.T) {
	kms := NewLocalKMS("super_secret_master_key_32bytes!") // Tiene que ser 32 bytes para AES-256

	ctx := context.Background()
	payload := []byte(`{"message": "secret content"}`)

	encrypted, dekID, err := kms.EncryptPayload(ctx, payload)
	if err != nil {
		t.Fatalf("error cifrando: %v", err)
	}

	if len(encrypted) == 0 || string(encrypted) == string(payload) {
		t.Fatal("el payload no fue cifrado correctamente")
	}

	decrypted, err := kms.DecryptPayload(ctx, encrypted, dekID)
	if err != nil {
		t.Fatalf("error descifrando: %v", err)
	}

	if string(decrypted) != string(payload) {
		t.Errorf("esperaba %q, obtuve %q", payload, decrypted)
	}
}

// AC2: Devuelve error si la llave maestra es inválida (ej. longitud incorrecta)
func TestKMS_InvalidKey(t *testing.T) {
	// Menos de 32 bytes
	kms := NewLocalKMS("short_key")
	ctx := context.Background()

	_, _, err := kms.EncryptPayload(ctx, []byte("test"))
	if err == nil {
		t.Fatal("esperaba error por longitud de llave inválida")
	}
}

// AC3: Seal cifra un Event y retorna EncryptedEvent
func TestKMS_Seal(t *testing.T) {
	kms := NewLocalKMS("super_secret_master_key_32bytes!")
	e := Event{ID: "req-123", Timestamp: 123456}
	sealed, err := kms.Seal(e)
	if err != nil {
		t.Fatalf("error sealing: %v", err)
	}
	if len(sealed.Ciphertext) == 0 {
		t.Fatal("ciphertext está vacío")
	}
	if sealed.DEKID != "local-static-v1" {
		t.Errorf("esperaba DEKID local-static-v1, obtuve %s", sealed.DEKID)
	}
}
