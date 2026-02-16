package handler

import (
	"context"
	"io"
	"log"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) CreateTransaction(ctx context.Context, userID int32, req specs.CreateTransactionRequest) (specs.CreateTransactionResponse, error) {
	args := m.Called(ctx, userID, req)
	log.Println(args...)
	return args.Get(0).(specs.CreateTransactionResponse), args.Error(1)
}

func (m *MockTransactionService) ProcessBulkTransactions(ctx context.Context, userID int32, reader io.Reader, filename string) (specs.BulkProcessResponse, error) {
	args := m.Called(ctx, userID, reader, filename)
	log.Println(args...)
	return args.Get(0).(specs.BulkProcessResponse), args.Error(1)
}
