package services

import (
	"context"
	"fmt"
	"time"

	"nhit-note/services/paymentnote-service/internal/core/domain"
	"nhit-note/services/paymentnote-service/internal/core/ports"
	"nhit-note/services/paymentnote-service/internal/utils"
	paymentnotepb "nhit-note/api/pb/paymentnotepb"
)

type paymentNoteService struct {
	repo ports.PaymentNoteRepository
}

// NewPaymentNoteService creates a new payment note service
func NewPaymentNoteService(repo ports.PaymentNoteRepository) *paymentNoteService {
	return &paymentNoteService{
		repo: repo,
	}
}

// CreatePaymentNote creates a new payment note with financial calculations
func (s *paymentNoteService) CreatePaymentNote(ctx context.Context, payload *paymentnotepb.PaymentNotePayload) (*domain.PaymentNote, error) {
	// Convert proto to domain
	note := s.protoToDomain(payload)
	
	// Calculate financial fields if not provided
	note = s.calculateFinancials(note)
	
	// Generate note number if not provided
	if note.NoteNo == "" {
		noteNo, err := s.repo.GenerateOrderNumber(ctx, "PN")
		if err != nil {
			return nil, fmt.Errorf("failed to generate note number: %w", err)
		}
		note.NoteNo = noteNo
	}
	
	// Create in database
	created, err := s.repo.Create(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment note: %w", err)
	}
	
	return created, nil
}

// UpdatePaymentNote updates an existing payment note
func (s *paymentNoteService) UpdatePaymentNote(ctx context.Context, id int64, payload *paymentnotepb.PaymentNotePayload) (*domain.PaymentNote, error) {
	// Get existing note
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("payment note not found: %w", err)
	}
	
	// Convert proto to domain
	updated := s.protoToDomain(payload)
	updated.ID = existing.ID
	updated.NoteNo = existing.NoteNo // Keep existing note number
	updated.CreatedAt = existing.CreatedAt
	
	// Recalculate financials
	updated = s.calculateFinancials(updated)
	
	// Update in database
	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment note: %w", err)
	}
	
	return result, nil
}

// GetPaymentNoteByID retrieves a payment note by ID
func (s *paymentNoteService) GetPaymentNoteByID(ctx context.Context, id int64) (*domain.PaymentNote, error) {
	return s.repo.GetByID(ctx, id)
}

// GetPaymentNoteByNoteNo retrieves a payment note by note number
func (s *paymentNoteService) GetPaymentNoteByNoteNo(ctx context.Context, noteNo string) (*domain.PaymentNote, error) {
	return s.repo.GetByNoteNo(ctx, noteNo)
}

// ListPaymentNotes retrieves payment notes with filters
func (s *paymentNoteService) ListPaymentNotes(ctx context.Context, filters domain.PaymentNoteFilters) ([]*domain.PaymentNote, int64, error) {
	return s.repo.List(ctx, filters)
}

// DeletePaymentNote deletes a payment note
func (s *paymentNoteService) DeletePaymentNote(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// CreateDraftFromGreenNote creates a draft payment note from an approved green note
func (s *paymentNoteService) CreateDraftFromGreenNote(ctx context.Context, greenNoteID string, greenNoteData *paymentnotepb.PaymentGreenNoteReference) (*domain.PaymentNote, error) {
	// Check if draft already exists
	existing, err := s.repo.GetByGreenNoteID(ctx, greenNoteID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already exists, return it
	}
	
	// Create draft payment note
	note := &domain.PaymentNote{
		GreenNoteID:       &greenNoteID,
		GreenNoteNo:       stringPtr(greenNoteData.GetGreenNoteNo()),
		GreenNoteApprover: stringPtr(greenNoteData.GetApproverName()),
		Status:            "D", // Draft
		IsDraft:           true,
		AutoCreated:       true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	
	// Generate note number
	noteNo, err := s.repo.GenerateOrderNumber(ctx, "PN")
	if err != nil {
		return nil, fmt.Errorf("failed to generate note number: %w", err)
	}
	note.NoteNo = noteNo
	
	// Create in database
	created, err := s.repo.Create(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}
	
	return created, nil
}

// ConvertDraftToActive converts a draft payment note to active status
func (s *paymentNoteService) ConvertDraftToActive(ctx context.Context, id int64) (*domain.PaymentNote, error) {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	if !note.IsDraft {
		return nil, fmt.Errorf("payment note is not a draft")
	}
	
	// Update status to Pending
	return s.repo.UpdateStatus(ctx, id, "P", false)
}

// PutOnHold puts a payment note on hold
func (s *paymentNoteService) PutOnHold(ctx context.Context, id int64, reason string, userID int64) (*domain.PaymentNote, error) {
	return s.repo.PutOnHold(ctx, id, reason, userID)
}

// RemoveFromHold removes a payment note from hold
func (s *paymentNoteService) RemoveFromHold(ctx context.Context, id int64, newStatus string) (*domain.PaymentNote, error) {
	return s.repo.RemoveFromHold(ctx, id, newStatus)
}

// UpdateUTR updates the UTR information
func (s *paymentNoteService) UpdateUTR(ctx context.Context, id int64, utrNo string, utrDate string) (*domain.PaymentNote, error) {
	return s.repo.UpdateUTR(ctx, id, utrNo, utrDate)
}

// AddComment adds a comment to a payment note
func (s *paymentNoteService) AddComment(ctx context.Context, paymentNoteID int64, comment string, status string, userID int64, userName string, userEmail string) (*domain.PaymentComment, error) {
	c := &domain.PaymentComment{
		PaymentNoteID: paymentNoteID,
		Comment:       comment,
		Status:        stringPtr(status),
		UserID:        userID,
		UserName:      stringPtr(userName),
		UserEmail:     stringPtr(userEmail),
	}
	return s.repo.AddComment(ctx, c)
}

// GeneratePaymentNoteOrderNumber generates a payment note order number
func (s *paymentNoteService) GeneratePaymentNoteOrderNumber(ctx context.Context) (string, error) {
	return s.repo.GenerateOrderNumber(ctx, "PN")
}

// UploadDocument uploads a document to a payment note
func (s *paymentNoteService) UploadDocument(ctx context.Context, paymentNoteID int64, filename string, data []byte, mimeType string, uploadedBy int64, uploadedByName string) (*domain.PaymentNoteDocument, error) {
	return s.repo.UploadDocument(ctx, paymentNoteID, filename, data, mimeType, uploadedBy, uploadedByName)
}

// calculateFinancials calculates all financial fields for a payment note
func (s *paymentNoteService) calculateFinancials(note *domain.PaymentNote) *domain.PaymentNote {
	// Calculate total additions
	totalAdd := 0.0
	for _, p := range note.AddParticulars {
		totalAdd += p.Amount
	}
	note.TotalAdditions = totalAdd
	
	// Calculate total deductions (including TDS)
	totalDeduct := 0.0
	for _, p := range note.LessParticulars {
		totalDeduct += p.Amount
	}
	
	// Calculate TDS if percentage is provided
	if note.TdsPercentage > 0 {
		note.TdsAmount = utils.CalculateTDS(note.GrossAmount, note.TdsPercentage)
		totalDeduct += note.TdsAmount
	}
	note.TotalDeductions = totalDeduct
	
	// Calculate net payable: Gross + Additions - Deductions
	note.NetPayableAmount = utils.CalculateNetPayable(note.GrossAmount, note.TotalAdditions, note.TotalDeductions)
	
	// Round off
	note.NetPayableRoundOff = utils.RoundOff(note.NetPayableAmount)
	
	// Convert to words
	words := utils.NumberToWords(note.NetPayableRoundOff)
	note.NetPayableWords = &words
	
	return note
}

// protoToDomain converts proto payload to domain model
func (s *paymentNoteService) protoToDomain(proto *paymentnotepb.PaymentNotePayload) *domain.PaymentNote {
	note := &domain.PaymentNote{
		UserID:                 proto.UserId,
		GreenNoteID:            protoStringPtr(proto.GreenNoteId),
		GreenNoteNo:            protoStringPtr(proto.GreenNoteNo),
		GreenNoteApprover:      protoStringPtr(proto.GreenNoteApprover),
		GreenNoteAppDate:       protoStringPtr(proto.GreenNoteAppDate),
		ReimbursementNoteID:    protoInt64Ptr(proto.ReimbursementNoteId),
		NoteNo:                 proto.NoteNo,
		Subject:                protoStringPtr(proto.Subject),
		Department:             protoStringPtr(proto.Department),
		VendorCode:             protoStringPtr(proto.VendorCode),
		VendorName:             protoStringPtr(proto.VendorName),
		ProjectName:            protoStringPtr(proto.ProjectName),
		InvoiceNo:              protoStringPtr(proto.InvoiceNo),
		InvoiceDate:            protoStringPtr(proto.InvoiceDate),
		InvoiceAmount:          proto.InvoiceAmount,
		InvoiceApprovedBy:      protoStringPtr(proto.InvoiceApprovedBy),
		LoaPoNo:                protoStringPtr(proto.LoaPoNo),
		LoaPoAmount:            proto.LoaPoAmount,
		LoaPoDate:              protoStringPtr(proto.LoaPoDate),
		GrossAmount:            proto.GrossAmount,
		TdsPercentage:          proto.TdsPercentage,
		TdsSection:             protoStringPtr(proto.TdsSection),
		AccountHolderName:      protoStringPtr(proto.AccountHolderName),
		BankName:               protoStringPtr(proto.BankName),
		AccountNumber:          protoStringPtr(proto.AccountNumber),
		IfscCode:               protoStringPtr(proto.IfscCode),
		RecommendationOfPayment: protoStringPtr(proto.RecommendationOfPayment),
		Status:                 proto.Status,
		IsDraft:                proto.IsDraft,
		AutoCreated:            proto.AutoCreated,
		CreatedBy:              protoInt64Ptr(proto.CreatedBy),
	}
	
	// Parse date if provided
	if proto.Date != "" {
		if t, err := time.Parse(time.RFC3339, proto.Date); err == nil {
			note.Date = &t
		}
	}
	
	// Convert add particulars
	for _, p := range proto.AddParticulars {
		note.AddParticulars = append(note.AddParticulars, domain.PaymentParticular{
			ParticularType: "ADD",
			Particular:     p.Particular,
			Amount:         p.Amount,
		})
	}
	
	// Convert less particulars
	for _, p := range proto.LessParticulars {
		note.LessParticulars = append(note.LessParticulars, domain.PaymentParticular{
			ParticularType: "LESS",
			Particular:     p.Particular,
			Amount:         p.Amount,
		})
	}
	
	// Note: Documents are handled separately via UploadDocument
	
	return note
}

// domainToProto converts domain model to proto response
func (s *paymentNoteService) DomainToProto(note *domain.PaymentNote) *paymentnotepb.PaymentNote {
	proto := &paymentnotepb.PaymentNote{
		Id:                     note.ID,
		UserId:                 note.UserID,
		GreenNoteId:            int64PtrToProto(protoStringToInt64(note.GreenNoteID)),
		GreenNoteNo:            stringPtrToProto(note.GreenNoteNo),
		GreenNoteApprover:      stringPtrToProto(note.GreenNoteApprover),
		GreenNoteAppDate:       stringPtrToProto(note.GreenNoteAppDate),
		ReimbursementNoteId:    int64PtrToProto(note.ReimbursementNoteID),
		NoteNo:                 note.NoteNo,
		Subject:                stringPtrToProto(note.Subject),
		Date:                   timePtrToProto(note.Date),
		Department:             stringPtrToProto(note.Department),
		VendorCode:             stringPtrToProto(note.VendorCode),
		VendorName:             stringPtrToProto(note.VendorName),
		ProjectName:            stringPtrToProto(note.ProjectName),
		InvoiceNo:              stringPtrToProto(note.InvoiceNo),
		InvoiceDate:            stringPtrToProto(note.InvoiceDate),
		InvoiceAmount:          note.InvoiceAmount,
		InvoiceApprovedBy:      stringPtrToProto(note.InvoiceApprovedBy),
		LoaPoNo:                stringPtrToProto(note.LoaPoNo),
		LoaPoAmount:            note.LoaPoAmount,
		LoaPoDate:              stringPtrToProto(note.LoaPoDate),
		GrossAmount:            note.GrossAmount,
		TotalAdditions:         note.TotalAdditions,
		TotalDeductions:        note.TotalDeductions,
		NetPayableAmount:       note.NetPayableAmount,
		NetPayableRoundOff:     note.NetPayableRoundOff,
		NetPayableWords:        stringPtrToProto(note.NetPayableWords),
		TdsPercentage:          note.TdsPercentage,
		TdsSection:             stringPtrToProto(note.TdsSection),
		TdsAmount:              note.TdsAmount,
		AccountHolderName:      stringPtrToProto(note.AccountHolderName),
		BankName:               stringPtrToProto(note.BankName),
		AccountNumber:          stringPtrToProto(note.AccountNumber),
		IfscCode:               stringPtrToProto(note.IfscCode),
		RecommendationOfPayment: stringPtrToProto(note.RecommendationOfPayment),
		Status:                 note.Status,
		IsDraft:                note.IsDraft,
		AutoCreated:            note.AutoCreated,
		CreatedBy:              int64PtrToProto(note.CreatedBy),
		CreatedAt:              note.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              note.UpdatedAt.Format(time.RFC3339),
	}
	
	// Convert add particulars
	for _, p := range note.AddParticulars {
		proto.AddParticulars = append(proto.AddParticulars, &paymentnotepb.PaymentParticular{
			Id:         p.ID,
			Particular: p.Particular,
			Amount:     p.Amount,
		})
	}
	
	// Convert less particulars
	for _, p := range note.LessParticulars {
		proto.LessParticulars = append(proto.LessParticulars, &paymentnotepb.PaymentParticular{
			Id:         p.ID,
			Particular: p.Particular,
			Amount:     p.Amount,
		})
	}
	
	// Convert approval logs
	for _, log := range note.ApprovalLogs {
		proto.ApprovalLogs = append(proto.ApprovalLogs, &paymentnotepb.PaymentApprovalLog{
			Id:            log.ID,
			Status:        log.Status,
			Comments:      stringPtrToProto(log.Comments),
			ReviewerId:    log.ReviewerID,
			ReviewerName:  stringPtrToProto(log.ReviewerName),
			ReviewerEmail: stringPtrToProto(log.ReviewerEmail),
			ApproverLevel: int32PtrToProto(log.ApproverLevel),
			CreatedAt:     log.CreatedAt.Format(time.RFC3339),
		})
	}
	
	// Convert comments
	for _, c := range note.Comments {
		proto.Comments = append(proto.Comments, &paymentnotepb.PaymentNoteComment{
			Id:      c.ID,
			Comment: c.Comment,
			Status:  stringPtrToProto(c.Status),
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		})
	}
	
	// Convert documents
	for _, d := range note.Documents {
		proto.Documents = append(proto.Documents, &paymentnotepb.PaymentNoteDocument{
			Id:               d.ID,
			FileName:         d.FileName,
			OriginalFilename: d.OriginalFilename,
			MimeType:         stringPtrToProto(d.MimeType),
			FileSize:         d.FileSize,
			ObjectKey:        d.ObjectKey,
			UploadedBy:       d.UploadedBy,
			UploadedByName:   stringPtrToProto(d.UploadedByName),
			CreatedAt:        d.CreatedAt.Format(time.RFC3339),
		})
	}
	
	return proto
}

// Helper functions
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func protoStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func protoInt64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func stringPtrToProto(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int64PtrToProto(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func int32PtrToProto(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func timePtrToProto(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func protoStringToInt64(s *string) *int64 {
	// This is a placeholder - you may need to parse string to int64
	// For now, return nil
	return nil
}
