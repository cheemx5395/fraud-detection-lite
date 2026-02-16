package specs

import (
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/errors"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
)

type CreateTransactionRequest struct {
	Amount float64 `json:"amount"`
	Mode   string  `json:"mode"`
}

func (r CreateTransactionRequest) Validate() error {
	switch {
	case r.Amount == 0.0 && r.Mode == "":
		return errors.ErrInvalidBody
	case r.Amount == 0.0:
		return errors.ErrMissingAmountInRequest
	case r.Mode == "":
		return errors.ErrMissingModeInRequest
	}

	if r.Amount < 0 || r.Amount > 1e16 {
		return errors.ErrAmountOutOfRange
	}

	switch repository.Mode(r.Mode) {
	case repository.ModeUPI, repository.ModeCARD, repository.ModeNETBANKING:
		return nil
	default:
		return errors.ErrInvalidPaymentMode
	}
}

type CreateBulkTransactionRequest struct {
	Amount    float64   `json:"amount"`
	Mode      string    `json:"mode"`
	CreatedAt time.Time `json:"created_at"`
}

type FraudAnalysisResult struct {
	Message           string                         `json:"message"`
	Decision          repository.TransactionDecision `json:"decision"`
	FinalRiskScore    int32                          `json:"final_risk_score"`
	RawRiskScore      float64                        `json:"raw_risk_score"`
	ProfileConfidence float64                        `json:"profile_confidence"`
	TriggeredFactors  []string                       `json:"triggered_factors"`
	AmountRisk        float64                        `json:"amount_risk"`
	FrequencyRisk     float64                        `json:"frequency_risk"`
	ModeRisk          float64                        `json:"mode_risk"`
	TimeRisk          float64                        `json:"time_risk"`
}

type CreateTransactionResponse struct {
	TransactionID    int32                          `json:"id"`
	Decision         repository.TransactionDecision `json:"decision"`
	RiskScore        int32                          `json:"risk_score"`
	TriggeredFactors []string                       `json:"triggered_factors"`
	CreatedAt        time.Time                      `json:"created_at"`
}

type BulkProcessResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Processed int    `json:"processed"`
	Success   int    `json:"success"`
	Failed    int    `json:"failed"`
}
