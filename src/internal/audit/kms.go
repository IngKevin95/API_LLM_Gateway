package audit

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// KMSClient define las operaciones para obtener llaves de cifrado (Envelope Encryption).
type KMSClient interface {
	// EncryptPayload cifra los datos directamente (simplificado para uso del worker)
	EncryptPayload(ctx context.Context, payload []byte) ([]byte, string, error)
	DecryptPayload(ctx context.Context, ciphertext []byte, dekID string) ([]byte, error)
}

// LocalKMS implementa KMSClient usando una llave maestra estática en memoria.
// Ideal para desarrollo y pruebas.
type LocalKMS struct {
	masterKey []byte
	keyID     string
}

// NewLocalKMS crea un KMS local con una llave maestra (debe ser 32 bytes para AES-256).
func NewLocalKMS(key string) *LocalKMS {
	return &LocalKMS{
		masterKey: []byte(key),
		keyID:     "local-static-v1",
	}
}

// EncryptPayload cifra el payload usando AES-GCM.
// Retorna el payload cifrado (con nonce prefijado) y el ID de la llave usada.
func (k *LocalKMS) EncryptPayload(ctx context.Context, payload []byte) ([]byte, string, error) {
	if len(k.masterKey) != 16 && len(k.masterKey) != 24 && len(k.masterKey) != 32 {
		return nil, "", errors.New("invalid master key size: must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(k.masterKey)
	if err != nil {
		return nil, "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, payload, nil)
	// Retornamos el ciphertext en base64 para facilitar el almacenamiento en BYTEA (o lo retornamos raw)
	// Para mayor compacidad y seguridad, PostgreSQL BYTEA puede almacenar raw bytes.
	return ciphertext, k.keyID, nil
}

// DecryptPayload descifra el payload usando AES-GCM.
func (k *LocalKMS) DecryptPayload(ctx context.Context, ciphertext []byte, dekID string) ([]byte, error) {
	if dekID != k.keyID {
		return nil, fmt.Errorf("unknown dekID: %s", dekID)
	}

	block, err := aes.NewCipher(k.masterKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Seal implementa audit.Encryptor. Serializa el evento como JSON, lo cifra y retorna EncryptedEvent.
func (k *LocalKMS) Seal(e Event) (EncryptedEvent, error) {

	b, err := json.Marshal(e)
	if err != nil {
		return EncryptedEvent{}, err
	}

	ctx := context.Background()
	cipher, dek, err := k.EncryptPayload(ctx, b)
	if err != nil {
		return EncryptedEvent{}, err
	}

	return EncryptedEvent{
		Ciphertext: cipher,
		DEKID:      dek,
	}, nil
}

// Base64 helper (opcional, si se prefiere base64)
func encodeBase64(b []byte) []byte {
	out := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(out, b)
	return out
}
