package handler

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	error "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
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
			middleware.ErrorResponse(w, http.StatusBadRequest, error.ErrInvalidBody)
			return
		}

		if txnParams.Amount <= 0 {
			middleware.ErrorResponse(w, http.StatusForbidden, error.ErrTxnBlocked)
			return
		}

		var profile *repository.UserProfileBehavior

		userProfile, err := DB.GetUserProfileByUserID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// New user → empty profile
				profile = &repository.UserProfileBehavior{
					UserID:                            userID,
					AverageTransactionAmount:          pgtype.Float8{Float64: 0, Valid: true},
					AverageNumberOfTransactionsPerDay: pgtype.Int4{Int32: 0, Valid: true},
					MaxTransactionAmountSeen:          pgtype.Float8{Float64: 0, Valid: true},
					RegisteredPaymentModes:            []repository.Mode{},
					TotalTransactions:                 0,
					AllowedTransactions:               0,
				}
			} else {
				middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrDB)
				return
			}
		} else {
			// Existing user
			profile = &repository.UserProfileBehavior{
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
		}

		recentCount, err := DB.CountRecentTransactions(
			r.Context(),
			repository.CountRecentTransactionsParams{
				UserID: userID,
				Secs:   constants.FrequencyWindowHours.Seconds(),
			},
		)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrDB)
			return
		}

		analysis := helpers.AnalyzeTransaction(
			&txnParams,
			profile,
			int(recentCount),
			time.Now(),
		)

		txn, err := DB.CreateTransaction(r.Context(), repository.CreateTransactionParams{
			UserID:                  userID,
			Amount:                  float64(txnParams.Amount),
			Mode:                    repository.Mode(txnParams.Mode),
			RiskScore:               analysis.FinalRiskScore,
			Column5:                 analysis.TriggeredFactors,
			Decision:                analysis.Decision,
			AmountDeviationScore:    int32(analysis.AmountRisk),
			FrequencyDeviationScore: int32(analysis.FrequencyRisk),
			ModeDeviationScore:      int32(analysis.ModeRisk),
			TimeDeviationScore:      int32(analysis.TimeRisk),
			CreatedAt:               pgtype.Timestamp{Time: time.Now(), Valid: true},
		})
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		if analysis.Decision != repository.TransactionDecisionBLOCK {
			_ = DB.UpsertUserProfileByUserID(r.Context(), userID)
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
		limit := constants.DefaultTransactionsLimit
		offset := constants.DefaultTransactionsOffset

		q := r.URL.Query()

		if l := q.Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		if o := q.Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		id, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}
		txns, err := DB.GetAllTransactionsByUserID(r.Context(), repository.GetAllTransactionsByUserIDParams{
			UserID: id,
			Limit:  int32(limit),
			Offset: int32(offset),
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
				CreatedAt:        txn.CreatedAt.Time,
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

func ProcessBulkTransactions(DB *repository.Queries, RD *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, error.ErrInvalidBody)
			return
		}
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrInternalService)
			return
		}

		totalRows, err := helpers.CountCSVRows(bytes.NewReader(buf))
		if err != nil || totalRows == 0 {
			middleware.ErrorResponse(w, http.StatusBadRequest, error.ErrInvalidBody)
			return
		}

		jobID := uuid.NewString()
		jobKey := "bulk_txn_job:" + jobID

		RD.HSet(r.Context(), jobKey, map[string]interface{}{
			"user_id":   userID,
			"status":    "PENDING",
			"total":     totalRows,
			"processed": 0,
			"success":   0,
			"failed":    0,
		})
		RD.Expire(r.Context(), jobKey, 24*time.Hour)

		go helpers.ProcessBulkTransactionJob(
			context.Background(),
			DB,
			RD,
			jobID,
			userID,
			bytes.NewReader(buf),
		)

		middleware.SuccessResponse(w, http.StatusAccepted, map[string]string{
			"job_id": jobID,
			"status": "PENDING",
		})
	}
}

func TrackProcessProgress(RD *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		jobID := vars["job_id"]

		if jobID == "" {
			middleware.ErrorResponse(w, http.StatusBadRequest, error.ErrInvalidBody)
			return
		}

		jobKey := "bulk_txn_job:" + jobID

		exists, err := RD.Exists(r.Context(), jobKey).Result()
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrInternalService)
			return
		}
		if exists == 0 {
			middleware.ErrorResponse(w, http.StatusNotFound, error.ErrNotFound)
			return
		}

		data, err := RD.HGetAll(r.Context(), jobKey).Result()
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, error.ErrInternalService)
			return
		}

		total, _ := strconv.Atoi(data["total"])
		processed, _ := strconv.Atoi(data["processed"])
		success, _ := strconv.Atoi(data["success"])
		failed, _ := strconv.Atoi(data["failed"])

		response := map[string]any{
			"job_id": jobID,
			"status": data["status"],
			"progress": map[string]any{
				"total":     total,
				"processed": processed,
				"success":   success,
				"failed":    failed,
				"percent":   helpers.CalculateProgressPercent(processed, total),
			},
		}

		middleware.SuccessResponse(w, http.StatusOK, response)
	}
}
