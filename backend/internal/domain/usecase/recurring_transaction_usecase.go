package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/google/uuid"
)

const maxProjectionYears = 50

type RecurringTransactionUsecase struct {
	recurringRepo   repository.RecurringTransactionRepository
	transactionRepo repository.TransactionRepository
}

func NewRecurringTransactionUsecase(recurringRepo repository.RecurringTransactionRepository, transactionRepo repository.TransactionRepository) *RecurringTransactionUsecase {
	return &RecurringTransactionUsecase{recurringRepo: recurringRepo, transactionRepo: transactionRepo}
}

func (uc *RecurringTransactionUsecase) Create(ctx context.Context, rt *entity.RecurringTransaction) error {
	if !isValidFrequency(rt.Frequency) {
		return domain.ErrInvalidFrequency
	}
	rt.IsActive = true

	if err := uc.recurringRepo.Create(ctx, rt); err != nil {
		return err
	}

	return uc.generateTransactions(ctx, rt, rt.StartDate)
}

func (uc *RecurringTransactionUsecase) Delete(ctx context.Context, id uuid.UUID, mode entity.DeleteMode) error {
	if _, err := uc.recurringRepo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := uc.transactionRepo.DeleteByRecurringID(ctx, id, mode); err != nil {
		return err
	}

	return uc.recurringRepo.Delete(ctx, id)
}

func (uc *RecurringTransactionUsecase) Pause(ctx context.Context, id uuid.UUID) error {
	rt, err := uc.recurringRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !rt.IsActive {
		return domain.ErrAlreadyPaused
	}

	now := time.Now()
	cutoff := computePauseCutoff(rt, now)

	if err := uc.transactionRepo.DeleteFutureByRecurringID(ctx, id, cutoff); err != nil {
		return err
	}

	return uc.recurringRepo.Pause(ctx, id, now)
}

func (uc *RecurringTransactionUsecase) Resume(ctx context.Context, id uuid.UUID, onConflict string) error {
	rt, err := uc.recurringRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if rt.IsActive {
		return domain.ErrAlreadyActive
	}

	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC)

	existing, err := uc.transactionRepo.FindByRecurringIDAndDateRange(
		ctx, id, firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"),
	)
	if err != nil {
		return err
	}

	hasConflict := len(existing) > 0

	if hasConflict && onConflict == "" {
		return &entity.ResumeConflictError{ExistingTransactions: existing}
	}

	if err := uc.recurringRepo.Resume(ctx, id); err != nil {
		return err
	}

	rt.IsActive = true
	rt.PausedAt = nil

	if hasConflict && onConflict == "update" {
		var startOrdinal int
		if rt.MaxOccurrences != nil {
			preCount, err := uc.transactionRepo.CountByRecurringIDBeforeDate(ctx, id, firstDay.Format("2006-01-02"))
			if err != nil {
				return err
			}
			startOrdinal = preCount
		}
		if err := uc.updateExistingTransactions(ctx, rt, existing, firstDay, lastDay, startOrdinal); err != nil {
			return err
		}
		// Generate future transactions from next month onward
		nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		return uc.generateTransactions(ctx, rt, nextMonth.Format("2006-01-02"))
	}

	// onConflict == "create" or no conflict: generate normally
	resumeStart := computeResumeStart(rt, now)
	return uc.generateTransactions(ctx, rt, resumeStart)
}

func (uc *RecurringTransactionUsecase) updateExistingTransactions(ctx context.Context, rt *entity.RecurringTransaction, existing []entity.Transaction, firstDay, lastDay time.Time, startOrdinal int) error {
	expectedDates := computeAllDates(rt, firstDay, lastDay)

	existingByDate := make(map[string]entity.Transaction)
	for _, tx := range existing {
		existingByDate[tx.Date] = tx
	}

	var toUpdate []entity.Transaction
	var toCreate []entity.Transaction

	for i, d := range expectedDates {
		dateStr := d.Format("2006-01-02")
		amt := rt.Amount
		desc := rt.Description
		if rt.MaxOccurrences != nil {
			amt = installmentAmount(rt.Amount, *rt.MaxOccurrences, startOrdinal+i)
			desc = installmentDescription(rt.Description, startOrdinal+i+1, *rt.MaxOccurrences)
		}
		if ex, ok := existingByDate[dateStr]; ok {
			ex.Type = rt.Type
			ex.Amount = amt
			ex.Description = desc
			ex.CategoryID = rt.CategoryID
			toUpdate = append(toUpdate, ex)
		} else {
			toCreate = append(toCreate, entity.Transaction{
				UserID:      rt.UserID,
				CategoryID:  rt.CategoryID,
				Type:        rt.Type,
				Amount:      amt,
				Description: desc,
				Date:        dateStr,
				RecurringID: &rt.ID,
			})
		}
	}

	if len(toUpdate) > 0 {
		if err := uc.transactionRepo.BulkUpdate(ctx, toUpdate); err != nil {
			return err
		}
	}
	if len(toCreate) > 0 {
		if err := uc.transactionRepo.BulkCreate(ctx, toCreate); err != nil {
			return err
		}
	}
	return nil
}

func (uc *RecurringTransactionUsecase) List(ctx context.Context, userID uuid.UUID, filter entity.RecurringTransactionFilter) (*entity.PaginatedRecurringTransactions, error) {
	return uc.recurringRepo.FindAll(ctx, userID, filter)
}

func (uc *RecurringTransactionUsecase) generateTransactions(ctx context.Context, rt *entity.RecurringTransaction, fromDateStr string) error {
	fromDate, err := time.Parse("2006-01-02", fromDateStr)
	if err != nil {
		return err
	}

	toDate := fromDate.AddDate(maxProjectionYears, 0, 0)

	// Respect end_date
	if rt.EndDate != nil {
		endDate, parseErr := time.Parse("2006-01-02", *rt.EndDate)
		if parseErr == nil && endDate.Before(toDate) {
			toDate = endDate
		}
	}

	dates := computeAllDates(rt, fromDate, toDate)

	// Respect max_occurrences
	var startOrdinal int
	if rt.MaxOccurrences != nil {
		existingCount, err := uc.transactionRepo.CountByRecurringID(ctx, rt.ID)
		if err != nil {
			return err
		}
		startOrdinal = existingCount
		remaining := *rt.MaxOccurrences - existingCount
		if remaining <= 0 {
			return nil
		}
		if len(dates) > remaining {
			dates = dates[:remaining]
		}
	}

	if len(dates) == 0 {
		return nil
	}

	txs := make([]entity.Transaction, len(dates))
	for i, d := range dates {
		amt := rt.Amount
		desc := rt.Description
		if rt.MaxOccurrences != nil {
			amt = installmentAmount(rt.Amount, *rt.MaxOccurrences, startOrdinal+i)
			desc = installmentDescription(rt.Description, startOrdinal+i+1, *rt.MaxOccurrences)
		}
		txs[i] = entity.Transaction{
			UserID:      rt.UserID,
			CategoryID:  rt.CategoryID,
			Type:        rt.Type,
			Amount:      amt,
			Description: desc,
			Date:        d.Format("2006-01-02"),
			RecurringID: &rt.ID,
		}
	}

	return uc.transactionRepo.BulkCreate(ctx, txs)
}

// installmentAmount calculates the amount for a specific installment.
// absoluteIndex is 0-based. The last installment absorbs rounding differences.
func installmentAmount(total float64, n, absoluteIndex int) float64 {
	base := math.Round((total/float64(n))*100) / 100
	if absoluteIndex == n-1 {
		return math.Round((total-base*float64(n-1))*100) / 100
	}
	return base
}

// installmentDescription returns "base - Parcela X/N".
func installmentDescription(base string, ordinal, total int) string {
	label := fmt.Sprintf("Parcela %d/%d", ordinal, total)
	if base == "" {
		return label
	}
	return fmt.Sprintf("%s - %s", base, label)
}

func computeAllDates(rt *entity.RecurringTransaction, from, to time.Time) []time.Time {
	startDate, err := time.Parse("2006-01-02", rt.StartDate)
	if err != nil {
		return nil
	}

	var dates []time.Time

	switch rt.Frequency {
	case "monthly":
		day := startDate.Day()
		if rt.DayOfMonth != nil {
			day = *rt.DayOfMonth
		}

		// Start from the month of `from`
		current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
		for !current.After(to) {
			// Clamp day to last day of this month
			lastDay := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			d := day
			if d > lastDay {
				d = lastDay
			}
			candidate := time.Date(current.Year(), current.Month(), d, 0, 0, 0, 0, time.UTC)
			if !candidate.Before(from) && !candidate.Before(startDate) && !candidate.After(to) {
				dates = append(dates, candidate)
			}
			current = current.AddDate(0, 1, 0)
		}

	case "weekly":
		current := startDate
		for current.Before(from) {
			diff := from.Sub(current)
			weeks := int(diff.Hours()/24) / 7
			if weeks > 0 {
				current = current.AddDate(0, 0, weeks*7)
			} else {
				current = current.AddDate(0, 0, 7)
			}
		}
		for !current.After(to) {
			if !current.Before(from) {
				dates = append(dates, current)
			}
			current = current.AddDate(0, 0, 7)
		}

	case "biweekly":
		current := startDate
		for current.Before(from) {
			diff := from.Sub(current)
			periods := int(diff.Hours()/24) / 14
			if periods > 0 {
				current = current.AddDate(0, 0, periods*14)
			} else {
				current = current.AddDate(0, 0, 14)
			}
		}
		for !current.After(to) {
			if !current.Before(from) {
				dates = append(dates, current)
			}
			current = current.AddDate(0, 0, 14)
		}

	case "yearly":
		startYear := from.Year()
		if from.After(time.Date(startYear, startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)) {
			startYear++
		}
		for y := startYear; ; y++ {
			candidate := time.Date(y, startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
			if candidate.After(to) {
				break
			}
			if !candidate.Before(from) && !candidate.Before(startDate) {
				dates = append(dates, candidate)
			}
		}
	}

	return dates
}

// computePauseCutoff determines the date from which to delete future transactions when pausing.
// Transactions from the 1st of next month onward are deleted.
func computePauseCutoff(rt *entity.RecurringTransaction, now time.Time) string {
	_ = rt
	firstDayNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return firstDayNextMonth.Format("2006-01-02")
}

// computeResumeStart determines from which date to start generating transactions when resuming.
// Generates from the 1st of the current month onward.
func computeResumeStart(rt *entity.RecurringTransaction, now time.Time) string {
	_ = rt
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return firstDayOfMonth.Format("2006-01-02")
}

func isValidFrequency(f string) bool {
	switch f {
	case "weekly", "biweekly", "monthly", "yearly":
		return true
	}
	return false
}
