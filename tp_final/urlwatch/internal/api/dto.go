package api

type CheckOptions struct {
	Concurrency int   `json:"concurrency"`
	TimeoutMs   int64 `json:"timeout_ms"`
}

type CreateBatchRequest struct {
	URLs    []string     `json:"urls"`
	Options CheckOptions `json:"options"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
