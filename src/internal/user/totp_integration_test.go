//go:build integration
// +build integration

package user

import (
	"context"
	"testing"
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
	// Nota: El test fallará hasta que se implemente y se añada la librería otp en el verde.
	
	// Validar con un código incorrecto
	err = s.MfaVerify(context.Background(), u.ID, "000000")
	if err != ErrInvalidMfaCode {
		t.Fatalf("se esperaba ErrInvalidMfaCode, se obtuvo %v", err)
	}

	// Aquí deberíamos generar un código válido y verificarlo,
	// pero por ahora el test está en rojo y ni siquiera compila porque
	// los métodos no existen en Store.

	// Disable
	err = s.MfaDisable(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("MfaDisable error: %v", err)
	}
}
