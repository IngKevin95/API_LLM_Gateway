// Package mtls autentica servicios por certificado cliente (mutual TLS),
// extrayendo la identidad/scope del certificado verificado.
package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"math/big"
	"net/http"
	"sync"

	"github.com/IngKevin95/API_LLM_Gateway/internal/auth"
)

// RevocationList es una lista de certificados revocados por número de serie,
// consultada en el handshake. Seguro concurrente. (Alternativa liviana a
// CRL/OCSP externos; la fuente de revocación se alimenta desde configuración.)
type RevocationList struct {
	mu      sync.RWMutex
	serials map[string]bool
}

// NewRevocationList crea una lista vacía.
func NewRevocationList() *RevocationList { return &RevocationList{serials: make(map[string]bool)} }

// Revoke marca un serial como revocado.
func (rl *RevocationList) Revoke(serial *big.Int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.serials[serial.String()] = true
}

func (rl *RevocationList) isRevoked(serial *big.Int) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.serials[serial.String()]
}

// ServerTLSConfig arma un tls.Config que exige y verifica el certificado
// cliente contra el trust store dado (rechaza ausentes/CA no confiable/expirados
// en el handshake). Si `revoked` no es nil, además rechaza certificados
// revocados por serial.
func ServerTLSConfig(trustStore *x509.CertPool, revoked *RevocationList) *tls.Config {
	cfg := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  trustStore,
		MinVersion: tls.VersionTLS12,
	}
	if revoked != nil {
		cfg.VerifyPeerCertificate = func(_ [][]byte, chains [][]*x509.Certificate) error {
			for _, chain := range chains {
				if len(chain) > 0 && revoked.isRevoked(chain[0].SerialNumber) {
					return errors.New("mtls: certificado revocado")
				}
			}
			return nil
		}
	}
	return cfg
}

// Middleware extrae la identidad del certificado cliente ya verificado por TLS
// y la inyecta en el context. (El rechazo de certs inválidos ocurre antes, en
// el handshake TLS, no aquí.)
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				http.Error(w, `{"error":"client certificate required"}`, http.StatusUnauthorized)
				return
			}
			cert := r.TLS.PeerCertificates[0]
			id := auth.Identity{
				Subject: cert.Subject.CommonName,
				Scopes:  cert.Subject.OrganizationalUnit, // scope embebido en el cert
			}
			if len(cert.Subject.Organization) > 0 {
				id.Tenant = cert.Subject.Organization[0]
			}
			next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), id)))
		})
	}
}
