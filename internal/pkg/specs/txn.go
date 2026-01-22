package specs

import "github.com/cheemx5395/fraud-detection-lite/internal/repository"

type CreateTransactionRequest struct {
	Amount int    `json:"amount"`
	Mode   string `json:"mode"`
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
	TransactionID    int32                          `json:"txn_id"`
	Decision         repository.TransactionDecision `json:"decision"`
	RiskScore        int32                          `json:"risk_score"`
	TriggeredFactors []string                       `json:"triggered_factors"`
}
