package anthropic_test

import (
	"errors"

	"api-llm-gateway/internal/adapter"
)

func isProviderErr(err error, target **adapter.ProviderError) bool {
	return errors.As(err, target)
}
