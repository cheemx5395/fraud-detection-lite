package helpers

import (
	"slices"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
)

// applyTransactionToProfile mutates the in-memory profile after a transaction
func ApplyTransactionToProfile(
	profile *repository.UserProfileBehavior,
	amount float64,
	mode repository.Mode,
	createdAt time.Time,
	decision repository.TransactionDecision,
) {
	profile.TotalTransactions++

	if decision != repository.TransactionDecisionALLOW &&
		decision != repository.TransactionDecisionFLAG {
		return
	}

	profile.AllowedTransactions++

	if profile.AllowedTransactions == 1 {
		profile.AverageTransactionAmount.Float64 = amount
		profile.AverageTransactionAmount.Valid = true
	} else {
		prevAvg := profile.AverageTransactionAmount.Float64
		n := float64(profile.AllowedTransactions)

		profile.AverageTransactionAmount.Float64 =
			((prevAvg * (n - 1)) + amount) / n
	}

	if !profile.MaxTransactionAmountSeen.Valid ||
		amount > profile.MaxTransactionAmountSeen.Float64 {
		profile.MaxTransactionAmountSeen.Float64 = amount
		profile.MaxTransactionAmountSeen.Valid = true
	}

	if !slices.Contains(profile.RegisteredPaymentModes, mode) {
		profile.RegisteredPaymentModes = append(
			profile.RegisteredPaymentModes,
			mode,
		)
	}

	hour := createdAt.Hour()

	if !profile.UsualTransactionStartHour.Valid ||
		hour < profile.UsualTransactionStartHour.Time.Hour() {
		profile.UsualTransactionStartHour.Time =
			time.Date(0, 1, 1, hour, 0, 0, 0, time.UTC)
		profile.UsualTransactionStartHour.Valid = true
	}

	if !profile.UsualTransactionEndHour.Valid ||
		hour > profile.UsualTransactionEndHour.Time.Hour() {
		profile.UsualTransactionEndHour.Time =
			time.Date(0, 1, 1, hour, 0, 0, 0, time.UTC)
		profile.UsualTransactionEndHour.Valid = true
	}
}
