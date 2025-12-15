package handler

import (
	"context"
	
	"nhit-note/services/paymentnote-service/internal/core/services"
	paymentnotepb "nhit-note/api/pb/paymentnotepb"
)

type PaymentNoteHandler struct {
	paymentnotepb.UnimplementedPaymentNoteServiceServer
	service *services.PaymentNoteService
}

// NewPaymentNoteHandler creates a new payment note handler
func NewPaymentNoteHandler(service *services.PaymentNoteService) *PaymentNoteHandler {
	return &PaymentNoteHandler{
		service: service,
	}
}

// CreatePaymentNote creates a new payment note
func (h *PaymentNoteHandler) CreatePaymentNote(ctx context.Context, req *paymentnotepb.CreatePaymentNoteRequest) (*paymentnotepb.CreatePaymentNoteResponse, error) {
	note, err := h.service.CreatePaymentNote(ctx, req.GetNote())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.CreatePaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// GetPaymentNote retrieves a payment note by ID
func (h *PaymentNoteHandler) GetPaymentNote(ctx context.Context, req *paymentnotepb.GetPaymentNoteRequest) (*paymentnotepb.GetPaymentNoteResponse, error) {
	note, err := h.service.GetPaymentNoteByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.GetPaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// ListPaymentNotes lists payment notes with filters
func (h *PaymentNoteHandler) ListPaymentNotes(ctx context.Context, req *paymentnotepb.ListPaymentNotesRequest) (*paymentnotepb.ListPaymentNotesResponse, error) {
	filters := domain.PaymentNoteFilters{
		Status:  &req.Status,
		IsDraft: &req.IsDraft,
		Search:  &req.Search,
		Page:    req.GetPage(),
		PerPage: req.GetPerPage(),
	}
	
	notes, total, err := h.service.ListPaymentNotes(ctx, filters)
	if err != nil {
		return nil, err
	}
	
	protoNotes := make([]*paymentnotepb.PaymentNote, len(notes))
	for i, note := range notes {
		protoNotes[i] = h.service.DomainToProto(note)
	}
	
	return &paymentnotepb.ListPaymentNotesResponse{
		Notes: protoNotes,
		Total: total,
	}, nil
}

// UpdatePaymentNote updates a payment note
func (h *PaymentNoteHandler) UpdatePaymentNote(ctx context.Context, req *paymentnotepb.UpdatePaymentNoteRequest) (*paymentnotepb.UpdatePaymentNoteResponse, error) {
	note, err := h.service.UpdatePaymentNote(ctx, req.GetId(), req.GetNote())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.UpdatePaymentNoteResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// DeletePaymentNote deletes a payment note
func (h *PaymentNoteHandler) DeletePaymentNote(ctx context.Context, req *paymentnotepb.DeletePaymentNoteRequest) (*paymentnotepb.DeletePaymentNoteResponse, error) {
	err := h.service.DeletePaymentNote(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.DeletePaymentNoteResponse{
		Success: true,
	}, nil
}

// GeneratePaymentNoteOrderNumber generates an order number
func (h *PaymentNoteHandler) GeneratePaymentNoteOrderNumber(ctx context.Context, req *paymentnotepb.GeneratePaymentNoteOrderNumberRequest) (*paymentnotepb.GeneratePaymentNoteOrderNumberResponse, error) {
	orderNo, err := h.service.GeneratePaymentNoteOrderNumber(ctx)
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.GeneratePaymentNoteOrderNumberResponse{
		OrderNumber: orderNo,
	}, nil
}

// PutPaymentNoteOnHold puts a payment note on hold
func (h *PaymentNoteHandler) PutPaymentNoteOnHold(ctx context.Context, req *paymentnotepb.PutPaymentNoteOnHoldRequest) (*paymentnotepb.PutPaymentNoteOnHoldResponse, error) {
	note, err := h.service.PutOnHold(ctx, req.GetId(), req.GetReason(), req.GetUserId())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.PutPaymentNoteOnHoldResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// RemovePaymentNoteFromHold removes a payment note from hold
func (h *PaymentNoteHandler) RemovePaymentNoteFromHold(ctx context.Context, req *paymentnotepb.RemovePaymentNoteFromHoldRequest) (*paymentnotepb.RemovePaymentNoteFromHoldResponse, error) {
	note, err := h.service.RemoveFromHold(ctx, req.GetId(), req.GetNewStatus())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.RemovePaymentNoteFromHoldResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}

// UpdatePaymentNoteUTR updates UTR information
func (h *PaymentNoteHandler) UpdatePaymentNoteUTR(ctx context.Context, req *paymentnotepb.UpdatePaymentNoteUTRRequest) (*paymentnotepb.UpdatePaymentNoteUTRResponse, error) {
	note, err := h.service.UpdateUTR(ctx, req.GetId(), req.GetUtrNo(), req.GetUtrDate())
	if err != nil {
		return nil, err
	}
	
	return &paymentnotepb.UpdatePaymentNoteUTRResponse{
		Note: h.service.DomainToProto(note),
	}, nil
}
