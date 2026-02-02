package helpers

import (
	"slices"
	"time"

	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/constants"
	"github.com/cheemx5395/fraud-detection-lite/internal/pkg/specs"
	"github.com/cheemx5395/fraud-detection-lite/internal/repository"
)

// CalculateProfileConfidence calculates the user's profile confidence score
// Formula: (allowed_transactions / total_transactions) * 100
func CalculateProfileConfidence(profile *repository.UserProfileBehavior) float64 {
	if profile.AllowedTransactions <= 0 {
		return 0.0
	}

	confidence := (float64(profile.AllowedTransactions) / 50.0) * 100.0
	return min(confidence, 100.0)
}

// CalculateAmountDeviationRisk calculates risk based on how
// much the transaction amount deviates from user's average
// spending patterns
func CalculateAmountDeviationRisk(transactionAmount int32, profile *repository.UserProfileBehavior) float64 {
	if !profile.AverageTransactionAmount.Valid || profile.TotalTransactions < constants.MinTransactionsForProfiling {
		if !profile.MaxTransactionAmountSeen.Valid {
			return 30.0
		}

		// comparing with maximum transaction seen till now
		if transactionAmount > int32(profile.MaxTransactionAmountSeen.Float64) {
			ratio := float64(transactionAmount) / float64(profile.MaxTransactionAmountSeen.Float64)

			risk := 20.0 + (ratio-1.0)*30.0
			return min(risk, 100.0)
		}
		return 10.0
	}

	avgAmount := float64(profile.AverageTransactionAmount.Float64)
	txAmount := float64(transactionAmount)

	// Calculate deviation ratio
	ratio := txAmount / avgAmount

	var risk float64
	if ratio <= 1.0 {
		risk = 0.0
	} else if ratio <= constants.AmountDeviationModerate {
		risk = 5.0 + ((ratio-1.0)/(constants.AmountDeviationModerate-1.0))*20.0
	} else if ratio <= constants.AmountDeviationHigh {
		risk = 25.0 + ((ratio-constants.AmountDeviationModerate)/(constants.AmountDeviationHigh-constants.AmountDeviationModerate))*45.0
	} else {
		risk = 70.0 + min((ratio-constants.AmountDeviationHigh)*10.0, 30.0)
	}
	return min(risk, 100.0)
}

// CalculateFrequencySpikeRisk calculates risk based on transaction frequency
// compared to user's normal pattern
func CalculateFrequencySpikeRisk(
	profile *repository.UserProfileBehavior,
	recentTransactionCount int,
) float64 {

	windowHours := constants.FrequencyWindowHours.Hours()

	if !profile.AverageNumberOfTransactionsPerDay.Valid ||
		profile.TotalTransactions < constants.MinTransactionsForProfiling {

		// Allow ~3 tx per hour for new users
		maxAllowed := int(3 * windowHours)

		if recentTransactionCount > maxAllowed {
			risk := 40.0 + float64(recentTransactionCount-maxAllowed)*10.0
			return min(risk, 100.0)
		}
		return 10.0
	}

	avgPerDay := float64(profile.AverageNumberOfTransactionsPerDay.Int32)

	expectedInWindow := avgPerDay * (windowHours / 24.0)
	if expectedInWindow < 0.5 {
		expectedInWindow = 0.5
	}

	actual := float64(recentTransactionCount)
	ratio := actual / expectedInWindow

	var risk float64
	switch {
	case ratio <= 1.5:
		risk = 0.0
	case ratio <= 3.0:
		risk = 15.0 + ((ratio-1.5)/1.5)*35.0
	default:
		risk = 50.0 + min((ratio-3.0)*15.0, 50.0)
	}

	return min(risk, 100.0)
}

// CalculateModeDeviationRisk calculates risk when user uses a payment mode
// that's not in their registered modes
func CalculateModeDeviationRisk(transactionMode repository.Mode, profile *repository.UserProfileBehavior) float64 {
	// check if mode is registered
	if slices.Contains(profile.RegisteredPaymentModes, transactionMode) {
		return 0.0
	}

	// New Mode detected
	// Risk is lower to users with high profile confidence
	profileConfidence := CalculateProfileConfidence(profile)

	// Base risk for new mode is 60
	// Reduced by profile confidence: high confidence users get lower penalty
	// Formula: 60 - (profile_confidence * 0.3)
	// 100% confidence: 60 - 30 = 30 risk
	// 50% confidence: 60 - 15 = 45 risk
	// 0% confidence: 60 - 0 = 60 risk
	baseRisk := 60.0
	reduction := (profileConfidence / 100.0) * 30.0
	risk := baseRisk - reduction

	// risk for new mode will be always [20.0, 60.0]
	return max(risk, 20.0)
}

// CalculateTimeAnomalyRisk calculates risk based on transaction time
// compared to user's usual trasaction hours
func CalculateTimeAnomalyRisk(transactionTime time.Time, profile *repository.UserProfileBehavior) float64 {
	currentHour := transactionTime.Hour()

	// No pattern established - use general heuristics
	if !profile.UsualTransactionStartHour.Valid || !profile.UsualTransactionEndHour.Valid {
		// Late night/early morning (12 AM - 5 AM) is riskier
		if currentHour >= 0 && currentHour < 5 {
			return 35.0
		}
		// Very early morning (5 AM - 7 AM)
		if currentHour >= 5 && currentHour < 7 {
			return 20.0
		}
		return 5.0
	}

	// Extract hours from timestamps
	startHour := profile.UsualTransactionStartHour.Time.Hour()
	endHour := profile.UsualTransactionEndHour.Time.Hour()

	// check if within usual hours
	isWithinUsualHours := false
	if startHour <= endHour {
		isWithinUsualHours = currentHour >= startHour && currentHour <= endHour
	} else {
		isWithinUsualHours = currentHour >= startHour || currentHour <= endHour
	}

	if isWithinUsualHours {
		return 0.0
	}

	var hoursOutside int
	if startHour <= endHour {
		if currentHour < startHour {
			hoursOutside = startHour - currentHour
		} else {
			hoursOutside = currentHour - endHour
		}
	} else {
		if currentHour > endHour && currentHour < startHour {
			distFromEnd := currentHour - endHour
			distFromStart := startHour - currentHour
			hoursOutside = min(distFromEnd, distFromStart)
		}
	}

	risk := float64(hoursOutside) * 10.0

	if currentHour >= 0 && currentHour < 4 {
		risk += 15.0
	}

	return min(risk, 100.0)
}

// CalculateAggregateRiskScore combines all facor scores into final risk score
// using weighted sum
func CalculateAggregateRiskScore(
	amountRisk float64,
	frequencyRisk float64,
	modeRisk float64,
	timeRisk float64,
) float64 {
	aggregateRisk := (amountRisk * constants.WeightAmountDeviation) + (frequencyRisk * constants.WeightFrequencySpike) + (modeRisk * constants.WeightModeDeviation) + (timeRisk * constants.WeightTimeAnomaly)

	return min(aggregateRisk, 100.0)
}

// DampenRiskWithProfileConfidence reduces risk score based on
// users' trustworthiness
// High confidence users get more benefits of doubt
// Formula: dampened_risk = raw_risk * (1 - (profile_confidence / 200))
func DampenRiskWithProfileConfidence(rawRiskScore float64, profileConfidence float64) float64 {
	dampeningFactor := 1.0 - (profileConfidence / 200.0)

	// cap trust benefit
	if dampeningFactor < 0.5 {
		dampeningFactor = 0.5
	}

	dampenedRisk := rawRiskScore * dampeningFactor

	// dynamic floor: 10% of raw risk
	minFloor := rawRiskScore * 0.1
	if dampenedRisk < minFloor {
		return minFloor
	}

	return dampenedRisk
}

// DetermineTriggeredFactors identifies which factors exceeded their thresholds
func DetermineTriggeredFactors(
	amountRisk float64,
	frequencyRisk float64,
	modeRisk float64,
	timeRisk float64,
) []string {
	triggered := []string{}

	if amountRisk > constants.ThresholdAmountDeviation {
		triggered = append(triggered, constants.TriggerFactorsAMOUNTDEVIATION)
	}
	if frequencyRisk > constants.ThresholdFrequencySpike {
		triggered = append(triggered, constants.TriggerFactorsFREQUENCYSPIKE)
	}
	if modeRisk > constants.ThresholdModeDeviation {
		triggered = append(triggered, constants.TriggerFactorsNEWMODE)
	}
	if timeRisk > constants.ThresholdTimeAnomaly {
		triggered = append(triggered, constants.TriggerFactorsTIMEANOMALY)
	}
	return triggered
}

// DetermineTransactionDecision decides the action based on final risk score
func DetermineTransactionDecision(finalRiskScore float64, profile *repository.UserProfileBehavior) repository.TransactionDecision {
	if profile.TotalTransactions < constants.MinTransactionsForProfiling {
		if finalRiskScore < 60.0 {
			return repository.TransactionDecisionALLOW
		} else if finalRiskScore < 75.0 {
			return repository.TransactionDecisionFLAG
		}
		return repository.TransactionDecisionMFAREQUIRED
	}

	if finalRiskScore < constants.RiskThresholdAllow {
		return repository.TransactionDecisionALLOW
	} else if finalRiskScore < constants.RiskThresholdFlag {
		return repository.TransactionDecisionFLAG
	} else if finalRiskScore < constants.RiskThresholdMFA {
		return repository.TransactionDecisionMFAREQUIRED
	}
	return repository.TransactionDecisionBLOCK
}

func AnalyzeBulkTransactions(
	req *specs.CreateBulkTransactionRequest,
	profile *repository.UserProfileBehavior,
	recentTransactionCount int,
) specs.FraudAnalysisResult {
	amountRisk := CalculateAmountDeviationRisk(int32(req.Amount), profile)
	frequencyRisk := CalculateFrequencySpikeRisk(profile, recentTransactionCount)
	modeRisk := CalculateModeDeviationRisk(repository.Mode(req.Mode), profile)
	timeRisk := CalculateTimeAnomalyRisk(req.CreatedAt, profile)

	rawRiskScore := CalculateAggregateRiskScore(amountRisk, frequencyRisk, modeRisk, timeRisk)

	profileConfidence := CalculateProfileConfidence(profile)

	finalRiskScore := DampenRiskWithProfileConfidence(rawRiskScore, profileConfidence)

	triggeredFactors := DetermineTriggeredFactors(amountRisk, frequencyRisk, modeRisk, timeRisk)

	decision := DetermineTransactionDecision(finalRiskScore, profile)

	return specs.FraudAnalysisResult{
		Message:           "analysis result",
		Decision:          decision,
		FinalRiskScore:    int32(finalRiskScore),
		RawRiskScore:      rawRiskScore,
		ProfileConfidence: profileConfidence,
		TriggeredFactors:  triggeredFactors,
		AmountRisk:        amountRisk,
		FrequencyRisk:     frequencyRisk,
		ModeRisk:          modeRisk,
		TimeRisk:          timeRisk,
	}

}

// AnalyzeTransaction performs complete fraud analysis and returns specs.FraudAnalysisResult
func AnalyzeTransaction(
	req *specs.CreateTransactionRequest,
	profile *repository.UserProfileBehavior,
	recentTransactionCount int,
	transactionTime time.Time,
) specs.FraudAnalysisResult {
	amountRisk := CalculateAmountDeviationRisk(int32(req.Amount), profile)
	frequencyRisk := CalculateFrequencySpikeRisk(profile, recentTransactionCount)
	modeRisk := CalculateModeDeviationRisk(repository.Mode(req.Mode), profile)
	timeRisk := CalculateTimeAnomalyRisk(transactionTime, profile)

	rawRiskScore := CalculateAggregateRiskScore(amountRisk, frequencyRisk, modeRisk, timeRisk)

	profileConfidence := CalculateProfileConfidence(profile)

	finalRiskScore := DampenRiskWithProfileConfidence(rawRiskScore, profileConfidence)

	triggeredFactors := DetermineTriggeredFactors(amountRisk, frequencyRisk, modeRisk, timeRisk)

	decision := DetermineTransactionDecision(finalRiskScore, profile)

	return specs.FraudAnalysisResult{
		Message:           "analysis result",
		Decision:          decision,
		FinalRiskScore:    int32(finalRiskScore),
		RawRiskScore:      rawRiskScore,
		ProfileConfidence: profileConfidence,
		TriggeredFactors:  triggeredFactors,
		AmountRisk:        amountRisk,
		FrequencyRisk:     frequencyRisk,
		ModeRisk:          modeRisk,
		TimeRisk:          timeRisk,
	}
}
