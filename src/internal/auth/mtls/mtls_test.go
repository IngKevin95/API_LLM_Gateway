package mtls_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IngKevin95/API_LLM_Gateway/internal/auth"
	"github.com/IngKevin95/API_LLM_Gateway/internal/auth/mtls"
)

type ca struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
	der  []byte
}

func newCA(t *testing.T) ca {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, KeyUsage: x509.KeyUsageCertSign, BasicConstraintsValid: true,
	}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	cert, _ := x509.ParseCertificate(der)
	return ca{cert: cert, key: key, der: der}
}

// issue firma un cert cliente con el CA dado; notAfter controla expiración.
func (c ca) issue(t *testing.T, cn string, ou []string, notAfter time.Time) tls.Certificate {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn, OrganizationalUnit: ou, Organization: []string{"T1"}},
		NotBefore:    time.Now().Add(-time.Hour), NotAfter: notAfter,
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, c.cert, &key.PublicKey, c.key)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: mustParse(der)}
}

func mustParse(der []byte) *x509.Certificate { c, _ := x509.ParseCertificate(der); return c }

// server mTLS que confía en trustCA y usa el middleware para extraer identidad.
func newServer(t *testing.T, trustCA ca) *httptest.Server {
	return newServerRevoking(t, trustCA, nil)
}

func newServerRevoking(t *testing.T, trustCA ca, revoked *mtls.RevocationList) *httptest.Server {
	t.Helper()
	pool := x509.NewCertPool()
	pool.AddCert(trustCA.cert)

	h := mtls.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := auth.FromContext(r.Context())
		_, _ = io.WriteString(w, id.Subject)
	}))
	srv := httptest.NewUnstartedServer(h)
	srv.TLS = mtls.ServerTLSConfig(pool, revoked)
	srv.StartTLS()
	return srv
}

// client con el cert dado (o sin cert si clientCert es zero).
func doRequest(srv *httptest.Server, serverCA ca, clientCert *tls.Certificate) (*http.Response, error) {
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	tlsCfg := &tls.Config{RootCAs: roots}
	_ = serverCA
	if clientCert != nil {
		tlsCfg.Certificates = []tls.Certificate{*clientCert}
	}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsCfg}}
	return client.Get(srv.URL)
}

// HU-025b AC1 — Happy: cert cliente válido → conexión + scope extraído.
func TestMTLS_ValidCert(t *testing.T) {
	trust := newCA(t)
	srv := newServer(t, trust)
	defer srv.Close()
	cert := trust.issue(t, "svc-interno", []string{"capability:coding"}, time.Now().Add(time.Hour))
	resp, err := doRequest(srv, trust, &cert)
	if err != nil {
		t.Fatalf("handshake válido no debe fallar: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "svc-interno" {
		t.Errorf("esperaba scope/identidad del cert, obtuve %q", body)
	}
}

// HU-025b AC3 — Error: sin certificado → rechazo TLS.
func TestMTLS_NoCert(t *testing.T) {
	trust := newCA(t)
	srv := newServer(t, trust)
	defer srv.Close()
	if _, err := doRequest(srv, trust, nil); err == nil {
		t.Error("sin cert cliente el handshake debe fallar")
	}
}

// HU-025b AC4 — Edge: CA no confiable → handshake falla.
func TestMTLS_UntrustedCA(t *testing.T) {
	trust := newCA(t)
	srv := newServer(t, trust)
	defer srv.Close()
	rogue := newCA(t) // CA distinta, no en el trust store
	cert := rogue.issue(t, "atacante", nil, time.Now().Add(time.Hour))
	if _, err := doRequest(srv, trust, &cert); err == nil {
		t.Error("cert de CA no confiable debe fallar el handshake")
	}
}

// HU-025b AC2 — Error: certificado expirado → handshake falla.
func TestMTLS_ExpiredCert(t *testing.T) {
	trust := newCA(t)
	srv := newServer(t, trust)
	defer srv.Close()
	cert := trust.issue(t, "svc", nil, time.Now().Add(-time.Minute)) // ya expirado
	if _, err := doRequest(srv, trust, &cert); err == nil {
		t.Error("cert expirado debe fallar el handshake")
	}
}

// HU-025b AC2 — Error: certificado revocado (vigente) → handshake falla.
func TestMTLS_RevokedCert(t *testing.T) {
	trust := newCA(t)
	revoked := mtls.NewRevocationList()
	srv := newServerRevoking(t, trust, revoked)
	defer srv.Close()

	cert := trust.issue(t, "svc-revocado", nil, time.Now().Add(time.Hour)) // vigente...
	revoked.Revoke(cert.Leaf.SerialNumber)                                 // ...pero revocado
	if _, err := doRequest(srv, trust, &cert); err == nil {
		t.Error("cert revocado (aunque vigente) debe fallar el handshake")
	}
}
