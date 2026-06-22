package api

import (
	"fmt"
	"strings"

	"github.com/alhai/urlwatch/internal/domain"
)

func validateRequest(req *CreateBatchRequest) error {
	if len(req.URLs) == 0 {
		return &domain.ValidationError{Field: "urls", Message: "required, must have at least 1 URL"}
	}
	if len(req.URLs) > 100 {
		return &domain.ValidationError{Field: "urls", Message: "must have at most 100 URLs"}
	}
	for _, u := range req.URLs {
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			return &domain.ValidationError{Field: "urls", Message: fmt.Sprintf("%q is not a valid http/https URL", u)}
		}
	}
	c := req.Options.Concurrency
	if c != 0 && (c < 1 || c > 50) {
		return &domain.ValidationError{Field: "options.concurrency", Message: "must be between 1 and 50"}
	}
	tm := req.Options.TimeoutMs
	if tm != 0 && (tm < 100 || tm > 30000) {
		return &domain.ValidationError{Field: "options.timeout_ms", Message: "must be between 100 and 30000"}
	}
	return nil
}
