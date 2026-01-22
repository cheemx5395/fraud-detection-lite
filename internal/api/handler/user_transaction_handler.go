package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	error "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
)

func PostTransaction(DB *repository.Queries) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}

		txnParams, err := decodeCreateTransaction(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, error.ErrInvalidBody)
			return
		}

		if txnParams.Amount <= 0 {
			middleware.ErrorResponse(w, http.StatusForbidden, error.ErrTxnBlocked)
			return
		}

		profile, err := DB.GetUserProfileByUserID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.ErrorResponse(w, http.StatusBadRequest, error.ErrUserNotFound)
				return
			}
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrDB)
			return
		}

		recentCount, err := DB.CountRecentTransactions(r.Context(), repository.CountRecentTransactionsParams{
			UserID: userID,
			Secs:   constants.FrequencyWindowHours.Seconds(),
		})
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrDB)
			return
		}

		analysis := helpers.AnalyzeTransaction(
			&txnParams,
			&profile,
			int(recentCount),
			time.Now(),
		)

		txn, err := DB.CreateTransaction(r.Context(), repository.CreateTransactionParams{
			UserID:           userID,
			Amount:           int32(txnParams.Amount),
			Type:             repository.TransactionType(txnParams.Type),
			Mode:             repository.Mode(txnParams.Mode),
			RiskScore:        analysis.FinalRiskScore,
			TriggeredFactors: analysis.TriggeredFactors,
			Decision:         analysis.Decision,
		})
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrDB)
			return
		}

		if analysis.Decision == repository.TransactionDecisionALLOW || analysis.Decision == repository.TransactionDecisionFLAG {
			_ = DB.RebuildUserProfileByID(r.Context(), userID)
		}

		middleware.SuccessResponse(w, http.StatusOK, specs.CreateTransactionResponse{
			TransactionID:    txn.ID,
			Decision:         analysis.Decision,
			RiskScore:        analysis.FinalRiskScore,
			TriggeredFactors: analysis.TriggeredFactors,
		})

	}
}

func GetTransactions(DB *repository.Queries) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}
		txns, err := DB.GetAllTransactionsByUserID(r.Context(), id)
		middleware.SuccessResponse(w, http.StatusOK, txns)
	}
}
