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
const (
	ReferralCommissionRate = 0.05 // 5%
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
