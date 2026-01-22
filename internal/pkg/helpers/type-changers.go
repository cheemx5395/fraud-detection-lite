package helpers

import "github.com/cheemx5395/fraud-detection-lite/internal/repository"

func GetStringSliceFromModeSlice(modes []repository.Mode) []string {
	stringModes := []string{}
	for _, mode := range modes {
		stringModes = append(stringModes, string(mode))
	}
	return stringModes
}

func GetModeSliceFromStringSlice(stringModes []string) []repository.Mode {
	modes := []repository.Mode{}
	for _, strMode := range stringModes {
		modes = append(modes, repository.Mode(strMode))
	}
	return modes
}

func GetStringSliceFromTriggerFactorsSlice(factors []repository.TriggerFactors) []string {
	stringFactors := []string{}
	for _, factor := range factors {
		stringFactors = append(stringFactors, string(factor))
	}
	return stringFactors
}
