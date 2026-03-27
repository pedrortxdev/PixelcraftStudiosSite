package apierrors

import "errors"

// System Error definitions (BT-027)
var (
	ErrNotFound      = errors.New("recurso não encontrado")
	ErrForbidden     = errors.New("acesso negado ou privilégios insuficientes")
	ErrUnauthorized  = errors.New("não autorizado")
	ErrInvalidInput  = errors.New("dados de entrada inválidos")
	ErrInternal      = errors.New("erro interno do servidor")
	ErrConflict      = errors.New("conflito operacional detectado")
	ErrPaymentFailed = errors.New("falha no processamento da transação")
)

// APIError details the standard JSON response format (BT-029)
type APIError struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Convert maps standardized internal errors to public representations preventing leaks (BT-030)
func Convert(err error) APIError {
	if errors.Is(err, ErrNotFound) {
		return APIError{Error: err.Error(), Code: "ERR_NOT_FOUND"}
	}
	if errors.Is(err, ErrForbidden) {
		return APIError{Error: err.Error(), Code: "ERR_FORBIDDEN"}
	}
	if errors.Is(err, ErrUnauthorized) {
		return APIError{Error: err.Error(), Code: "ERR_UNAUTHORIZED"}
	}
	if errors.Is(err, ErrInvalidInput) {
		return APIError{Error: err.Error(), Code: "ERR_INVALID_INPUT"}
	}
	if errors.Is(err, ErrConflict) {
		return APIError{Error: err.Error(), Code: "ERR_CONFLICT"}
	}
	if errors.Is(err, ErrPaymentFailed) {
		return APIError{Error: err.Error(), Code: "ERR_PAYMENT_FAILED"}
	}

	// Strict Fallback masking detailed Go/SQL panics in production logic (BT-030)
	return APIError{Error: ErrInternal.Error(), Code: "ERR_INTERNAL"}
}

// New constructs ad-hoc standard API format (BT-028)
func New(msg, code string) APIError {
	return APIError{
		Error: msg,
		Code:  code,
	}
}
