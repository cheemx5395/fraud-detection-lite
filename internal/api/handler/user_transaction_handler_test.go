package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	pkgerrors "github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/helpers"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostTransaction(t *testing.T) {
	t.Run("invalid request method", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := PostTransaction(mockService)

		req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("successful transaction", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := PostTransaction(mockService)
		os.Setenv("JWT_SECRET", "testsecret")

		txnReq := specs.CreateTransactionRequest{
			Amount: 1000,
			Mode:   "UPI",
		}
		reqBody, _ := json.Marshal(txnReq)

		token, _ := helpers.MakeJWT(1, "Test User", "test@example.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodPost, "/api/transaction", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		mockService.On("CreateTransaction", mock.Anything, int32(1), txnReq).Return(
			specs.CreateTransactionResponse{
				TransactionID:    1,
				Decision:         "Allow",
				RiskScore:        55,
				TriggeredFactors: []string{string(repository.TriggerFactorsNEWMODE)},
				CreatedAt:        time.Now(),
			},
			nil,
		).Once()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]any)
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, "Allow", data["decision"])
		assert.Equal(t, float64(55), data["risk_score"])
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized - no token", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := PostTransaction(mockService)

		req := httptest.NewRequest(http.MethodPost, "/api/transactions", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid request body: invalid json", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := PostTransaction(mockService)
		os.Setenv("JWT_SECRET", "testsecret")

		token, _ := helpers.MakeJWT(2, "Test User", "test-2@example.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodPost, "/api/transaction", bytes.NewBufferString("invalid request"))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler(w, req)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, pkgerrors.ErrInvalidBody.Error(), response["error_message"])
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid request body: missing amount", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := PostTransaction(mockService)
		os.Setenv("JWT_SECRET", "testsecret")

		txnReq := specs.CreateTransactionRequest{
			Mode: "CARD",
		}
		reqBody, _ := json.Marshal(txnReq)

		token, _ := helpers.MakeJWT(2, "Test User", "test-2@example.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodPost, "/api/transaction", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler(w, req)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, pkgerrors.ErrMissingAmountInRequest.Error(), response["error_message"])
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestProcessBulkTransactions(t *testing.T) {
	t.Run("invalid request: not allowed method", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := ProcessBulkTransactions(mockService)

		req := httptest.NewRequest(http.MethodGet, "/api/transactions/upload", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request: unauthorized", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := ProcessBulkTransactions(mockService)

		req := httptest.NewRequest(http.MethodPost, "/api/transactions/upload", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid request: missing file in request", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := ProcessBulkTransactions(mockService)
		os.Setenv("JWT_SECRET", "testsecret")

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		writer.Close()

		token, _ := helpers.MakeJWT(2, "Test User", "test-2@example.com", "testsecret", time.Hour)
		req := httptest.NewRequest(http.MethodPost, "/api/transactions/upload", &body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, pkgerrors.ErrMissingFileInRequest.Error(), response["error_message"])

		mockService.AssertNotCalled(t, "ProcessBulkTransactions", mock.Anything)
	})

	t.Run("invalid request: unexpected file format", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := ProcessBulkTransactions(mockService)
		os.Setenv("JWT_SECRET", "testsecret")

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		fileWriter, _ := writer.CreateFormFile("file", "test.txt")
		fileWriter.Write([]byte("amount,mode\n1000,UPI"))

		writer.Close()

		token, _ := helpers.MakeJWT(1, "Test User", "test@example.com", "testsecret", time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/api/transactions/upload", &body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, pkgerrors.ErrInvalidBody.Error(), response["error_message"])

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request: unexpected header fields", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := ProcessBulkTransactions(mockService)
		os.Setenv("JWT_SECRET", "testsecret")

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		fileWriter, _ := writer.CreateFormFile("file", "test.csv")
		fileWriter.Write([]byte("amount,mode\n1000,UPI"))

		writer.Close()

		token, _ := helpers.MakeJWT(1, "Test User", "test@example.com", "testsecret", time.Hour)

		req := httptest.NewRequest(http.MethodPost, "/api/transactions/upload", &body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, pkgerrors.ErrInvalidBody.Error(), response["error_message"])

		mockService.AssertExpectations(t)
	})
}
