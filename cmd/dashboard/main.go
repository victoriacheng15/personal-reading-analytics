package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"

	schema "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal"
	dashboard "github.com/victoriacheng15/personal-reading-analytics-dashboard/cmd/internal/dashboard"
)

const (
	dashboardTitle = "📚 Personal Reading Analytics"
)

// loadLatestMetrics reads the most recent metrics JSON file from metrics/ folder
func loadLatestMetrics() (schema.Metrics, error) {
	entries, err := os.ReadDir("metrics")
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to read metrics directory: %w", err)
	}

	if len(entries) == 0 {
		return schema.Metrics{}, fmt.Errorf("no metrics files found in metrics/ folder")
	}

	// Find the latest metrics file (they are named YYYY-MM-DD.json)
	var latestFile string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() > latestFile {
			latestFile = entry.Name()
		}
	}

	if latestFile == "" {
		return schema.Metrics{}, fmt.Errorf("no valid metrics files found")
	}

	log.Printf("Loading metrics from: metrics/%s\n", latestFile)

	// Read and parse the JSON file
	data, err := os.ReadFile(fmt.Sprintf("metrics/%s", latestFile))
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to read metrics file: %w", err)
	}

	var metrics schema.Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return schema.Metrics{}, fmt.Errorf("unable to parse metrics JSON: %w", err)
	}

	return metrics, nil
}

// generateHTMLDashboard creates and saves the HTML dashboard file
func generateHTMLDashboard(metrics schema.Metrics) error {
	// Sort sources by count
	var sources []schema.SourceInfo
	for name, count := range metrics.BySource {
		readStatus := metrics.BySourceReadStatus[name]
		read := readStatus[0]
		unread := readStatus[1]
		readPct := 0.0
		if count > 0 {
			readPct = (float64(read) / float64(count)) * 100
		}

		authorCount := 0
		if name == "Substack" {
			authorCount = metrics.BySourceReadStatus["substack_author_count"][0]
		}

		sources = append(sources, schema.SourceInfo{
			Name:        name,
			Count:       count,
			Read:        read,
			Unread:      unread,
			ReadPct:     readPct,
			AuthorCount: authorCount,
		})
	}

	// Sort by count descending
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Count > sources[j].Count
	})

	// Build month info
	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}
	var months []schema.MonthInfo
	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		if total, exists := metrics.ByMonthOnly[monthStr]; exists && total > 0 {
			months = append(months, schema.MonthInfo{
				Name:    monthNames[month],
				Month:   monthStr,
				Total:   total,
				Sources: metrics.ByMonthAndSource[monthStr],
			})
		}
	}

	// Build year info
	var years []schema.YearInfo
	for year, count := range metrics.ByYear {
		years = append(years, schema.YearInfo{Year: year, Count: count})
	}
	sort.Slice(years, func(i, j int) bool {
		return years[i].Year > years[j].Year
	})

	// Load HTML template from file
	templateContent, err := dashboard.LoadTemplateContent()
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Parse and execute template
	tmpl := template.New("dashboard")
	tmpl.Funcs(template.FuncMap{
		"divideFloat": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
	})

	tmpl, err = tmpl.Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Create site directory
	os.MkdirAll("site", 0755)

	// Create output file
	file, err := os.Create("site/index.html")
	if err != nil {
		return fmt.Errorf("failed to create site/index.html: %w", err)
	}
	defer file.Close()

	// Prepare chart data using dashboard helpers
	yearChartData := dashboard.PrepareYearChartData(years)
	monthChartData := dashboard.PrepareMonthChartData(months, sources)

	// Execute template
	data := map[string]interface{}{
		"DashboardTitle":      dashboardTitle,
		"TotalArticles":       metrics.TotalArticles,
		"ReadCount":           metrics.ReadCount,
		"UnreadCount":         metrics.UnreadCount,
		"ReadRate":            metrics.ReadRate,
		"AvgArticlesPerMonth": metrics.AvgArticlesPerMonth,
		"LastUpdated":         metrics.LastUpdated,
		"Sources":             sources,
		"Months":              months,
		"Years":               years,
		"YearChartLabels":     template.JS(yearChartData.LabelsJSON),
		"YearChartData":       template.JS(yearChartData.DataJSON),
		"MonthChartLabels":    template.JS(monthChartData.LabelsJSON),
		"MonthChartDatasets":  template.JS(monthChartData.DatasetsJSON),
		"MonthTotalData":      template.JS(monthChartData.TotalDataJSON),
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Println("✅ HTML dashboard generated at site/index.html")
	return nil
}

func main() {
	// Load latest metrics from metrics/ folder
	metrics, err := loadLatestMetrics()
	if err != nil {
		log.Fatalf("Failed to load metrics: %v", err)
	}

	// Generate HTML dashboard
	if err := generateHTMLDashboard(metrics); err != nil {
		log.Fatalf("failed to generate dashboard: %v", err)
	}

	log.Println("✅ Successfully generated dashboard from metrics")
}
