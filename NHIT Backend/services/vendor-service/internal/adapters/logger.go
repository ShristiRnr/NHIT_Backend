package adapters

import (
	"context"
	"fmt"
	"log"

	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/google/uuid"
)

// simpleLogger implements the Logger interface with basic logging
type simpleLogger struct{}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger() ports.Logger {
	return &simpleLogger{}
}

func (l *simpleLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	log.Printf("[INFO] %s %v", msg, fields)
}

func (l *simpleLogger) Error(ctx context.Context, msg string, err error, fields map[string]interface{}) {
	log.Printf("[ERROR] %s: %v %v", msg, err, fields)
}

func (l *simpleLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *simpleLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	log.Printf("[WARN] %s %v", msg, fields)
}

// noOpEventPublisher implements EventPublisher interface with no-ops
type noOpEventPublisher struct{}

// NewNoOpEventPublisher creates a new no-op event publisher
func NewNoOpEventPublisher() ports.EventPublisher {
	return &noOpEventPublisher{}
}

func (p *noOpEventPublisher) PublishVendorCreated(ctx context.Context, vendor *domain.Vendor) error {
	fmt.Println("📢 Event: Vendor created (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishVendorUpdated(ctx context.Context, vendor *domain.Vendor) error {
	fmt.Println("📢 Event: Vendor updated (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishVendorDeleted(ctx context.Context, tenantID, vendorID uuid.UUID) error {
	fmt.Println("📢 Event: Vendor deleted (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishAccountCreated(ctx context.Context, account *domain.VendorAccount) error {
	fmt.Println("📢 Event: Account created (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishAccountUpdated(ctx context.Context, account *domain.VendorAccount) error {
	fmt.Println("📢 Event: Account updated (no-op)")
	return nil
}

func (p *noOpEventPublisher) PublishAccountDeleted(ctx context.Context, accountID uuid.UUID) error {
	fmt.Println("📢 Event: Account deleted (no-op)")
	return nil
}
