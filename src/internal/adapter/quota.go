package adapter

import (
	"net/http"
	"strconv"
	"time"
)

// QuotaInfo normaliza información de rate limit desde headers HTTP
type QuotaInfo struct {
	Limit     int64
	Remaining int64
	ResetAt   *time.Time
}

// ExtractQuota parsea headers HTTP estándar (X-RateLimit-*, Retry-After, variantes por proveedor)
// y devuelve estructura normalizada QuotaInfo.
// Si no hay headers de cuota, devuelve QuotaInfo{Limit: 0, Remaining: 0, ResetAt: nil}
func ExtractQuota(headers http.Header) QuotaInfo {
	quota := QuotaInfo{
		Limit:     0,
		Remaining: 0,
		ResetAt:   nil,
	}

	// Intentar parsear OpenAI estándar (X-RateLimit-Limit-Requests, etc.)
	if ok := parseOpenAIHeaders(headers, &quota); ok {
		return quota
	}

	// Intentar parsear Anthropic (anthropic-ratelimit-requests-*)
	if ok := parseAnthropicHeaders(headers, &quota); ok {
		return quota
	}

	// Intentar parsear genérico x-ratelimit-* (Groq, otros)
	if ok := parseGenericHeaders(headers, &quota); ok {
		return quota
	}

	// No headers encontrados, devolver vacío (graceful fallback)
	return quota
}

func parseOpenAIHeaders(headers http.Header, quota *QuotaInfo) bool {
	limitStr := getHeaderCaseInsensitive(headers, "x-ratelimit-limit-requests")
	remainingStr := getHeaderCaseInsensitive(headers, "x-ratelimit-remaining-requests")
	resetStr := getHeaderCaseInsensitive(headers, "x-ratelimit-reset-requests")

	if limitStr == "" || remainingStr == "" {
		return false
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return false
	}
	remaining, err := strconv.ParseInt(remainingStr, 10, 64)
	if err != nil {
		return false
	}

	quota.Limit = limit
	quota.Remaining = remaining

	// Parsear ResetAt desde Retry-After o X-RateLimit-Reset
	resetAt := parseResetTime(headers, resetStr)
	quota.ResetAt = resetAt

	return true
}

func parseAnthropicHeaders(headers http.Header, quota *QuotaInfo) bool {
	limitStr := getHeaderCaseInsensitive(headers, "Anthropic-Ratelimit-Requests-Limit")
	remainingStr := getHeaderCaseInsensitive(headers, "Anthropic-Ratelimit-Requests-Remaining")
	resetStr := getHeaderCaseInsensitive(headers, "Anthropic-Ratelimit-Requests-Reset")

	if limitStr == "" || remainingStr == "" {
		return false
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return false
	}
	remaining, err := strconv.ParseInt(remainingStr, 10, 64)
	if err != nil {
		return false
	}

	quota.Limit = limit
	quota.Remaining = remaining
	resetAt := parseResetTime(headers, resetStr)
	quota.ResetAt = resetAt

	return true
}

func parseGenericHeaders(headers http.Header, quota *QuotaInfo) bool {
	limitStr := getHeaderCaseInsensitive(headers, "X-RateLimit-Limit")
	remainingStr := getHeaderCaseInsensitive(headers, "X-RateLimit-Remaining")
	resetStr := getHeaderCaseInsensitive(headers, "X-RateLimit-Reset")

	if limitStr == "" || remainingStr == "" {
		return false
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return false
	}
	remaining, err := strconv.ParseInt(remainingStr, 10, 64)
	if err != nil {
		return false
	}

	quota.Limit = limit
	quota.Remaining = remaining
	resetAt := parseResetTime(headers, resetStr)
	quota.ResetAt = resetAt

	return true
}

// getHeaderCaseInsensitive obtiene header sin importar mayúsculas (HTTP Headers son case-insensitive)
func getHeaderCaseInsensitive(headers http.Header, name string) string {
	return headers.Get(name)
}

// parseResetTime parsea Retry-After (segundos o RFC1123) o unix timestamp
func parseResetTime(headers http.Header, unixTimestampStr string) *time.Time {
	// Retry-After: segundos o RFC1123
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
			t := time.Now().Add(time.Duration(seconds) * time.Second)
			return &t
		}
		if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
			return &t
		}
	}
	// Fallback: unix timestamp
	if ts, err := strconv.ParseInt(unixTimestampStr, 10, 64); err == nil {
		t := time.Unix(ts, 0)
		return &t
	}
	return nil
}
