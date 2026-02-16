package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	pkgerrors "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/middleware"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/gorilla/mux"
)

type transactionServiceInterface interface {
	CreateTransaction(ctx context.Context, userID int32, req specs.CreateTransactionRequest) (specs.CreateTransactionResponse, error)
	ProcessBulkTransactions(ctx context.Context, userID int32, reader io.Reader, filename string) (specs.BulkProcessResponse, error)
}

type repositoryInterface interface {
	GetAllTransactionsByUserID(ctx context.Context, arg repository.GetAllTransactionsByUserIDParams) ([]repository.Transaction, error)
	GetTransactionByTxnID(ctx context.Context, arg repository.GetTransactionByTxnIDParams) (repository.Transaction, error)
}

func PostTransaction(s transactionServiceInterface) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			middleware.ErrorResponse(w, http.StatusMethodNotAllowed, pkgerrors.ErrMethodNotAllowed)
			return
		}

		userID, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}

		txnReq, err := decodeCreateTransaction(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, pkgerrors.ErrInvalidBody)
			return
		}

		if err := txnReq.Validate(); err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		res, err := s.CreateTransaction(r.Context(), userID, txnReq)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		middleware.SuccessResponse(w, http.StatusOK, res)
	}
}

func GetTransactions(DB repositoryInterface) func(w http.ResponseWriter, r *http.Request) {
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

func GetTransaction(DB repositoryInterface) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		txnStrID := vars["id"]

		id, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, pkgerrors.ErrInvalidToken)
			return
		}

		txnID, err := strconv.ParseInt(txnStrID, 10, 32)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusBadRequest, pkgerrors.ErrInvalidBody)
			return
		}

		txn, err := DB.GetTransactionByTxnID(r.Context(), repository.GetTransactionByTxnIDParams{
			ID:     int32(txnID),
			UserID: id,
		})
		if err != nil {
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		middleware.SuccessResponse(w, http.StatusOK, txn)
	}
}

func ProcessBulkTransactions(s transactionServiceInterface) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			middleware.ErrorResponse(w, http.StatusMethodNotAllowed, pkgerrors.ErrMethodNotAllowed)
			return
		}

		userID, err := helpers.GetIDFromRequest(r)
		if err != nil {
			middleware.ErrorResponse(w, http.StatusUnauthorized, err)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				middleware.ErrorResponse(w, http.StatusBadRequest, pkgerrors.ErrMissingFileInRequest)
				return
			}
			middleware.ErrorResponse(w, http.StatusBadRequest, pkgerrors.ErrInvalidBody)
			return
		}
		if !strings.HasSuffix(header.Filename, ".xlsx") && !strings.HasSuffix(header.Filename, ".csv") {
			middleware.ErrorResponse(w, http.StatusBadRequest, pkgerrors.ErrInvalidBody)
			return
		}
		defer file.Close()

		res, err := s.ProcessBulkTransactions(r.Context(), userID, file, header.Filename)
		if err != nil {
			if errors.Is(err, pkgerrors.ErrUnexpectedHeadersInFile) {
				middleware.ErrorResponse(w, http.StatusBadRequest, err)
				return
			}
			middleware.ErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		middleware.SuccessResponse(w, http.StatusOK, res)
	}
}
