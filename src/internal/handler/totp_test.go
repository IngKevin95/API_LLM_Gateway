package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMfaHandler_Enroll(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/auth/mfa/enroll", nil)
	rr := httptest.NewRecorder()
	_ = rr
	_ = req
	h := NewMfaHandler(nil)
	_ = h
}

func TestMfaHandler_Verify(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/auth/mfa/verify", nil)
	rr := httptest.NewRecorder()
	_ = rr
	_ = req
	h := NewMfaHandler(nil)
	_ = h
}

func TestMfaHandler_Disable(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/auth/mfa/disable", nil)
	rr := httptest.NewRecorder()
	_ = rr
	_ = req
	h := NewMfaHandler(nil)
	_ = h
}
