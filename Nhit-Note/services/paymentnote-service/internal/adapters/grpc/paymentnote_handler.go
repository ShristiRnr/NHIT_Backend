package grpcadapter

import (
	"context"

	paymentnotepb "nhit-note/api/pb/paymentnotepb"
	"nhit-note/services/paymentnote-service/internal/core/services"

	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentNoteHandler struct {
	paymentnotepb.UnimplementedPaymentNoteServiceServer
	service *services.PaymentNoteService
}

// NewPaymentNoteHandler creates a new payment note gRPC handler
func NewPaymentNoteHandler(service *services.PaymentNoteService) *PaymentNoteHandler {
	return &PaymentNoteHandler{
		service: service,
	}
}

// ListPaymentNotes lists payment notes
func (h *PaymentNoteHandler) ListPaymentNotes(ctx context.Context, req *paymentnotepb.ListPaymentNotesRequest) (*paymentnotepb.ListPaymentNotesResponse, error) {
	return h.service.ListPaymentNotes(ctx, req)
}

// ListDraftPaymentNotes lists draft payment notes
func (h *PaymentNoteHandler) ListDraftPaymentNotes(ctx context.Context, req *paymentnotepb.ListDraftPaymentNotesRequest) (*payment notepb.ListPaymentNotesResponse, error) {
	return h.service.ListDraftPaymentNotes(ctx, req)
}

// GetPaymentNote gets a payment note by ID
func (h *PaymentNoteHandler) GetPaymentNote(ctx context.Context, req *paymentnotepb.GetPaymentNoteRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.GetPaymentNote(ctx, req)
}

// CreatePaymentNote creates a new payment note
func (h *PaymentNoteHandler) CreatePaymentNote(ctx context.Context, req *paymentnotepb.CreatePaymentNoteRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.CreatePaymentNote(ctx, req)
}

// UpdatePaymentNote updates a payment note
func (h *PaymentNoteHandler) UpdatePaymentNote(ctx context.Context, req *paymentnotepb.UpdatePaymentNoteRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.UpdatePaymentNote(ctx, req)
}

// DeletePaymentNote deletes a payment note
func (h *PaymentNoteHandler) DeletePaymentNote(ctx context.Context, req *paymentnotepb.DeletePaymentNoteRequest) (*emptypb.Empty, error) {
	return h.service.DeletePaymentNote(ctx, req)
}

// CreateDraftFromGreenNote creates a draft payment note from an approved green note
func (h *PaymentNoteHandler) CreateDraftFromGreenNote(ctx context.Context, req *paymentnotepb.CreateDraftFromGreenNoteRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.CreateDraftFromGreenNote(ctx, req)
}

// ConvertDraftToActive converts a draft to active payment note
func (h *PaymentNoteHandler) ConvertDraftToActive(ctx context.Context, req *paymentnotepb.ConvertDraftToActiveRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.ConvertDraftToActive(ctx, req)
}

// DeleteDraftPaymentNote deletes a draft payment note
func (h *PaymentNoteHandler) DeleteDraftPaymentNote(ctx context.Context, req *paymentnotepb.DeleteDraftPaymentNoteRequest) (*emptypb.Empty, error) {
	return h.service.DeleteDraftPaymentNote(ctx, req)
}

// PutPaymentNoteOnHold puts a payment note on hold
func (h *PaymentNoteHandler) PutPaymentNoteOnHold(ctx context.Context, req *paymentnotepb.PutPaymentNoteOnHoldRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.PutPaymentNoteOnHold(ctx, req)
}

// RemovePaymentNoteFromHold removes a payment note from hold
func (h *PaymentNoteHandler) RemovePaymentNoteFromHold(ctx context.Context, req *paymentnotepb.RemovePaymentNoteFromHoldRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.RemovePaymentNoteFromHold(ctx, req)
}

// UpdatePaymentNoteUtr updates UTR information
func (h *PaymentNoteHandler) UpdatePaymentNoteUtr(ctx context.Context, req *paymentnotepb.UpdatePaymentNoteUtrRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	return h.service.UpdatePaymentNoteUtr(ctx, req)
}

// GeneratePaymentNoteOrderNumber generates a payment note order number
func (h *PaymentNoteHandler) GeneratePaymentNoteOrderNumber(ctx context.Context, req *emptypb.Empty) (*paymentnotepb.GeneratePaymentNoteOrderNumberResponse, error) {
	return h.service.GeneratePaymentNoteOrderNumber(ctx, req)
}

// DownloadPaymentNotePdf downloads payment note PDF
func (h *PaymentNoteHandler) DownloadPaymentNotePdf(ctx context.Context, req *paymentnotepb.DownloadPaymentNotePdfRequest) (*paymentnotepb.DownloadPaymentNotePdfResponse, error) {
	// TODO: Implement PDF generation
	return &paymentnotepb.DownloadPaymentNotePdfResponse{
		FileContent: []byte{},
		Filename:    "payment_note.pdf",
		ContentType: "application/pdf",
	}, nil
}

// CreatePaymentNoteForSuperAdmin creates a payment note for super admin
func (h *PaymentNoteHandler) CreatePaymentNoteForSuperAdmin(ctx context.Context, req *paymentnotepb.CreatePaymentNoteForSuperAdminRequest) (*paymentnotepb.PaymentNoteResponse, error) {
	// Convert to regular create request
	createReq := &paymentnotepb.CreatePaymentNoteRequest{
		Note: &paymentnotepb.PaymentNotePayload{
			GreenNoteId:            req.GreenNoteId,
			Subject:                req.Subject,
			RecommendationOfPayment: req.RecommendationOfPayment,
			IsDraft:                req.CreateAsDraft,
			CreatedBy:              req.ActorId,
			UserId:                 req.ActorId,
		},
	}

	return h.service.CreatePaymentNote(ctx, createReq)
}

// TestPaymentNoteAPI is a test endpoint
func (h *PaymentNoteHandler) TestPaymentNoteAPI(ctx context.Context, req *emptypb.Empty) (*paymentnotepb.TestPaymentNoteAPIResponse, error) {
	return &paymentnotepb.TestPaymentNoteAPIResponse{
		Status:    "ok",
		Message:   "Payment Note Service is running",
		Timestamp: 0,
	}, nil
}
