package helpers

import (
	"bytes"
	"encoding/csv"
	"io"
	"strconv"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
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
		Amount:    amount,
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
