package sqlc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	greennotepb "nhit-note/api/pb/greennotepb"
	sqlcgen "nhit-note/services/greennote-service/internal/adapters/repository/sqlc/generated"
	"nhit-note/services/greennote-service/internal/core/ports"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
)

// Repository is a minimal Postgres-backed implementation of ports.GreenNoteRepository.
//
// For now, this implementation delegates to the in-memory repository to keep the
// runtime behaviour consistent with the simplified greennote.proto API while the
// legacy rich SQLC-backed repository is guarded behind a build tag.
//
// This allows the service to compile and run regardless of whether a database
// URL is configured, without pulling in the legacy GreenNote/SupportingDocument
// types that no longer exist in the protobuf definitions.

type Repository struct {
	db   *sql.DB
	q    *sqlcgen.Queries
	docs ports.DocumentStorage
}

// NewPostgresGreenNoteRepository constructs a repository. The db and docs
// parameters are accepted for API compatibility but are not used by the current
// minimal implementation.
func NewPostgresGreenNoteRepository(db *sql.DB, docs ports.DocumentStorage) *Repository {
	return &Repository{db: db, q: sqlcgen.New(db), docs: docs}
}

func (r *Repository) List(ctx context.Context, req *greennotepb.ListGreenNotesRequest) (*greennotepb.ListGreenNotesResponse, error) {
	if r == nil || r.db == nil {
		return &greennotepb.ListGreenNotesResponse{}, nil
	}

	statusFilter := req.GetStatus()
	page := req.GetPage()
	perPage := req.GetPerPage()
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Total count for pagination
	var total int64
	countQuery := `SELECT COUNT(*) FROM green_notes WHERE ($1 = '' OR status::text = $1)`
	if err := r.db.QueryRowContext(ctx, countQuery, statusFilter).Scan(&total); err != nil {
		return nil, err
	}

	query := `
		SELECT id, project_name, supplier_name, total_amount, created_at, status
		FROM green_notes
		WHERE ($1 = '' OR status::text = $1)
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, statusFilter, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*greennotepb.GreenNoteListItem, 0)
	for rows.Next() {
		var (
			id          string
			projectName sql.NullString
			vendorName  sql.NullString
			amount      float64
			created     time.Time
			status      string
		)
		if err := rows.Scan(&id, &projectName, &vendorName, &amount, &created, &status); err != nil {
			return nil, err
		}
		idStr := id
		item := &greennotepb.GreenNoteListItem{
			Id:          idStr,
			ProjectName: projectName.String,
			VendorName:  vendorName.String,
			Amount:      amount,
			Date:        created.Format(time.RFC3339),
			Status:      status,
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	resp := &greennotepb.ListGreenNotesResponse{
		Notes:   items,
		Page:    page,
		PerPage: perPage,
		Total:   total,
	}
	return resp, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*greennotepb.GreenNotePayload, error) {
	if r == nil || r.db == nil {
		return nil, ports.ErrNotFound
	}
	if r.q == nil {
		return nil, ports.ErrNotFound
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ports.ErrNotFound
	}

	noteRow, err := r.q.GetGreenNote(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	invoices, err := r.q.ListInvoicesByGreenNote(ctx, id)
	if err != nil {
		return nil, err
	}
	docs, err := r.q.ListSupportingDocuments(ctx, id)
	if err != nil {
		return nil, err
	}

	p := &greennotepb.GreenNotePayload{
		ProjectName:                       noteRow.ProjectName.String,
		SupplierName:                      noteRow.SupplierName.String,
		ExpenseCategory:                   noteRow.ExpenseCategory.String,
		ProtestNoteRaised:                 fromNullYesNoEnum(noteRow.ProtestNoteRaised),
		WhetherContract:                   fromNullYesNoEnum(noteRow.WhetherContract),
		ExtensionOfContractPeriodExecuted: fromNullYesNoEnum(noteRow.ExtensionOfContractPeriodExecuted),
		ExpenseAmountWithinContract:       fromNullYesNoEnum(noteRow.ExpenseAmountWithinContract),
		MilestoneAchieved:                 fromNullYesNoEnum(noteRow.MilestoneAchieved),
		PaymentApprovedWithDeviation:      fromNullYesNoEnum(noteRow.PaymentApprovedWithDeviation),
		RequiredDocumentsSubmitted:        fromNullYesNoEnum(noteRow.RequiredDocumentsSubmitted),
		DocumentsVerified:                 noteRow.DocumentsVerified.String,
		ContractStartDate:                 noteRow.ContractStartDate.String,
		ContractEndDate:                   noteRow.ContractEndDate.String,
		AppointedStartDate:                noteRow.AppointedStartDate.String,
		SupplyPeriodStart:                 noteRow.SupplyPeriodStart.String,
		SupplyPeriodEnd:                   noteRow.SupplyPeriodEnd.String,
		ContractPeriodCompleted:           fromNullYesNoEnum(noteRow.ContractPeriodCompleted),
		DepartmentName:                    noteRow.DepartmentName.String,
		WorkOrderNo:                       noteRow.WorkOrderNo.String,
		PoNumber:                          noteRow.PoNumber.String,
		MsmeClassification:                noteRow.MsmeClassification.String,
		ActivityType:                      noteRow.ActivityType.String,
		BriefOfGoodsServices:              noteRow.BriefOfGoodsServices.String,
		Status:                            string(noteRow.Status.StatusEnum),
		EnableMultipleInvoices:            noteRow.EnableMultipleInvoices,
	}

	// Convert numeric fields from string to float64
	p.BaseValue = parseDecimal(noteRow.BaseValue.String)
	p.Gst = parseDecimal(noteRow.Gst.String)
	p.OtherCharges = parseDecimal(noteRow.OtherCharges.String)
	p.TotalAmount = parseDecimal(noteRow.TotalAmount.String)

	p.BudgetExpenditure = parseDecimal(noteRow.BudgetExpenditure.String)
	p.ActualExpenditure = parseDecimal(noteRow.ActualExpenditure.String)
	p.ExpenditureOverBudget = parseDecimal(noteRow.ExpenditureOverBudget.String)

	// Map enums stored as text (if present).
	switch noteRow.ApprovalFor.ApprovalFor {
	case "APPROVAL_FOR_INVOICE":
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_INVOICE
	case "APPROVAL_FOR_ADVANCE":
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_ADVANCE
	case "APPROVAL_FOR_ADHOC":
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_ADHOC
	default:
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_UNSPECIFIED
	}

	switch noteRow.NatureOfExpenses.NatureOfExpenses {
	case "NATURE_OHC_001_MANPOWER":
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OHC_001_MANPOWER
	case "NATURE_OHC_002_STAFF_WELFARE":
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OHC_002_STAFF_WELFARE
	case "NATURE_OHC_003_OFFICE_RENT_UTILITIES":
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OHC_003_OFFICE_RENT_UTILITIES
	default:
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OF_EXPENSES_UNSPECIFIED
	}

	if noteRow.ExpenseCategoryType.Valid {
		switch noteRow.ExpenseCategoryType.ExpenseCategoryType {
		case "EXPENSE_CATEGORY_CAPITAL":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_CAPITAL
		case "EXPENSE_CATEGORY_REVENUE":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_REVENUE
		case "EXPENSE_CATEGORY_OPERATIONAL":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_OPERATIONAL
		case "EXPENSE_CATEGORY_ADMINISTRATIVE":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_ADMINISTRATIVE
		case "EXPENSE_CATEGORY_MAINTENANCE":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_MAINTENANCE
		default:
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_UNSPECIFIED
		}
	} else {
		p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_UNSPECIFIED
	}

	// Map invoice rows back into InvoiceInput messages.
	for _, in := range invoices {
		taxableValue, _ := strconv.ParseFloat(in.TaxableValue, 64)
		gst, _ := strconv.ParseFloat(in.Gst, 64)
		otherCharges, _ := strconv.ParseFloat(in.OtherCharges, 64)
		total, _ := strconv.ParseFloat(in.InvoiceValue, 64)

		inv := &greennotepb.InvoiceInput{
			InvoiceNumber: in.InvoiceNumber,
			InvoiceDate:   in.InvoiceDate.String,
			TaxableValue:  taxableValue,
			Gst:           gst,
			OtherCharges:  otherCharges,
			InvoiceValue:  total,
		}
		p.Invoices = append(p.Invoices, inv)
	}
	if len(p.Invoices) > 0 {
		p.Invoice = p.Invoices[0]
	}

	// Map supporting documents into existing_documents.
	for _, d := range docs {
		p.ExistingDocuments = append(p.ExistingDocuments, &greennotepb.SupportingDocument{
			Id:               d.ID,
			Name:             d.Name,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			FileSize:         d.FileSize,
			ObjectKey:        d.ObjectKey,
			CreatedAt:        timestamppb.New(d.CreatedAt),
			UpdatedAt:        timestamppb.New(d.UpdatedAt),
		})
	}

	return p, nil
}

func (r *Repository) Create(ctx context.Context, payload *greennotepb.GreenNotePayload) (string, error) {
	if r == nil || r.db == nil {
		return "", nil
	}
	if payload == nil {
		return "", nil
	}

	if r.q == nil {
		return "", nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	qtx := r.q.WithTx(tx)

	// Aggregate invoice totals for storage.
	invBase, invGst, invOther, invTotal := sumInvoiceInputs(payload.GetInvoices())
	if invBase == 0 && invGst == 0 && invOther == 0 && invTotal == 0 {
		if payload.GetInvoice() != nil {
			invBase, invGst, invOther, invTotal = sumInvoiceInputs([]*greennotepb.InvoiceInput{payload.GetInvoice()})
		}
	}
	if invBase == 0 {
		invBase = payload.GetBaseValue()
	}
	if invGst == 0 {
		invGst = payload.GetGst()
	}
	if invOther == 0 {
		invOther = payload.GetOtherCharges()
	}
	if invTotal == 0 {
		invTotal = payload.GetTotalAmount()
	}

	brief := payload.GetBriefOfGoodsServices()
	if brief == "" {
		brief = payload.GetProjectName()
	}

	approvalFor := ""
	if payload.GetApprovalFor() != greennotepb.ApprovalFor_APPROVAL_FOR_UNSPECIFIED {
		approvalFor = payload.GetApprovalFor().String()
	}
	natureOfExpenses := ""
	if payload.GetNatureOfExpenses() != greennotepb.NatureOfExpenses_NATURE_OF_EXPENSES_UNSPECIFIED {
		natureOfExpenses = payload.GetNatureOfExpenses().String()
	}

	status := strings.TrimSpace(payload.GetStatus())
	if status == "" {
		status = "draft"
	}
	// Convert enums to their storage representations
	var approvalForEnum sqlcgen.NullApprovalFor
	if approvalFor != "" {
		approvalForEnum = sqlcgen.NullApprovalFor{
			ApprovalFor: sqlcgen.ApprovalFor(approvalFor),
			Valid:       true,
		}
	}

	var natureOfExpensesEnum sqlcgen.NullNatureOfExpenses
	if natureOfExpenses != "" {
		natureOfExpensesEnum = sqlcgen.NullNatureOfExpenses{
			NatureOfExpenses: sqlcgen.NatureOfExpenses(natureOfExpenses),
			Valid:            true,
		}
	}

	statusEnum := toNullStatusEnumFromString(status)

	noteRow, err := qtx.CreateGreenNote(ctx, sqlcgen.CreateGreenNoteParams{
		ID:                                uuid.NewString(),
		ProjectName:                       toNullString(payload.GetProjectName()),
		SupplierName:                      toNullString(payload.GetSupplierName()),
		ExpenseCategory:                   toNullString(payload.GetExpenseCategory()),
		ProtestNoteRaised:                 toNullYesNoEnum(payload.GetProtestNoteRaised()),
		WhetherContract:                   toNullYesNoEnum(payload.GetWhetherContract()),
		ExtensionOfContractPeriodExecuted: toNullYesNoEnum(payload.GetExtensionOfContractPeriodExecuted()),
		ExpenseAmountWithinContract:       toNullYesNoEnum(payload.GetExpenseAmountWithinContract()),
		MilestoneAchieved:                 toNullYesNoEnum(payload.GetMilestoneAchieved()),
		PaymentApprovedWithDeviation:      toNullYesNoEnum(payload.GetPaymentApprovedWithDeviation()),
		RequiredDocumentsSubmitted:        toNullYesNoEnum(payload.GetRequiredDocumentsSubmitted()),
		ContractPeriodCompleted:           toNullYesNoEnum(payload.GetContractPeriodCompleted()),
		DocumentsVerified:                 toNullString(payload.GetDocumentsVerified()),
		ContractStartDate:                 toNullString(payload.GetContractStartDate()),
		ContractEndDate:                   toNullString(payload.GetContractEndDate()),
		AppointedStartDate:                toNullString(payload.GetAppointedStartDate()),
		SupplyPeriodStart:                 toNullString(payload.GetSupplyPeriodStart()),
		SupplyPeriodEnd:                   toNullString(payload.GetSupplyPeriodEnd()),
		BaseValue:                         toNullString(formatDecimal(invBase)),
		Gst:                               toNullString(formatDecimal(invGst)),
		OtherCharges:                      toNullString(formatDecimal(invOther)),
		TotalAmount:                       toNullString(formatDecimal(invTotal)),
		EnableMultipleInvoices:            payload.GetEnableMultipleInvoices(),
		Status:                            statusEnum,
		ApprovalFor:                       approvalForEnum,
		DepartmentName:                    toNullString(payload.GetDepartmentName()),
		WorkOrderNo:                       toNullString(payload.GetWorkOrderNo()),
		PoNumber:                          toNullString(payload.GetPoNumber()),
		WorkOrderDate:                     toNullString(payload.GetWorkOrderDate()),
		ExpenseCategoryType: sqlcgen.NullExpenseCategoryType{
			ExpenseCategoryType: sqlcgen.ExpenseCategoryType(payload.GetExpenseCategoryType().String()),
			Valid:               payload.GetExpenseCategoryType() != greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_UNSPECIFIED,
		},
		MsmeClassification:             toNullString(payload.GetMsmeClassification()),
		ActivityType:                   toNullString(payload.GetActivityType()),
		BriefOfGoodsServices:           toNullString(brief),
		DelayedDamages:                 toNullString(payload.GetDelayedDamages()),
		NatureOfExpenses:               natureOfExpensesEnum,
		BudgetExpenditure:              toNullString(formatDecimal(payload.GetBudgetExpenditure())),
		ActualExpenditure:              toNullString(formatDecimal(payload.GetActualExpenditure())),
		ExpenditureOverBudget:          toNullString(formatDecimal(payload.GetExpenditureOverBudget())),
		MilestoneRemarks:               toNullString(payload.GetMilestoneRemarks()),
		SpecifyDeviation:               toNullString(payload.GetSpecifyDeviation()),
		DocumentsWorkdoneSupply:        toNullString(payload.GetDocumentsWorkdoneSupply()),
		DocumentsDiscrepancy:           toNullString(payload.GetDocumentsDiscrepancy()),
		Remarks:                        toNullString(payload.GetRemarks()),
		AuditorRemarks:                 toNullString(payload.GetAuditorRemarks()),
		AmountRetainedForNonSubmission: toNullString(formatDecimal(payload.GetAmountRetainedForNonSubmission())),
	})
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	// Persist invoice rows if provided.
	allInvoices := payload.GetInvoices()
	if len(allInvoices) == 0 && payload.GetInvoice() != nil {
		allInvoices = []*greennotepb.InvoiceInput{payload.GetInvoice()}
	}
	for _, in := range allInvoices {
		if in == nil {
			continue
		}
		_, err := qtx.InsertInvoice(ctx, sqlcgen.InsertInvoiceParams{
			ID:            uuid.NewString(),
			GreenNoteID:   noteRow.ID,
			InvoiceNumber: in.GetInvoiceNumber(),
			InvoiceDate:   toNullString(in.GetInvoiceDate()),
			TaxableValue:  formatDecimal(in.GetTaxableValue()),
			Gst:           formatDecimal(in.GetGst()),
			OtherCharges:  formatDecimal(in.GetOtherCharges()),
			InvoiceValue:  formatDecimal(in.GetInvoiceValue()),
		})
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}

	// Persist supporting documents if provided.
	if err := r.insertDocumentsTx(ctx, qtx, noteRow.ID, payload.GetNewDocuments()); err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return noteRow.ID, nil
}

func (r *Repository) Update(ctx context.Context, id string, payload *greennotepb.GreenNotePayload) error {
	if r == nil || r.db == nil {
		return nil
	}
	if payload == nil {
		return nil
	}

	if r.q == nil {
		return nil
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return ports.ErrNotFound
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	qtx := r.q.WithTx(tx)

	// Recompute invoice totals.
	invBase, invGst, invOther, invTotal := sumInvoiceInputs(payload.GetInvoices())
	if invBase == 0 && invGst == 0 && invOther == 0 && invTotal == 0 {
		if payload.GetInvoice() != nil {
			invBase, invGst, invOther, invTotal = sumInvoiceInputs([]*greennotepb.InvoiceInput{payload.GetInvoice()})
		}
	}
	if invBase == 0 {
		invBase = payload.GetBaseValue()
	}
	if invGst == 0 {
		invGst = payload.GetGst()
	}
	if invOther == 0 {
		invOther = payload.GetOtherCharges()
	}
	if invTotal == 0 {
		invTotal = payload.GetTotalAmount()
	}

	brief := payload.GetBriefOfGoodsServices()
	if brief == "" {
		brief = payload.GetProjectName()
	}

	approvalFor := ""
	if payload.GetApprovalFor() != greennotepb.ApprovalFor_APPROVAL_FOR_UNSPECIFIED {
		approvalFor = payload.GetApprovalFor().String()
	}
	natureOfExpenses := ""
	if payload.GetNatureOfExpenses() != greennotepb.NatureOfExpenses_NATURE_OF_EXPENSES_UNSPECIFIED {
		natureOfExpenses = payload.GetNatureOfExpenses().String()
	}

	status := strings.TrimSpace(payload.GetStatus())
	if status == "" {
		status = "draft"
	}
	// Convert enums to their storage representations
	var approvalForEnum sqlcgen.NullApprovalFor
	if approvalFor != "" {
		approvalForEnum = sqlcgen.NullApprovalFor{
			ApprovalFor: sqlcgen.ApprovalFor(approvalFor),
			Valid:       true,
		}
	}

	var natureOfExpensesEnum sqlcgen.NullNatureOfExpenses
	if natureOfExpenses != "" {
		natureOfExpensesEnum = sqlcgen.NullNatureOfExpenses{
			NatureOfExpenses: sqlcgen.NatureOfExpenses(natureOfExpenses),
			Valid:            true,
		}
	}

	statusEnum := toNullStatusEnumFromString(status)

	_, err = qtx.UpdateGreenNote(ctx, sqlcgen.UpdateGreenNoteParams{
		ID:                                id,
		ProjectName:                       toNullString(payload.GetProjectName()),
		SupplierName:                      toNullString(payload.GetSupplierName()),
		ExpenseCategory:                   toNullString(payload.GetExpenseCategory()),
		ProtestNoteRaised:                 toNullYesNoEnum(payload.GetProtestNoteRaised()),
		WhetherContract:                   toNullYesNoEnum(payload.GetWhetherContract()),
		ExtensionOfContractPeriodExecuted: toNullYesNoEnum(payload.GetExtensionOfContractPeriodExecuted()),
		ExpenseAmountWithinContract:       toNullYesNoEnum(payload.GetExpenseAmountWithinContract()),
		MilestoneAchieved:                 toNullYesNoEnum(payload.GetMilestoneAchieved()),
		PaymentApprovedWithDeviation:      toNullYesNoEnum(payload.GetPaymentApprovedWithDeviation()),
		RequiredDocumentsSubmitted:        toNullYesNoEnum(payload.GetRequiredDocumentsSubmitted()),
		ContractPeriodCompleted:           toNullYesNoEnum(payload.GetContractPeriodCompleted()),
		DocumentsVerified:                 toNullString(payload.GetDocumentsVerified()),
		ContractStartDate:                 toNullString(payload.GetContractStartDate()),
		ContractEndDate:                   toNullString(payload.GetContractEndDate()),
		AppointedStartDate:                toNullString(payload.GetAppointedStartDate()),
		SupplyPeriodStart:                 toNullString(payload.GetSupplyPeriodStart()),
		SupplyPeriodEnd:                   toNullString(payload.GetSupplyPeriodEnd()),
		BaseValue:                         toNullString(formatDecimal(invBase)),
		Gst:                               toNullString(formatDecimal(invGst)),
		OtherCharges:                      toNullString(formatDecimal(invOther)),
		TotalAmount:                       toNullString(formatDecimal(invTotal)),
		EnableMultipleInvoices:            payload.GetEnableMultipleInvoices(),
		Status:                            statusEnum,
		ApprovalFor:                       approvalForEnum,
		DepartmentName:                    toNullString(payload.GetDepartmentName()),
		WorkOrderNo:                       toNullString(payload.GetWorkOrderNo()),
		PoNumber:                          toNullString(payload.GetPoNumber()),
		WorkOrderDate:                     toNullString(payload.GetWorkOrderDate()),
		ExpenseCategoryType: sqlcgen.NullExpenseCategoryType{
			ExpenseCategoryType: sqlcgen.ExpenseCategoryType(payload.GetExpenseCategoryType().String()),
			Valid:               payload.GetExpenseCategoryType() != greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_UNSPECIFIED,
		},
		MsmeClassification:             toNullString(payload.GetMsmeClassification()),
		ActivityType:                   toNullString(payload.GetActivityType()),
		BriefOfGoodsServices:           toNullString(brief),
		DelayedDamages:                 toNullString(payload.GetDelayedDamages()),
		NatureOfExpenses:               natureOfExpensesEnum,
		BudgetExpenditure:              toNullString(formatDecimal(payload.GetBudgetExpenditure())),
		ActualExpenditure:              toNullString(formatDecimal(payload.GetActualExpenditure())),
		ExpenditureOverBudget:          toNullString(formatDecimal(payload.GetExpenditureOverBudget())),
		MilestoneRemarks:               toNullString(payload.GetMilestoneRemarks()),
		SpecifyDeviation:               toNullString(payload.GetSpecifyDeviation()),
		DocumentsWorkdoneSupply:        toNullString(payload.GetDocumentsWorkdoneSupply()),
		DocumentsDiscrepancy:           toNullString(payload.GetDocumentsDiscrepancy()),
		Remarks:                        toNullString(payload.GetRemarks()),
		AuditorRemarks:                 toNullString(payload.GetAuditorRemarks()),
		AmountRetainedForNonSubmission: toNullString(formatDecimal(payload.GetAmountRetainedForNonSubmission())),
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Replace invoices if provided.
	allInvoices := payload.GetInvoices()
	if len(allInvoices) == 0 && payload.GetInvoice() != nil {
		allInvoices = []*greennotepb.InvoiceInput{payload.GetInvoice()}
	}
	if len(allInvoices) > 0 {
		// Delete existing invoices for this note
		if err := qtx.DeleteInvoicesByGreenNoteID(ctx, id); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to delete existing invoices: %w", err)
		}
		for _, in := range allInvoices {
			if in == nil {
				continue
			}
			_, err := qtx.InsertInvoice(ctx, sqlcgen.InsertInvoiceParams{
				ID:            uuid.NewString(),
				GreenNoteID:   id,
				InvoiceNumber: in.GetInvoiceNumber(),
				InvoiceDate:   toNullString(in.GetInvoiceDate()),
				TaxableValue:  formatDecimal(in.GetTaxableValue()),
				Gst:           formatDecimal(in.GetGst()),
				OtherCharges:  formatDecimal(in.GetOtherCharges()),
				InvoiceValue:  formatDecimal(in.GetInvoiceValue()),
			})
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	// Append any newly uploaded documents (existing documents are left untouched).
	if err := r.insertDocumentsTx(ctx, qtx, id, payload.GetNewDocuments()); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Cancel(ctx context.Context, id string, reason string) error {
	if r == nil || r.db == nil {
		return nil
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ports.ErrNotFound
	}
	query := `
		UPDATE green_notes SET
			status = 'cancelled',
			remarks = CASE WHEN remarks IS NULL OR remarks = '' THEN $2 ELSE remarks || ' | cancel: ' || $2 END,
			updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id, reason)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ports.ErrNotFound
	}
	return nil
}

// --- helpers ---

func fromNullYesNoEnum(v sqlcgen.NullYesNoEnum) greennotepb.YesNo {
	if !v.Valid {
		return greennotepb.YesNo_YES_NO_UNSPECIFIED
	}
	switch v.YesNoEnum {
	case sqlcgen.YesNoEnumYES:
		return greennotepb.YesNo_YES
	case sqlcgen.YesNoEnumNO:
		return greennotepb.YesNo_NO
	default:
		return greennotepb.YesNo_YES_NO_UNSPECIFIED
	}
}

// insertDocumentsTx saves uploaded document binaries via DocumentStorage and
// inserts their metadata into green_note_documents within the provided
// transaction. It is used by both Create and Update to attach documents to a
// green note.
func (r *Repository) insertDocumentsTx(ctx context.Context, qtx *sqlcgen.Queries, noteID string, uploads []*greennotepb.SupportingDocumentUpload) error {
	if r == nil || r.docs == nil {
		return nil
	}
	for _, u := range uploads {
		if u == nil {
			continue
		}
		content := u.GetFileContent()
		if len(content) == 0 {
			continue
		}
		objectKey := fmt.Sprintf("note-%s/%d-%s", noteID, time.Now().UnixNano(), u.GetOriginalFilename())
		if err := r.docs.Save(ctx, objectKey, content, u.GetMimeType()); err != nil {
			return err
		}

		_, err := qtx.InsertSupportingDocument(ctx, sqlcgen.InsertSupportingDocumentParams{
			ID:               uuid.NewString(),
			GreenNoteID:      noteID,
			Name:             u.GetName(),
			OriginalFilename: u.GetOriginalFilename(),
			MimeType:         u.GetMimeType(),
			FileSize:         int64(len(content)),
			ObjectKey:        objectKey,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func toNullYesNoEnum(v greennotepb.YesNo) sqlcgen.NullYesNoEnum {
	switch v {
	case greennotepb.YesNo_YES:
		return sqlcgen.NullYesNoEnum{YesNoEnum: sqlcgen.YesNoEnumYES, Valid: true}
	case greennotepb.YesNo_NO:
		return sqlcgen.NullYesNoEnum{YesNoEnum: sqlcgen.YesNoEnumNO, Valid: true}
	default:
		return sqlcgen.NullYesNoEnum{}
	}
}

func toNullStatusEnumFromString(status string) sqlcgen.NullStatusEnum {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return sqlcgen.NullStatusEnum{StatusEnum: sqlcgen.StatusEnumSTATUSDRAFT, Valid: true}
	}
	switch s {
	case "draft", "status_draft":
		return sqlcgen.NullStatusEnum{StatusEnum: sqlcgen.StatusEnumSTATUSDRAFT, Valid: true}
	case "pending", "status_pending":
		return sqlcgen.NullStatusEnum{StatusEnum: sqlcgen.StatusEnumSTATUSPENDING, Valid: true}
	case "approved", "status_approved":
		return sqlcgen.NullStatusEnum{StatusEnum: sqlcgen.StatusEnumSTATUSAPPROVED, Valid: true}
	case "rejected", "reject", "status_rejected":
		return sqlcgen.NullStatusEnum{StatusEnum: sqlcgen.StatusEnumSTATUSREJECTED, Valid: true}
	default:
		// Fallback to draft if unknown value comes from client
		return sqlcgen.NullStatusEnum{StatusEnum: sqlcgen.StatusEnumSTATUSDRAFT, Valid: true}
	}
}

func sumInvoiceInputs(inputs []*greennotepb.InvoiceInput) (base, gst, other, total float64) {
	for _, in := range inputs {
		if in == nil {
			continue
		}
		if in.InvoiceValue == 0 {
			in.InvoiceValue = in.TaxableValue + in.Gst + in.OtherCharges
		}
		base += in.TaxableValue
		gst += in.Gst
		other += in.OtherCharges
		total += in.InvoiceValue
	}
	return
}

func formatDecimal(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func parseDecimal(s string) float64 {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
