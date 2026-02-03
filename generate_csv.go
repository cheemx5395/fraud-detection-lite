package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var (
	modes = []string{"UPI", "CARD", "NETBANKING"}
)

func main() {
	rand.Seed(time.Now().UnixNano())

	file, err := os.Create("stress_transactions_1000.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	writer.Write([]string{"amount", "mode", "created_at"})

	startTime := time.Now().AddDate(0, -4, 0) // 4 months ago
	baseAmount := 500.0

	lastTxnTime := startTime

	for i := 0; i < 1000; i++ {
		p := rand.Float64()

		var (
			amount  float64
			mode    string
			txnTime time.Time
		)

		switch {
		// ---------- Amount deviation ----------
		case p < 0.15:
			amount = baseAmount * (2.5 + rand.Float64()*2.5) // 2.5x–5x
			mode = randomMode()
			txnTime = randomDaytime(lastTxnTime)

		// ---------- Frequency spike ----------
		case p < 0.25:
			amount = baseAmount * (0.8 + rand.Float64()*0.4)
			mode = randomMode()
			txnTime = lastTxnTime.Add(time.Duration(rand.Intn(3)+1) * time.Minute)

		// ---------- New payment mode ----------
		case p < 0.30:
			amount = baseAmount * (0.9 + rand.Float64()*0.3)
			mode = "CRYPTO" // intentionally unknown / new
			txnTime = randomDaytime(lastTxnTime)

		// ---------- Time anomaly ----------
		case p < 0.35:
			amount = baseAmount * (0.9 + rand.Float64()*0.3)
			mode = randomMode()
			txnTime = randomLateNight(lastTxnTime)

		// ---------- Normal transaction ----------
		default:
			amount = baseAmount * (0.8 + rand.Float64()*0.4)
			mode = randomMode()
			txnTime = randomDaytime(lastTxnTime)
		}

		lastTxnTime = txnTime

		writer.Write([]string{
			fmt.Sprintf("%.2f", amount),
			mode,
			txnTime.Format(time.RFC3339),
		})
	}

	fmt.Println("stress_transactions_1000.csv generated")
}

func randomMode() string {
	return modes[rand.Intn(len(modes))]
}

func randomDaytime(prev time.Time) time.Time {
	// spread across days
	deltaDays := rand.Intn(2)
	hour := rand.Intn(10) + 9 // 9 AM – 7 PM
	min := rand.Intn(60)

	return time.Date(
		prev.Year(),
		prev.Month(),
		prev.Day()+deltaDays,
		hour,
		min,
		0,
		0,
		prev.Location(),
	)
}

func randomLateNight(prev time.Time) time.Time {
	deltaDays := rand.Intn(2)
	hour := rand.Intn(4) // 12–4 AM
	min := rand.Intn(60)

	return time.Date(
		prev.Year(),
		prev.Month(),
		prev.Day()+deltaDays,
		hour,
		min,
		0,
		0,
		prev.Location(),
	)
}
