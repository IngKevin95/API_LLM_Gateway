//go:build integration
// +build integration

package user

import (
	"context"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestStore_MfaEnrollVerifyDisable(t *testing.T) {
	db := startTestPostgres(t)
	defer db.Close()

	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	u, _ := s.Create(context.Background(), "mfa@example.com", RoleOperator, "t1", nil)

	// Enroll
	secret, uri, err := s.MfaEnroll(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("MfaEnroll error: %v", err)
	}
	if secret == "" || uri == "" {
		t.Fatalf("secret y uri no pueden estar vacíos")
	}

	// Como no tenemos un generador de tokens en el test para simular al usuario,
	// usaremos la librería otp internamente para generar un token válido para el test.

	// Validar con un código incorrecto
	err = s.MfaVerify(context.Background(), u.ID, "000000")
	if err != ErrInvalidMfaCode {
		t.Fatalf("se esperaba ErrInvalidMfaCode, se obtuvo %v", err)
	}

	// Generar un código válido para el instante actual
	validCode, _ := totp.GenerateCode(secret, time.Now())
	err = s.MfaVerify(context.Background(), u.ID, validCode)
	if err != nil {
		t.Fatalf("MfaVerify falló con código válido: %v", err)
	}

	// Disable
	err = s.MfaDisable(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("MfaDisable error: %v", err)
	}
}
