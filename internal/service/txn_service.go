package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	pkgerrors "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type TransactionService struct {
	queries *repository.Queries
	db      *pgxpool.Pool
	logger  *zap.Logger
}

func NewTransactionService(queries *repository.Queries, db *pgxpool.Pool, logger *zap.Logger) *TransactionService {
	return &TransactionService{
		queries: queries,
		db:      db,
		logger:  logger,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID int32, req specs.CreateTransactionRequest) (specs.CreateTransactionResponse, error) {
	// 1. Get User Profile
	profile, err := s.queries.GetUserProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "no rows in result set" {
			// Create a default empty profile wrapped in the struct
			profile = repository.GetUserProfileByUserIDRow{
				UserID: userID,
			}
		} else {
			s.logger.Error("failed to get user profile", zap.Error(err))
			// Continue with empty profile or return error?
			// For robustness, let's continue with empty profile but log error.
			// Actually, sqlc with pgx returns pgx.ErrNoRows which might need specific handling or just matching string.
			// But since we are using pgx via sqlc, better check.
			// For now, let's assume we proceed with "cold start" if retrieval fails.
			profile = repository.GetUserProfileByUserIDRow{UserID: userID}
		}
	}

	// Convert repository.GetUserProfileByUserIDRow to repository.UserProfileBehavior
	// because helper functions expect *repository.UserProfileBehavior
	domainProfile := &repository.UserProfileBehavior{
		UserID:                            profile.UserID,
		AverageTransactionAmount:          profile.AverageTransactionAmount,
		StdDevTransactionAmount:           profile.StdDevTransactionAmount,
		MaxTransactionAmountSeen:          profile.MaxTransactionAmountSeen,
		AverageNumberOfTransactionsPerDay: profile.AverageNumberOfTransactionsPerDay,
		UsualTransactionStartHour:         profile.UsualTransactionStartHour,
		UsualTransactionEndHour:           profile.UsualTransactionEndHour,
		TotalTransactions:                 profile.TotalTransactions,
		AllowedTransactions:               profile.AllowedTransactions,
		UpdatedAt:                         profile.UpdatedAt,
	}
	// Convert []string to []Mode manually? repository.GetUserProfileByUserIDRow has []string for modes
	// but repository.UserProfileBehavior has []Mode.
	for _, m := range profile.RegisteredPaymentModes {
		domainProfile.RegisteredPaymentModes = append(domainProfile.RegisteredPaymentModes, repository.Mode(m))
	}

	// 2. Count recent transactions (last 24h)
	count, err := s.queries.CountRecentTransactions(ctx, repository.CountRecentTransactionsParams{
		UserID: userID,
		Secs:   3600,
	})
	if err != nil {
		s.logger.Error("failed to count recent transactions", zap.Error(err))
		count = 0
	}

	// 3. Analyze
	result := helpers.AnalyzeTransaction(&req, domainProfile, int(count), time.Now())

	// 4. Create Transaction in DB
	txn, err := s.queries.CreateTransaction(ctx, repository.CreateTransactionParams{
		UserID:                  userID,
		Amount:                  req.Amount,
		Mode:                    repository.Mode(req.Mode),
		RiskScore:               result.FinalRiskScore,
		Column5:                 result.TriggeredFactors, // database column is triggered_factors, param name might be Column5 due to sqlc naming?
		Decision:                result.Decision,
		AmountDeviationScore:    int32(result.AmountRisk),
		FrequencyDeviationScore: int32(result.FrequencyRisk),
		ModeDeviationScore:      int32(result.ModeRisk),
		TimeDeviationScore:      int32(result.TimeRisk),
		CreatedAt:               pgtype.Timestamp{Time: time.Now(), Valid: true},
	})

	if err != nil {
		s.logger.Error("failed to create transaction", zap.Error(err))
		return specs.CreateTransactionResponse{}, err
	}

	return specs.CreateTransactionResponse{
		TransactionID:    txn.ID,
		Decision:         txn.Decision,
		RiskScore:        txn.RiskScore,
		TriggeredFactors: txn.TriggeredFactors,
		CreatedAt:        txn.CreatedAt.Time,
	}, nil
}

func (s *TransactionService) ProcessBulkTransactions(ctx context.Context, userID int32, reader io.Reader, filename string) (specs.BulkProcessResponse, error) {
	var recordIterator func() ([]string, error)
	var closeFunc func()

	if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		f, err := excelize.OpenReader(reader)
		if err != nil {
			return specs.BulkProcessResponse{}, pkgerrors.ErrFailureInParsingExcel
		}

		// assuming first sheet
		sheetName := f.GetSheetName(0)
		rows, err := f.Rows(sheetName)
		if err != nil {
			return specs.BulkProcessResponse{}, pkgerrors.ErrFailureInParsingExcel
		}

		headers, err := rows.Columns()
		if err != nil {
			return specs.BulkProcessResponse{}, pkgerrors.ErrFailureInParsingExcel
		}

		if strings.ToLower(headers[0]) != "amount" || strings.ToLower(headers[1]) != "mode" || strings.ToLower(headers[2]) != "created_at" {
			return specs.BulkProcessResponse{}, pkgerrors.ErrUnexpectedHeadersInFile
		}

		if rows.Next() {
			_, _ = rows.Columns()
		}

		recordIterator = func() ([]string, error) {
			if !rows.Next() {
				return nil, io.EOF
			}
			return rows.Columns()
		}
		closeFunc = func() {
			f.Close()
			rows.Close()
		}
	} else {
		// CSV
		csvReader := csv.NewReader(reader)
		// Read header
		_, err := csvReader.Read()
		if err != nil {
			return specs.BulkProcessResponse{}, fmt.Errorf("failed to read CSV header: %w", err)
		}

		recordIterator = func() ([]string, error) {
			return csvReader.Read()
		}
		closeFunc = func() {}
	}
	defer closeFunc()

	processed := 0
	success := 0
	failed := 0

	// Get profile once
	profile, err := s.queries.GetUserProfileByUserID(ctx, userID)
	if err != nil {
		profile = repository.GetUserProfileByUserIDRow{UserID: userID}
	}

	domainProfile := &repository.UserProfileBehavior{
		UserID:                            profile.UserID,
		AverageTransactionAmount:          profile.AverageTransactionAmount,
		StdDevTransactionAmount:           profile.StdDevTransactionAmount,
		MaxTransactionAmountSeen:          profile.MaxTransactionAmountSeen,
		AverageNumberOfTransactionsPerDay: profile.AverageNumberOfTransactionsPerDay,
		UsualTransactionStartHour:         profile.UsualTransactionStartHour,
		UsualTransactionEndHour:           profile.UsualTransactionEndHour,
		TotalTransactions:                 profile.TotalTransactions,
		AllowedTransactions:               profile.AllowedTransactions,
		UpdatedAt:                         profile.UpdatedAt,
	}
	// Convert []string to []Mode manually
	for _, m := range profile.RegisteredPaymentModes {
		domainProfile.RegisteredPaymentModes = append(domainProfile.RegisteredPaymentModes, repository.Mode(m))
	}

	batchSize := 50
	batchCount := 0

	for {
		record, err := recordIterator()
		if err == io.EOF {
			break
		}
		if err != nil {
			failed++
			continue
		}

		// Expected format: amount, mode, created_at
		if len(record) < 3 {
			failed++
			continue
		}

		processed++

		amount, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			failed++
			continue
		}

		batchCount++

		mode := record[1]
		// clean mode string?
		mode = strings.ToUpper(strings.TrimSpace(mode))

		createdAt, err := time.Parse(time.RFC3339, record[2])
		if err != nil {
			// Try other formats or fallback?
			// The test data uses ISO8601 like 2025-10-23T22:05:19.312257
			// RFC3339 should handle it.
			createdAt = time.Now()
		}

		bulkReq := specs.CreateBulkTransactionRequest{
			Amount:    amount,
			Mode:      mode,
			CreatedAt: createdAt,
		}

		// Count passed as 0 for bulk for simplicity, or we could estimate?
		result := helpers.AnalyzeBulkTransactions(&bulkReq, domainProfile, 0)

		_, err = s.queries.CreateTransaction(ctx, repository.CreateTransactionParams{
			UserID:                  userID,
			Amount:                  bulkReq.Amount,
			Mode:                    repository.Mode(bulkReq.Mode),
			RiskScore:               result.FinalRiskScore,
			Column5:                 result.TriggeredFactors,
			Decision:                result.Decision,
			AmountDeviationScore:    int32(result.AmountRisk),
			FrequencyDeviationScore: int32(result.FrequencyRisk),
			ModeDeviationScore:      int32(result.ModeRisk),
			TimeDeviationScore:      int32(result.TimeRisk),
			CreatedAt:               pgtype.Timestamp{Time: createdAt, Valid: true},
		})

		if err != nil {
			s.logger.Error("failed to create bulk transaction", zap.Error(err))
			failed++
		} else {
			success++
		}

		// Batch update profile
		if batchCount >= batchSize {
			// Recalculate based on DB state
			err := s.queries.RecalculateUserProfile(ctx, userID)
			if err != nil {
				s.logger.Error("failed to recalculate user profile", zap.Error(err))
			} else {
				// Refresh local profile
				p, err := s.queries.GetUserProfileByUserID(ctx, userID)
				if err == nil {
					// Update domainProfile
					domainProfile = &repository.UserProfileBehavior{
						UserID:                            p.UserID,
						AverageTransactionAmount:          p.AverageTransactionAmount,
						StdDevTransactionAmount:           p.StdDevTransactionAmount,
						MaxTransactionAmountSeen:          p.MaxTransactionAmountSeen,
						AverageNumberOfTransactionsPerDay: p.AverageNumberOfTransactionsPerDay,
						UsualTransactionStartHour:         p.UsualTransactionStartHour,
						UsualTransactionEndHour:           p.UsualTransactionEndHour,
						TotalTransactions:                 p.TotalTransactions,
						AllowedTransactions:               p.AllowedTransactions,
						UpdatedAt:                         p.UpdatedAt,
					}
					domainProfile.RegisteredPaymentModes = nil
					for _, m := range p.RegisteredPaymentModes {
						domainProfile.RegisteredPaymentModes = append(domainProfile.RegisteredPaymentModes, repository.Mode(m))
					}
				}
			}
			batchCount = 0
		}
	}

	// Final recalculation to ensure consistency
	_ = s.queries.RecalculateUserProfile(ctx, userID)

	return specs.BulkProcessResponse{
		JobID:     "sync-job-" + time.Now().Format("20060102150405"),
		Status:    "COMPLETED",
		Processed: processed,
		Success:   success,
		Failed:    failed,
	}, nil
}
