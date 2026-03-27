package models

// Payment Methods
const (
	PaymentMethodBalance     = "BALANCE"
	PaymentMethodMercadoPago = "MERCADOPAGO"
	PaymentMethodPix         = "PIX"
	PaymentMethodCreditCard  = "CREDIT_CARD"
	PaymentMethodDebitCard   = "DEBIT_CARD"
	PaymentMethodBoleto      = "BOLETO"
)

// Project Stages
const (
	ProjectStagePlanning      = "Planejamento"
	ProjectStageDesign        = "Design"
	ProjectStageDevelopment   = "Desenvolvimento"
	ProjectStageTesting       = "Testes"
	ProjectStageDeployment    = "Implantação"
	ProjectStageMaintenance   = "Manutenção"
)

// Commission Rates
// ReferralCommissionRate is 5% (use: amount * ReferralCommissionRate / 100 for integer math)
const (
	ReferralCommissionRate = 5 // 5% (divide by 100 when calculating)
)

// Product Types
const (
	ProductTypeDigital  = "digital"
	ProductTypePhysical = "physical"
	ProductTypeService  = "service"
)

// Invoice Status
const (
	InvoiceStatusPaid   = "paid"
	InvoiceStatusDue    = "due"
	InvoiceStatusOverdue = "overdue"
)
