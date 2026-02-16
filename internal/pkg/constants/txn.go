package constants

import "time"

const (
	// Factor weights (must sum to 1.0 for proper risk calculation)
	WeightAmountDeviation = 0.40 // 40%
	WeightFrequencySpike  = 0.30 // 30%
	WeightModeDeviation   = 0.20 // 20%
	WeightTimeAnomaly     = 0.10 // 10%

	// Risk thresholds for each factor (0-100 scale)
	ThresholdAmountDeviation = 30.0 // Trigger if score > 30
	ThresholdFrequencySpike  = 40.0 // Trigger if score > 40
	ThresholdModeDeviation   = 50.0 // Trigger if score > 50
	ThresholdTimeAnomaly     = 35.0 // Trigger if score > 35

	// Decision thresholds (after dampening with profile confidence)
	RiskThresholdAllow = 30.0 // < 30: Allow
	RiskThresholdFlag  = 60.0 // 30-60: Flag
	RiskThresholdMFA   = 80.0 // 60-80: MFA Required
	// > 80: Block

	// Time window for frequency calculation (in hours)
	FrequencyWindowHours = 1.0 * time.Hour

	// Frequency constants
	ThresholdFrequency       = 3
	RiskPerTxnAfterThreshold = 20.0

	// Amount deviation multipliers
	AmountDeviationModerate = 1.5 // 1.5x average is moderate risk
	AmountDeviationHigh     = 3.0 // 3x average is high risk

	// Minimum transactions needed for reliable profiling
	MinTransactionsForProfiling = 5

	TriggerFactorsAMOUNTDEVIATION = "AMOUNT_DEVIATION"
	TriggerFactorsFREQUENCYSPIKE  = "FREQUENCY_SPIKE"
	TriggerFactorsNEWMODE         = "NEW_MODE"
	TriggerFactorsTIMEANOMALY     = "TIME_ANOMALY"
)
