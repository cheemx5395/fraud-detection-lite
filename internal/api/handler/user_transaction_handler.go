package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	error "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/gorilla/mux"
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

		userProfile, err := DB.GetUserProfileByUserID(r.Context(), userID)
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

		profile := &repository.UserProfileBehavior{
			UserID:                            userProfile.UserID,
			AverageTransactionAmount:          userProfile.AverageTransactionAmount,
			AverageNumberOfTransactionsPerDay: userProfile.AverageNumberOfTransactionsPerDay,
			MaxTransactionAmountSeen:          userProfile.MaxTransactionAmountSeen,
			RegisteredPaymentModes:            helpers.GetModeSliceFromStringSlice(userProfile.RegisteredPaymentModes),
			UsualTransactionStartHour:         userProfile.UsualTransactionStartHour,
			UsualTransactionEndHour:           userProfile.UsualTransactionEndHour,
			TotalTransactions:                 userProfile.TotalTransactions,
			AllowedTransactions:               userProfile.AllowedTransactions,
			UpdatedAt:                         userProfile.UpdatedAt,
		}

		analysis := helpers.AnalyzeTransaction(
			&txnParams,
			profile,
			int(recentCount),
			time.Now(),
		)

		txn, err := DB.CreateTransaction(r.Context(), repository.CreateTransactionParams{
			UserID:                  userID,
			Amount:                  int32(txnParams.Amount),
			Mode:                    repository.Mode(txnParams.Mode),
			RiskScore:               analysis.FinalRiskScore,
			Column5:                 analysis.TriggeredFactors,
			Decision:                analysis.Decision,
			AmountDeviationScore:    int32(analysis.AmountRisk),
			FrequencyDeviationScore: int32(analysis.FrequencyRisk),
			ModeDeviationScore:      int32(analysis.ModeRisk),
			TimeDeviationScore:      int32(analysis.TimeRisk),
		})
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
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
		txns, err := DB.GetAllTransactionsByUserID(r.Context(), repository.GetAllTransactionsByUserIDParams{
			UserID: id,
			Limit:  20,
			Offset: 0,
		})
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}
		var res []specs.CreateTransactionResponse
		for _, txn := range txns {
			resTxn := specs.CreateTransactionResponse{
				TransactionID:    txn.ID,
				Decision:         txn.Decision,
				RiskScore:        txn.RiskScore,
				TriggeredFactors: txn.TriggeredFactors,
			}
			res = append(res, resTxn)
		}
		middleware.SuccessResponse(w, http.StatusOK, res)
	}
}

func GetTransaction(DB *repository.Queries) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		txnStrID := vars["id"]

		id, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, error.ErrInvalidToken)
			return
		}

		txnID, err := strconv.ParseInt(txnStrID, 10, 32)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrInternalService)
			return
		}

		txn, err := DB.GetTransactionByTxnID(r.Context(), repository.GetTransactionByTxnIDParams{
			ID:     int32(txnID),
			UserID: id,
		})

		middleware.SuccessResponse(w, http.StatusOK, txn)
	}
}
