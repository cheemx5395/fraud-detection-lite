package helpers

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

func ParseTransactionCSVRow(record []string) (specs.CreateBulkTransactionRequest, error) {
	if len(record) < 3 {
		return specs.CreateBulkTransactionRequest{}, errors.ErrInvalidBody
	}

	amount, err := strconv.ParseFloat(record[0], 64)
	if err != nil || amount <= 0 {
		return specs.CreateBulkTransactionRequest{}, errors.ErrInvalidBody
	}

	mode := record[1]

	createdAt, err := time.Parse(time.RFC3339, record[2])
	if err != nil {
		createdAt = time.Now()
	}

	return specs.CreateBulkTransactionRequest{
		Amount:    int(amount),
		Mode:      mode,
		CreatedAt: createdAt,
	}, nil
}

func NewEmptyUserProfile(userID int32) *repository.UserProfileBehavior {
	return &repository.UserProfileBehavior{
		UserID:                            userID,
		AverageTransactionAmount:          pgtype.Float8{Float64: 0, Valid: true},
		AverageNumberOfTransactionsPerDay: pgtype.Int4{Int32: 0, Valid: true},
		MaxTransactionAmountSeen:          pgtype.Float8{Float64: 0, Valid: true},
		RegisteredPaymentModes:            []repository.Mode{},
		TotalTransactions:                 0,
		AllowedTransactions:               0,
	}
}

func MapDBProfileToDomain(p repository.GetUserProfileByUserIDRow) *repository.UserProfileBehavior {
	return &repository.UserProfileBehavior{
		UserID:                            p.UserID,
		AverageTransactionAmount:          p.AverageTransactionAmount,
		AverageNumberOfTransactionsPerDay: p.AverageNumberOfTransactionsPerDay,
		MaxTransactionAmountSeen:          p.MaxTransactionAmountSeen,
		RegisteredPaymentModes:            GetModeSliceFromStringSlice(p.RegisteredPaymentModes),
		UsualTransactionStartHour:         p.UsualTransactionStartHour,
		UsualTransactionEndHour:           p.UsualTransactionEndHour,
		TotalTransactions:                 p.TotalTransactions,
		AllowedTransactions:               p.AllowedTransactions,
		UpdatedAt:                         p.UpdatedAt,
	}
}

func ProcessBulkTransactionJob(
	ctx context.Context,
	DB *repository.Queries,
	RD *redis.Client,
	jobID string,
	userID int32,
	reader io.Reader,
) {
	jobKey := "bulk_txn_job:" + jobID
	RD.HSet(ctx, jobKey, "status", "RUNNING")

	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	if _, err := csvReader.Read(); err != nil {
		RD.HSet(ctx, jobKey, "status", "FAILED")
		return
	}

	var profile *repository.UserProfileBehavior

	dbProfile, err := DB.GetUserProfileByUserID(ctx, userID)
	if err != nil {
		profile = NewEmptyUserProfile(userID)
	} else {
		profile = MapDBProfileToDomain(dbProfile)
	}

	processedCount := 0

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			RD.HIncrBy(ctx, jobKey, "failed", 1)
			continue
		}

		txnReq, err := ParseTransactionCSVRow(record)
		if err != nil {
			RD.HIncrBy(ctx, jobKey, "failed", 1)
			continue
		}

		recentCount, _ := DB.CountRecentTransactions(
			ctx,
			repository.CountRecentTransactionsParams{
				UserID: userID,
				Secs:   constants.FrequencyWindowHours.Seconds(),
			},
		)

		analysis := AnalyzeBulkTransactions(
			&txnReq,
			profile, // ← evolving profile
			int(recentCount),
		)

		_, err = DB.CreateTransaction(ctx, repository.CreateTransactionParams{
			UserID:                  userID,
			Amount:                  float64(txnReq.Amount),
			Mode:                    repository.Mode(txnReq.Mode),
			RiskScore:               analysis.FinalRiskScore,
			Column5:                 analysis.TriggeredFactors,
			Decision:                analysis.Decision,
			AmountDeviationScore:    int32(analysis.AmountRisk),
			FrequencyDeviationScore: int32(analysis.FrequencyRisk),
			ModeDeviationScore:      int32(analysis.ModeRisk),
			TimeDeviationScore:      int32(analysis.TimeRisk),
			CreatedAt:               pgtype.Timestamp{Time: txnReq.CreatedAt, Valid: true},
		})

		if err != nil {
			RD.HIncrBy(ctx, jobKey, "failed", 1)
			continue
		}

		if analysis.Decision != repository.TransactionDecisionBLOCK {
			ApplyTransactionToProfile(
				profile,
				float64(txnReq.Amount),
				repository.Mode(txnReq.Mode),
				txnReq.CreatedAt,
				analysis.Decision,
			)
		}

		processedCount++
		RD.HIncrBy(ctx, jobKey, "success", 1)
		RD.HIncrBy(ctx, jobKey, "processed", 1)

		if processedCount%10 == 0 {
			_ = DB.UpsertUserProfileFromProfile(ctx, repository.UpsertUserProfileFromProfileParams{
				UserID:                            profile.UserID,
				AverageTransactionAmount:          profile.AverageTransactionAmount,
				MaxTransactionAmountSeen:          profile.MaxTransactionAmountSeen,
				AverageNumberOfTransactionsPerDay: profile.AverageNumberOfTransactionsPerDay,
				RegisteredPaymentModes:            profile.RegisteredPaymentModes,
				UsualTransactionStartHour:         profile.UsualTransactionStartHour,
				UsualTransactionEndHour:           profile.UsualTransactionEndHour,
				TotalTransactions:                 profile.TotalTransactions,
				AllowedTransactions:               profile.AllowedTransactions,
			})
		}
	}

	_ = DB.UpsertUserProfileFromProfile(ctx, repository.UpsertUserProfileFromProfileParams{
		UserID:                            profile.UserID,
		AverageTransactionAmount:          profile.AverageTransactionAmount,
		MaxTransactionAmountSeen:          profile.MaxTransactionAmountSeen,
		AverageNumberOfTransactionsPerDay: profile.AverageNumberOfTransactionsPerDay,
		RegisteredPaymentModes:            profile.RegisteredPaymentModes,
		UsualTransactionStartHour:         profile.UsualTransactionStartHour,
		UsualTransactionEndHour:           profile.UsualTransactionEndHour,
		TotalTransactions:                 profile.TotalTransactions,
		AllowedTransactions:               profile.AllowedTransactions,
	})

	RD.HSet(ctx, jobKey, "status", "COMPLETED")
}

func CountCSVRows(bReader *bytes.Reader) (int, error) {
	reader := csv.NewReader(bReader)
	count := 0
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func CalculateProgressPercent(processed, total int) int {
	if total == 0 {
		return 0
	}
	return int((float64(processed) / float64(total)) * 100)
}
