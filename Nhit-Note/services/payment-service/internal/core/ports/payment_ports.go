package ports

import (
	"context"
	"nhit-note/services/payment-service/internal/core/domain"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	// CreatePaymentRequests creates multiple payments under the same sl_no
	CreatePaymentRequests(ctx context.Context, slNo string, payments []domain.Payment) error
	
	// GetPaymentGroup retrieves all payments with the given sl_no
	GetPaymentGroup(ctx context.Context, slNo string) (*domain.PaymentGroup, error)
	
	// GetPaymentByID retrieves a single payment by ID
	GetPaymentByID(ctx context.Context, id int64) (*domain.Payment, error)
	
	// List retrieves payment groups with filters
	List(ctx context.Context, filters domain.PaymentFilters) ([]*domain.PaymentGroup, int64, error)
	
	// UpdatePayment updates a single payment
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	
	// UpdatePaymentGroupStatus updates status for all payments in a group
	UpdatePaymentGroupStatus(ctx context.Context, slNo string, status string) error
	
	// DeletePayment deletes a single payment
	DeletePayment(ctx context.Context, id int64) error
	
	// DeletePaymentGroup deletes all payments in a group
	DeletePaymentGroup(ctx context.Context, slNo string) error
	
	// LinkVendorAccount links a payment to a vendor account
	LinkVendorAccount(ctx context.Context, paymentID int64, vendorID int64, vendorAccountID *int64) error
	
	// CreateShortcut creates a payment shortcut
	CreateShortcut(ctx context.Context, shortcut *domain.PaymentShortcut) (*domain.PaymentShortcut, error)
	
	// GetShortcut retrieves a shortcut by ID
	GetShortcut(ctx context.Context, id int64) (*domain.PaymentShortcut, error)
	
	// ListShortcuts lists shortcuts for a user
	ListShortcuts(ctx context.Context, userID int64) ([]*domain.PaymentShortcut, error)
	
	// GenerateSerialNumber generates the next payment serial number
	GenerateSerialNumber(ctx context.Context, prefix string) (string, error)
	
	// AddBankLetterLog adds a bank letter approval log
	AddBankLetterLog(ctx context.Context, log *domain.BankLetterApprovalLog) error
	
	// GetBankLetterLogs retrieves bank letter logs
	GetBankLetterLogs(ctx context.Context, slNo string) ([]*domain.BankLetterApprovalLog, error)
}
