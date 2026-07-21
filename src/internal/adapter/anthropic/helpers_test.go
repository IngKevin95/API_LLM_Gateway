package anthropic_test

import (
	"errors"

	"github.com/IngKevin95/API_LLM_Gateway/internal/adapter"
)

func isProviderErr(err error, target **adapter.ProviderError) bool {
	return errors.As(err, target)
}
