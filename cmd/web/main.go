package main

import (
	"log"
	"path/filepath"

	web "github.com/victoriacheng15/personal-reading-analytics/internal/web"
)

func main() {
	// 1. Get all available metrics dates
	dates, err := web.GetMetricsDates()
	if err != nil {
		log.Fatalf("Failed to discover metrics: %v", err)
	}

	// 2. Initialize Analytics Service
	service := web.NewAnalyticsService("dist")

	log.Printf("Generating reports for %d dates...\n", len(dates))

	// 3. Multi-pass generation
	for i, date := range dates {
		metrics, err := web.LoadMetricsByDate(date)
		if err != nil {
			log.Printf("⚠️ Warning: Skipping %s: %v\n", date, err)
			continue
		}

		// Historical: ONLY analytics.html in dist/history/YYYY-MM-DD
		err = service.GenerateAnalyticsOnly(metrics, web.GenConfig{
			OutputDir:    filepath.Join("dist", "history", date),
			BaseURL:      "../../",
			IsHistorical: true,
			HistoryDates: dates,
			ReportDate:   date,
		})
		if err != nil {
			log.Printf("⚠️ Warning: Failed historical generation for %s: %v\n", date, err)
		}

		// Latest (root): ALL pages in dist/
		if i == 0 {
			err = service.GenerateFullSite(metrics, web.GenConfig{
				OutputDir:    "dist",
				BaseURL:      "./",
				IsHistorical: false,
				HistoryDates: dates,
				ReportDate:   date,
			})
			if err != nil {
				log.Fatalf("Failed to generate latest site: %v", err)
			}
		}
	}

	log.Println("✅ Successfully generated all historical and latest analytics")
}
