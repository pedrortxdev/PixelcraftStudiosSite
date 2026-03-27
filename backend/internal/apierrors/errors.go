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

// Checkout/Payment Errors
var (
	ErrInsufficientBalance = errors.New("saldo insuficiente")
	ErrInsufficientStock   = errors.New("estoque insuficiente")
	ErrInvalidDiscount     = errors.New("cupom inválido")
	ErrDiscountExpired     = errors.New("cupom expirado")
	ErrDiscountExhausted   = errors.New("cupom esgotado")
	ErrDiscountInactive    = errors.New("cupom não está ativo")
)

// Discount Management Errors
var (
	ErrDiscountNotFound          = errors.New("cupom não encontrado")
	ErrDiscountCodeAlreadyExists = errors.New("código do cupom já existe")
	ErrDiscountInvalidValue      = errors.New("valor do desconto inválido")
	ErrDiscountInvalidPercentage = errors.New("porcentagem deve estar entre 0 e 100")
	ErrDiscountNegativeValue     = errors.New("valor do desconto não pode ser negativo")
)

// Product Errors
var (
	ErrProductNotFound = errors.New("produto não encontrado")
	ErrProductInactive = errors.New("produto inativo")
)

// Subscription Errors
var (
	ErrSubscriptionNotFound = errors.New("assinatura não encontrada")
	ErrSubscriptionInvalid  = errors.New("assinatura inválida")
)

// Support Errors
var (
	ErrTicketNotFound     = errors.New("ticket não encontrado")
	ErrTicketUnauthorized = errors.New("não autorizado para este ticket")
)

// File Errors
var (
	ErrFileNotFound    = errors.New("arquivo não encontrado")
	ErrFileUnauthorized = errors.New("não autorizado para este arquivo")
)

// User Errors
var (
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrEmailAlreadyExists = errors.New("email já cadastrado")
	ErrInvalidToken       = errors.New("token inválido ou já utilizado")
	ErrTokenExpired       = errors.New("token expirado")
	ErrInvalidVerification = errors.New("código de verificação incorreto")
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

	// Checkout/Payment Errors
	if errors.Is(err, ErrInsufficientBalance) {
		return APIError{Error: err.Error(), Code: "ERR_INSUFFICIENT_BALANCE"}
	}
	if errors.Is(err, ErrInsufficientStock) {
		return APIError{Error: err.Error(), Code: "ERR_INSUFFICIENT_STOCK"}
	}
	if errors.Is(err, ErrInvalidDiscount) {
		return APIError{Error: err.Error(), Code: "ERR_INVALID_DISCOUNT"}
	}
	if errors.Is(err, ErrDiscountExpired) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_EXPIRED"}
	}
	if errors.Is(err, ErrDiscountExhausted) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_EXHAUSTED"}
	}
	if errors.Is(err, ErrDiscountInactive) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_INACTIVE"}
	}

	// Discount Management Errors
	if errors.Is(err, ErrDiscountNotFound) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_NOT_FOUND"}
	}
	if errors.Is(err, ErrDiscountCodeAlreadyExists) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_CODE_EXISTS"}
	}
	if errors.Is(err, ErrDiscountInvalidValue) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_INVALID_VALUE"}
	}
	if errors.Is(err, ErrDiscountInvalidPercentage) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_INVALID_PERCENTAGE"}
	}
	if errors.Is(err, ErrDiscountNegativeValue) {
		return APIError{Error: err.Error(), Code: "ERR_DISCOUNT_NEGATIVE_VALUE"}
	}

	// Product Errors
	if errors.Is(err, ErrProductNotFound) {
		return APIError{Error: err.Error(), Code: "ERR_PRODUCT_NOT_FOUND"}
	}
	if errors.Is(err, ErrProductInactive) {
		return APIError{Error: err.Error(), Code: "ERR_PRODUCT_INACTIVE"}
	}

	// Subscription Errors
	if errors.Is(err, ErrSubscriptionNotFound) {
		return APIError{Error: err.Error(), Code: "ERR_SUBSCRIPTION_NOT_FOUND"}
	}
	if errors.Is(err, ErrSubscriptionInvalid) {
		return APIError{Error: err.Error(), Code: "ERR_SUBSCRIPTION_INVALID"}
	}

	// Support Errors
	if errors.Is(err, ErrTicketNotFound) {
		return APIError{Error: err.Error(), Code: "ERR_TICKET_NOT_FOUND"}
	}
	if errors.Is(err, ErrTicketUnauthorized) {
		return APIError{Error: err.Error(), Code: "ERR_TICKET_UNAUTHORIZED"}
	}

	// File Errors
	if errors.Is(err, ErrFileNotFound) {
		return APIError{Error: err.Error(), Code: "ERR_FILE_NOT_FOUND"}
	}
	if errors.Is(err, ErrFileUnauthorized) {
		return APIError{Error: err.Error(), Code: "ERR_FILE_UNAUTHORIZED"}
	}

	// User Errors
	if errors.Is(err, ErrUserNotFound) {
		return APIError{Error: err.Error(), Code: "ERR_USER_NOT_FOUND"}
	}
	if errors.Is(err, ErrEmailAlreadyExists) {
		return APIError{Error: err.Error(), Code: "ERR_EMAIL_ALREADY_EXISTS"}
	}
	if errors.Is(err, ErrInvalidToken) {
		return APIError{Error: err.Error(), Code: "ERR_INVALID_TOKEN"}
	}
	if errors.Is(err, ErrTokenExpired) {
		return APIError{Error: err.Error(), Code: "ERR_TOKEN_EXPIRED"}
	}
	if errors.Is(err, ErrInvalidVerification) {
		return APIError{Error: err.Error(), Code: "ERR_INVALID_VERIFICATION"}
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
