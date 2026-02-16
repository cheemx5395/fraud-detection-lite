package service_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
	"github.com/cheemx5395/fraud-detection-lite/internal/service"
	"github.com/cheemx5395/fraud-detection-lite/internal/worker"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestServices(t *testing.T) (*service.UserService, *service.TransactionService, *repository.Queries) {
	logger, _ := zap.NewDevelopment()

	_ = os.Setenv("DB_URI", "postgres://fraud_db:fraud_db@localhost:5433/fraud_db?sslmode=disable")
	_ = os.Setenv("REDIS_ADDR", "localhost:6379")

	// DB connection
	queries, pool, err := repository.InitializeDatabase(context.Background())
	require.NoError(t, err)

	// Redis connection
	redisClient := worker.InitializeRedis()

	userService := service.NewUserService(queries, redisClient, logger)
	txnService := service.NewTransactionService(queries, pool, logger)

	return userService, txnService, queries
}

func TestUserFlow(t *testing.T) {
	userService, _, _ := setupTestServices(t)
	ctx := context.Background()

	// Unique email
	email := "testuser_" + time.Now().Format("20060102150405") + "@example.com"

	// 1. Signup
	signupReq := specs.UserSignupRequest{
		Name:     "Test User",
		Email:    email,
		Password: "password123",
	}
	signupRes, err := userService.Signup(ctx, signupReq)
	require.NoError(t, err)
	assert.NotEmpty(t, signupRes.ID)
	assert.Equal(t, email, signupRes.Email)

	// 2. Login
	loginReq := specs.UserLoginRequest{
		Email:    email,
		Password: "password123",
	}
	loginRes, err := userService.Login(ctx, loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, loginRes.Token)

	// 3. Logout (simulated context with extracted claims)
	claims := &specs.UserTokenClaims{
		UserID: signupRes.ID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			ID:        "test-jti",
		},
	}

	// We didn't really parse the token, so we manually constructed claims.
	// The blacklist key will be "blacklist:test-jti".
	err = userService.Logout(ctx, claims)
	require.NoError(t, err)
}

func TestTransactionFlow(t *testing.T) {
	userService, txnService, _ := setupTestServices(t)
	ctx := context.Background()

	// Setup User
	email := "txnuser_" + time.Now().Format("20060102150405") + "@example.com"
	signupReq := specs.UserSignupRequest{
		Name:     "Txn User",
		Email:    email,
		Password: "password123",
	}
	signupRes, err := userService.Signup(ctx, signupReq)
	require.NoError(t, err)

	// 1. Create Transaction
	txnReq := specs.CreateTransactionRequest{
		Amount: 500.0,
		Mode:   "UPI",
	}
	txnRes, err := txnService.CreateTransaction(ctx, signupRes.ID, txnReq)
	require.NoError(t, err)
	assert.Equal(t, repository.TransactionDecisionALLOW, txnRes.Decision)
	assert.NotZero(t, txnRes.TransactionID)

	// 2. Bulk Transaction
	csvContent := `amount,mode,created_at
500.0,UPI,2023-10-01T10:00:00Z
1000.0,CARD,2023-10-01T11:00:00Z
invalid,UPI,2023-10-01T12:00:00Z`

	reader := strings.NewReader(csvContent)
	bulkRes, err := txnService.ProcessBulkTransactions(ctx, signupRes.ID, reader, "test.csv")
	require.NoError(t, err)
	assert.Equal(t, 2, bulkRes.Success) // 2 valid rows
	assert.Equal(t, 1, bulkRes.Failed)  // 1 invalid amount
	assert.Equal(t, 3, bulkRes.Processed)

	// 3. Excel Bulk Transaction
	// Assuming test is run from project root or we can find the file.
	// Try opening from root
	f, err := os.Open("../../test_data_bulk_large.xlsx")
	if err != nil {
		t.Log("Skipping Excel test because file not found:", err)
	} else {
		defer f.Close()
		excelRes, err := txnService.ProcessBulkTransactions(ctx, signupRes.ID, f, "test.xlsx")
		require.NoError(t, err)
		assert.Greater(t, excelRes.Processed, 0, "Should process Excel rows")
		t.Logf("Processed Excel rows: %d", excelRes.Processed)
	}
}

func TestBatchProfileUpdate(t *testing.T) {
	userService, txnService, queries := setupTestServices(t)
	ctx := context.Background()

	// Setup User
	email := "batchuser_" + time.Now().Format("20060102150405") + "@example.com"
	signupReq := specs.UserSignupRequest{
		Name:     "Batch Test User",
		Email:    email,
		Password: "password123",
	}
	signupRes, err := userService.Signup(ctx, signupReq)
	require.NoError(t, err)

	// Get initial profile (should be empty/default)
	profileBefore, err := queries.GetUserProfileByUserID(ctx, signupRes.ID)
	if err != nil {
		t.Logf("No profile exists yet (expected for new user): %v", err)
		// Profile may not exist yet, that's okay
	}

	// Generate 60 transactions to trigger batch update (batch size = 50)
	// First 50 will have avg ~500, next 10 will have avg ~1500
	var csvBuilder strings.Builder
	csvBuilder.WriteString("amount,mode,created_at\n")

	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// First 50: amounts around 500
	for i := 0; i < 50; i++ {
		amount := 500.0 + float64(i%10)*10.0 // 500-590
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		csvBuilder.WriteString(fmt.Sprintf("%.1f,UPI,%s\n", amount, timestamp.Format(time.RFC3339)))
	}

	// Next 10: amounts around 1500
	for i := 50; i < 60; i++ {
		amount := 1500.0 + float64(i%10)*10.0 // 1500-1590
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		csvBuilder.WriteString(fmt.Sprintf("%.1f,CARD,%s\n", amount, timestamp.Format(time.RFC3339)))
	}

	csvContent := csvBuilder.String()
	reader := strings.NewReader(csvContent)

	// Process bulk transactions
	bulkRes, err := txnService.ProcessBulkTransactions(ctx, signupRes.ID, reader, "batch_test.csv")
	require.NoError(t, err)
	assert.Equal(t, 60, bulkRes.Success, "All 60 transactions should succeed")
	assert.Equal(t, 0, bulkRes.Failed, "No transactions should fail")
	assert.Equal(t, 60, bulkRes.Processed, "All 60 transactions should be processed")

	// Get profile after batch processing
	profileAfter, err := queries.GetUserProfileByUserID(ctx, signupRes.ID)
	require.NoError(t, err, "Profile should exist after processing")

	// Verify profile was updated
	assert.True(t, profileAfter.AverageTransactionAmount.Valid, "Average should be calculated")
	assert.True(t, profileAfter.StdDevTransactionAmount.Valid, "Std Dev should be calculated")

	// Expected average: (50 * ~545 + 10 * ~1545) / 60 ≈ 720
	// Allow some margin for the exact calculation
	avgAmount := profileAfter.AverageTransactionAmount.Float64
	assert.Greater(t, avgAmount, 600.0, "Average should be greater than 600")
	assert.Less(t, avgAmount, 900.0, "Average should be less than 900")

	// Std dev should be > 0 since we have variance
	stdDev := float64(profileAfter.StdDevTransactionAmount.Int32)
	assert.Greater(t, stdDev, 0.0, "Standard deviation should be greater than 0")
	assert.Greater(t, stdDev, 100.0, "Standard deviation should reflect the variance in amounts")

	// Verify total transactions count
	assert.Equal(t, int32(60), profileAfter.TotalTransactions, "Should have 60 total transactions")

	// Log the profile for debugging
	t.Logf("Profile Before: AvgAmount=%v, StdDev=%v, TotalTxns=%d",
		profileBefore.AverageTransactionAmount,
		profileBefore.StdDevTransactionAmount,
		profileBefore.TotalTransactions)

	t.Logf("Profile After: AvgAmount=%.2f, StdDev=%.2f, TotalTxns=%d, AllowedTxns=%d",
		profileAfter.AverageTransactionAmount.Float64,
		float64(profileAfter.StdDevTransactionAmount.Int32),
		profileAfter.TotalTransactions,
		profileAfter.AllowedTransactions)

	// Additional verification: Check that batch update happened mid-processing
	// by verifying profile reflects all 60 transactions, not just first 50
	// If batch update didn't work, profile would only reflect first 50 (avg ~545)
	// With batch update, it should reflect all 60 (avg ~720)

	// The fact that average > 600 proves batch update occurred,
	// because first 50 transactions alone would give avg ~545
}
