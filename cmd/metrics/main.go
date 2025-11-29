package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	metrics "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal/metrics"
)

func main() {
	ctx := context.Background()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, will use environment variables")
	}

	sheetID := os.Getenv("SHEET_ID")
	credentialsPath := os.Getenv("CREDENTIALS_PATH")

	if sheetID == "" {
		log.Fatal("SHEET_ID environment variable is required")
	}
	if credentialsPath == "" {
		credentialsPath = "./credentials.json"
	}

	// Fetch metrics from Google Sheets
	metricsData, err := metrics.FetchMetricsFromSheets(ctx, sheetID, credentialsPath)
	if err != nil {
		log.Fatalf("Failed to fetch metrics: %v", err)
	}

	// Save metrics as JSON with timestamp
	os.MkdirAll("metrics", 0755)

	metricsJSON, err := json.MarshalIndent(metricsData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metrics: %v", err)
	}

	// Save to metrics folder with date filename (YYYY-MM-DD.json)
	dateFilename := metricsData.LastUpdated.Format("2006-01-02") + ".json"
	metricsFilePath := fmt.Sprintf("metrics/%s", dateFilename)
	err = os.WriteFile(metricsFilePath, metricsJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write metrics file: %v", err)
	}

	log.Printf("✅ Metrics saved to metrics/%s\n", dateFilename)
	log.Println("✅ Successfully generated metrics from Google Sheets")
}
