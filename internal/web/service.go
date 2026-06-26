package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	texttmpl "text/template"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics/internal"
	"github.com/victoriacheng15/personal-reading-analytics/internal/metrics"
	"gopkg.in/yaml.v3"
)

const (
	AnalyticsTitle = "📚 Personal Reading Analytics"
)

// AnalyticsService handles the generation of the HTML analytics
type AnalyticsService struct {
	outputDir string
}

// NewAnalyticsService creates a new AnalyticsService
func NewAnalyticsService(outputDir string) *AnalyticsService {
	return &AnalyticsService{outputDir: outputDir}
}

// GenerateFullSite generates all pages (index, analytics, evolution)
func (s *AnalyticsService) GenerateFullSite(m schema.Metrics, config GenConfig) error {
	vm, err := s.prepareViewModel(m, config)
	if err != nil {
		return fmt.Errorf("failed to prepare view model: %w", err)
	}

	pages := []struct {
		Filename string
		Title    string
	}{
		{"index.html", AnalyticsTitle},
		{"analytics.html", "📊 Analytics"},
		{"evolution.html", "⏳ Evolution"},
	}

	// Generate machine-readable registry
	if err := s.generateRegistry(vm, config.OutputDir); err != nil {
		log.Printf("⚠️ Warning: Failed to generate evolution registry: %v", err)
	}

	return s.render(vm, config.OutputDir, pages, true)
}

// GenerateAnalyticsOnly generates only the analytics.html page
func (s *AnalyticsService) GenerateAnalyticsOnly(m schema.Metrics, config GenConfig) error {
	vm, err := s.prepareViewModel(m, config)
	if err != nil {
		return fmt.Errorf("failed to prepare view model: %w", err)
	}

	pages := []struct {
		Filename string
		Title    string
	}{
		{"analytics.html", "📊 Analytics (Archived)"},
	}

	return s.render(vm, config.OutputDir, pages, false)
}

func (s *AnalyticsService) prepareViewModel(m schema.Metrics, config GenConfig) (ViewModel, error) {
	// Sort sources by count
	var sources []schema.SourceInfo
	for name, count := range m.BySource {
		readStatus := m.BySourceReadStatus[name]
		read := readStatus[0]
		unread := readStatus[1]
		readPct := 0.0
		if count > 0 {
			readPct = (float64(read) / float64(count)) * 100
		}

		authorCount := 0
		if name == "Substack" {
			authorCount = m.BySourceReadStatus["substack_author_count"][0]
		}

		color := ""
		if meta, exists := m.SourceMetadata[name]; exists {
			color = meta.Color
		}

		sources = append(sources, schema.SourceInfo{
			Name:        name,
			Count:       count,
			Read:        read,
			Unread:      unread,
			ReadPct:     readPct,
			AuthorCount: authorCount,
			Color:       color,
		})
	}

	// Sort by count descending
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Count > sources[j].Count
	})

	// Build year info
	var years []schema.YearInfo
	for year, count := range m.ByYear {
		years = append(years, schema.YearInfo{Year: year, Count: count})
	}
	sort.Slice(years, func(i, j int) bool {
		return years[i].Year > years[j].Year
	})

	// Create aggregated monthly data (Jan-Dec, all years combined)
	var monthlyAggregated []schema.MonthInfo
	// shortMonthNames is defined in preparation.go (same package)

	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		monthShort := shortMonthNames[month-1]

		// Get source data for this month from ByMonthAndSource (aggregated across all years)
		if monthSourceData, exists := m.ByMonthAndSource[monthStr]; exists {
			total := 0
			monthSources := make(map[string]int)

			for source, counts := range monthSourceData {
				articleCount := counts[0] + counts[1] // read + unread
				monthSources[source] = articleCount
				total += articleCount
			}

			if total > 0 {
				monthlyAggregated = append(monthlyAggregated, schema.MonthInfo{
					Name:    monthShort,
					Month:   monthStr,
					Year:    "", // No year for aggregated monthly view
					Total:   total,
					Sources: monthSources,
				})
			}
		}
	}

	// Extract all unique years for filtering
	var allYears []string
	for _, year := range years {
		allYears = append(allYears, year.Year)
	}

	// Extract all unique sources for filtering
	var allSources []string
	for _, source := range sources {
		allSources = append(allSources, source.Name)
	}

	// Determine current month (MM format) for badge calculation
	now := time.Now()
	currentMonth := now.Format("01")

	// If the current month (from system time) has no data,
	// fall back to the latest month available in the metrics to provide
	// a better "latest snapshot" view.
	if _, exists := m.ByMonth[currentMonth]; !exists {
		for month := 12; month >= 1; month-- {
			monthStr := fmt.Sprintf("%02d", month)
			if _, exists := m.ByMonth[monthStr]; exists {
				currentMonth = monthStr
				break
			}
		}
	}

	// Calculate badges using metrics package helpers
	topReadRateSource := metrics.CalculateTopReadRateSource(m)
	mostUnreadSource := metrics.CalculateMostUnreadSource(m)
	thisMonthArticles := metrics.CalculateThisMonthArticles(m, currentMonth)

	// Prepare chart data using analytics helpers
	yearChartData := PrepareYearChartData(years)
	monthChartData := PrepareMonthChartData(monthlyAggregated, sources)

	// Prepare read/unread data for both month and source views
	readUnreadByMonthJSON := PrepareReadUnreadByMonth(m)
	readUnreadBySourceJSON := PrepareReadUnreadBySource(sources)
	readUnreadByYearJSON := PrepareReadUnreadByYear(m)
	unreadArticleAgeDistributionJSON := PrepareUnreadArticleAgeDistribution(m)
	unreadByYearJSON := PrepareUnreadByYear(m)

	// Marshal AllYears and AllSources to JSON for JavaScript
	allYearsJSON, _ := json.Marshal(allYears)
	allSourcesJSON, _ := json.Marshal(allSources)

	// Prepare key metrics
	keyMetrics := []schema.KeyMetric{
		{Title: "Total Articles", Value: fmt.Sprintf("%d", m.TotalArticles)},
		{Title: "Read Rate", Value: fmt.Sprintf("%.1f%%", m.ReadRate)},
		{Title: "Read", Value: fmt.Sprintf("%d", m.ReadCount)},
		{Title: "Unread", Value: fmt.Sprintf("%d", m.UnreadCount)},
		{Title: "Avg/Month", Value: fmt.Sprintf("%.0f", m.AvgArticlesPerMonth)},
	}

	highlightMetrics := []schema.HightlightMetric{
		{Title: "🎯 Top Read Rate Source", Value: topReadRateSource},
		{Title: "📚 Most Unread Source", Value: mostUnreadSource},
		{Title: "✅ This Month's Articles", Value: fmt.Sprintf("%d", thisMonthArticles)},
	}

	// Load evolution data
	evolutionData, err := LoadEvolutionData()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to load evolution data: %v", err)
	} else {
		// Sort chapters by period descending (assuming order in YAML is chronological, we reverse it)
		// Or strictly, we just iterate backwards in the template.
		// But let's reverse the slice here for easier template logic.
		// Actually, let's reverse it here so index 0 is always the "latest".
		for i, j := 0, len(evolutionData.Chapters)-1; i < j; i, j = i+1, j-1 {
			evolutionData.Chapters[i], evolutionData.Chapters[j] = evolutionData.Chapters[j], evolutionData.Chapters[i]
		}

		// Also reverse the timeline events within each chapter to show newest first
		for c := range evolutionData.Chapters {
			for i, j := 0, len(evolutionData.Chapters[c].Timeline)-1; i < j; i, j = i+1, j-1 {
				evolutionData.Chapters[c].Timeline[i], evolutionData.Chapters[c].Timeline[j] = evolutionData.Chapters[c].Timeline[j], evolutionData.Chapters[c].Timeline[i]
			}
		}
	}

	// Load landing content
	landing, err := LoadLanding()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to load landing content: %v", err)
	}

	return ViewModel{
		AnalyticsTitle:                   AnalyticsTitle,
		KeyMetrics:                       keyMetrics,
		HighlightMetrics:                 highlightMetrics,
		TotalArticles:                    m.TotalArticles,
		ReadCount:                        m.ReadCount,
		UnreadCount:                      m.UnreadCount,
		ReadRate:                         m.ReadRate,
		AvgArticlesPerMonth:              m.AvgArticlesPerMonth,
		LastUpdated:                      m.LastUpdated,
		AIDeltaAnalysis:                  m.AIDeltaAnalysis,
		Sources:                          sources,
		Months:                           monthlyAggregated,
		Years:                            years,
		AllYears:                         allYears,
		AllSources:                       allSources,
		AllYearsJSON:                     template.JS(allYearsJSON),
		AllSourcesJSON:                   template.JS(allSourcesJSON),
		YearChartLabels:                  template.JS(yearChartData.LabelsJSON),
		YearChartData:                    template.JS(yearChartData.DataJSON),
		MonthChartLabels:                 template.JS(monthChartData.LabelsJSON),
		MonthChartDatasets:               template.JS(monthChartData.DatasetsJSON),
		MonthTotalData:                   template.JS(monthChartData.TotalDataJSON),
		ReadUnreadByMonthJSON:            readUnreadByMonthJSON,
		ReadUnreadBySourceJSON:           readUnreadBySourceJSON,
		ReadUnreadByYearJSON:             readUnreadByYearJSON,
		UnreadArticleAgeDistributionJSON: unreadArticleAgeDistributionJSON,
		UnreadByYearJSON:                 unreadByYearJSON,
		TopOldestUnreadArticles:          m.TopOldestUnreadArticles,
		EvolutionData:                    evolutionData,
		Landing:                          landing,

		// New fields from config
		BaseURL:      config.BaseURL,
		IsHistorical: config.IsHistorical,
		HistoryDates: config.HistoryDates,
		ReportDate:   config.ReportDate,
	}, nil
}

func (s *AnalyticsService) render(vm ViewModel, outputDir string, pages []struct {
	Filename string
	Title    string
}, isRoot bool) error {
	// Get templates directory
	tmplDir, err := GetTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	// Common function map
	funcMap := template.FuncMap{
		"divideFloat": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy static SEO/AI metadata files recursively
	if isRoot {
		staticSrc := filepath.Join(tmplDir, "static")
		if err := s.copyStaticFiles(staticSrc, outputDir, vm); err != nil {
			log.Printf("⚠️ Warning: Failed to process static directory: %v", err)
		}
	}

	// Loop and generate each page
	for _, page := range pages {
		// Create new template instance for this page
		tmpl := template.New("").Funcs(funcMap)

		// Parse shared templates and the specific page template
		files := []string{
			filepath.Join(tmplDir, "base.html"),
			filepath.Join(tmplDir, page.Filename),
		}

		// Parse files
		tmpl, err = tmpl.ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("failed to parse templates for %s: %w", page.Filename, err)
		}

		// Create output file
		outPath := filepath.Join(outputDir, page.Filename)
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", outPath, err)
		}
		defer f.Close()

		// Update PageTitle in ViewModel for this page
		vm.PageTitle = page.Title

		// Execute the template matching the filename
		err = tmpl.ExecuteTemplate(f, page.Filename, vm)
		if err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", page.Filename, err)
		}
	}

	return nil
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			if err = copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

// copyStaticFiles recursively processes the static directory, treating certain files as templates
func (s *AnalyticsService) copyStaticFiles(src, dst string, vm ViewModel) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := s.copyStaticFiles(srcPath, dstPath, vm); err != nil {
				return err
			}
			continue
		}

		// Treat text files as templates to inject config
		if entry.Name() == "llms.txt" || entry.Name() == "robots.txt" {
			t, err := texttmpl.ParseFiles(srcPath)
			if err != nil {
				log.Printf("⚠️ Warning: Failed to parse %s as template: %v", entry.Name(), err)
				continue
			}

			f, err := os.Create(dstPath)
			if err != nil {
				log.Printf("⚠️ Warning: Failed to create %s: %v", dstPath, err)
				continue
			}

			if err := t.Execute(f, vm); err != nil {
				log.Printf("⚠️ Warning: Failed to execute template %s: %v", entry.Name(), err)
			}
			f.Close()
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				log.Printf("⚠️ Warning: Failed to copy static file %s: %v", entry.Name(), err)
			}
		}
	}

	return nil
}

// generateRegistry creates the evolution-registry.json file from the evolution data
func (s *AnalyticsService) generateRegistry(vm ViewModel, outputDir string) error {
	registry := schema.Registry{
		Project:            vm.Landing.Header.ProjectName,
		Version:            "1.0.0",
		LastUpdated:        vm.LastUpdated.Format("2006-01-02"),
		MachineRegistryURL: "/api/evolution-registry.json",
	}

	for _, chapter := range vm.EvolutionData.Chapters {
		for _, milestone := range chapter.Timeline {
			registry.Milestones = append(registry.Milestones, schema.RegistryMilestone{
				Date:        milestone.Date,
				Title:       milestone.Title,
				Category:    chapter.Title,
				Description: strings.TrimSpace(milestone.Description),
			})
		}
	}

	// Create api directory in output
	apiDir := filepath.Join(outputDir, "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return fmt.Errorf("failed to create api directory: %w", err)
	}

	registryJSON, err := json.Marshal(registry)
	if err != nil {
		return fmt.Errorf("failed to marshal registry to JSON: %w", err)
	}

	registryPath := filepath.Join(apiDir, "evolution-registry.json")
	if err := os.WriteFile(registryPath, registryJSON, 0644); err != nil {
		return fmt.Errorf("failed to write evolution-registry.json: %w", err)
	}

	return nil
}

// ==============================================================================
// CHART PREPARATION HELPERS
// ==============================================================================

// PrepareYearChartData prepares year breakdown chart data
func PrepareYearChartData(years []schema.YearInfo) *YearChartData {
	labels := make([]string, 0)
	data := make([]int, 0)

	for _, year := range years {
		labels = append(labels, year.Year)
		data = append(data, year.Count)
	}

	labelsJSON, _ := json.Marshal(labels)
	dataJSON, _ := json.Marshal(data)

	return &YearChartData{
		LabelsJSON: labelsJSON,
		DataJSON:   dataJSON,
	}
}

// PrepareMonthChartData prepares month breakdown chart data with source stacking
func PrepareMonthChartData(months []schema.MonthInfo, sources []schema.SourceInfo) *MonthChartData {
	monthLabels := make([]string, 0)
	for _, month := range months {
		// Just use the month name for aggregated monthly view (no year)
		monthLabels = append(monthLabels, month.Name)
	}
	monthLabelsJSON, _ := json.Marshal(monthLabels)

	datasetsMap := make(map[string][]int)

	// Initialize all sources with data for each month
	for _, source := range sources {
		datasetsMap[source.Name] = make([]int, len(months))
	}

	// Populate data from month.Sources
	for monthIdx, month := range months {
		for sourceName, articleCount := range month.Sources {
			if _, exists := datasetsMap[sourceName]; exists {
				datasetsMap[sourceName][monthIdx] = articleCount
			}
		}
	}

	// Create Chart.js datasets
	var datasets []map[string]interface{}
	for _, source := range sources {
		if data, exists := datasetsMap[source.Name]; exists && len(data) > 0 {
			color := source.Color
			if color == "" {
				color = "#" + colorHash(source.Name)
			}
			dataset := map[string]interface{}{
				"label":           source.Name,
				"data":            data,
				"backgroundColor": color,
				"borderColor":     "#2d3748",
				"borderWidth":     1,
			}
			datasets = append(datasets, dataset)
		}
	}

	datasetsJSON, _ := json.Marshal(datasets)

	// Prepare total data for months (for the line chart view)
	monthTotalData := make([]int, 0)
	for _, month := range months {
		monthTotalData = append(monthTotalData, month.Total)
	}
	monthTotalDataJSON, _ := json.Marshal(monthTotalData)

	return &MonthChartData{
		LabelsJSON:    monthLabelsJSON,
		DatasetsJSON:  datasetsJSON,
		TotalDataJSON: monthTotalDataJSON,
	}
}

// colorHash generates a simple hash for generating colors
func colorHash(s string) string {
	h := uint32(5381)
	for i := 0; i < len(s); i++ {
		h = ((h << 5) + h) + uint32(s[i])
	}
	return formatHex(h % 16777215)
}

// formatHex formats a number as a 6-digit hex string
func formatHex(n uint32) string {
	const hex = "0123456789abcdef"
	b := make([]byte, 6)
	for i := 5; i >= 0; i-- {
		b[i] = hex[n%16]
		n /= 16
	}
	return string(b)
}

// ==============================================================================
// TEMPLATE & ASSET LOADER HELPERS
// ==============================================================================

// GetTemplatesDir finds the directory containing HTML templates
// It tries multiple path configurations to handle different execution contexts
func GetTemplatesDir() (string, error) {
	// Define canonical paths in priority order
	possibleDirs := []string{
		// When running from project root (most common during development)
		"internal/web/templates",
		// Fallback: explicit relative path construction
		filepath.Join(".", "internal", "web", "templates"),
	}

	var cwd string
	if wd, err := os.Getwd(); err == nil {
		cwd = wd
	}

	// Try each path
	for _, dir := range possibleDirs {
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			return dir, nil
		}
	}

	// Enhanced error message with debugging info
	return "", fmt.Errorf(
		"templates directory not found. Current working directory: %s. Tried paths: %v",
		cwd, possibleDirs,
	)
}

// findAndReadFile searches for a file in a list of possible relative paths and reads it.
func findAndReadFile(possiblePaths []string) ([]byte, string, error) {
	for _, path := range possiblePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			return content, path, nil
		}
	}
	return nil, "", fmt.Errorf("file not found in any of the paths: %v", possiblePaths)
}

// LoadEvolutionData reads the evolution.yml file and parses it into EvolutionData struct
func LoadEvolutionData() (schema.EvolutionData, error) {
	possiblePaths := []string{
		"internal/web/content/evolution.yml",
		filepath.Join(".", "internal", "web", "content", "evolution.yml"),
	}

	var data schema.EvolutionData

	content, _, err := findAndReadFile(possiblePaths)
	if err != nil {
		return schema.EvolutionData{}, fmt.Errorf("evolution.yml not found. Tried paths: %v", possiblePaths)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return schema.EvolutionData{}, fmt.Errorf("failed to parse evolution.yml: %w", err)
	}

	// Post-process descriptions into lines for each chapter's timeline
	for c := range data.Chapters {
		for i := range data.Chapters[c].Timeline {
			lines := strings.Split(strings.TrimSpace(data.Chapters[c].Timeline[i].Description), "\n")
			data.Chapters[c].Timeline[i].DescriptionLines = make([]string, 0, len(lines))
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Remove leading "- " if present
				line = strings.TrimPrefix(line, "- ")
				line = strings.TrimSpace(line)
				// Remove surrounding quotes if present
				if len(line) >= 2 && line[0] == '"' && line[len(line)-1] == '"' {
					line = line[1 : len(line)-1]
				}
				data.Chapters[c].Timeline[i].DescriptionLines = append(data.Chapters[c].Timeline[i].DescriptionLines, line)
			}
		}
	}

	return data, nil
}

// LoadLanding reads the landing.yml file and parses it into Landing struct
func LoadLanding() (schema.Landing, error) {
	possiblePaths := []string{
		"internal/web/content/landing.yml",
		filepath.Join(".", "internal", "web", "content", "landing.yml"),
	}

	var data schema.Landing

	content, _, err := findAndReadFile(possiblePaths)
	if err != nil {
		return schema.Landing{}, fmt.Errorf("landing.yml not found. Tried paths: %v", possiblePaths)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return schema.Landing{}, fmt.Errorf("failed to parse landing.yml: %w", err)
	}

	return data, nil
}

// ==============================================================================
// METRICS FORMATTING HELPERS
// ==============================================================================

var shortMonthNames = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// PrepareReadUnreadByYear creates JSON data for read/unread yearly breakdown chart
func PrepareReadUnreadByYear(metrics schema.Metrics) template.JS {
	// Get sorted years in descending order (latest first)
	years := make([]string, 0)
	for year := range metrics.ByYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	readByYearArray := make([]int, 0)
	unreadByYearArray := make([]int, 0)

	for _, year := range years {
		yearRead := 0
		yearUnread := 0

		// Sum up read/unread from all months in this year
		if yearMonthData, exists := metrics.ByYearAndMonth[year]; exists {
			for month, count := range yearMonthData {
				yearRead += count
				// Get unread for this month (if available, otherwise calculate from total)
				if monthUnread, unreadExists := metrics.UnreadByMonth[month]; unreadExists {
					yearUnread += monthUnread
				}
			}
		}

		readByYearArray = append(readByYearArray, yearRead)
		unreadByYearArray = append(unreadByYearArray, yearUnread)
	}

	data := map[string]interface{}{
		"labels":     years,
		"readData":   readByYearArray,
		"unreadData": unreadByYearArray,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// PrepareReadUnreadByMonth creates JSON data for read/unread monthly breakdown chart
func PrepareReadUnreadByMonth(metrics schema.Metrics) template.JS {
	readByMonthArray := make([]int, 12)
	unreadByMonthArray := make([]int, 12)

	for month := 1; month <= 12; month++ {
		monthStr := fmt.Sprintf("%02d", month)
		unread := 0
		if u, exists := metrics.UnreadByMonth[monthStr]; exists {
			unread = u
		}
		// Calculate read for this month
		read := 0
		if monthData, exists := metrics.ByMonth[monthStr]; exists {
			read = monthData - unread
		}
		readByMonthArray[month-1] = read
		unreadByMonthArray[month-1] = unread
	}

	data := map[string]interface{}{
		"labels":     shortMonthNames,
		"readData":   readByMonthArray,
		"unreadData": unreadByMonthArray,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// PrepareReadUnreadBySource creates JSON data for read/unread by source chart
func PrepareReadUnreadBySource(sources []schema.SourceInfo) template.JS {
	readUnreadBySourceLabels := make([]string, 0)
	readBySourceData := make([]int, 0)
	unreadBySourceData := make([]int, 0)
	for _, source := range sources {
		readUnreadBySourceLabels = append(readUnreadBySourceLabels, source.Name)
		readBySourceData = append(readBySourceData, source.Read)
		unreadBySourceData = append(unreadBySourceData, source.Unread)
	}

	data := map[string]interface{}{
		"labels":     readUnreadBySourceLabels,
		"readData":   readBySourceData,
		"unreadData": unreadBySourceData,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}

// PrepareUnreadArticleAgeDistribution creates JSON data for unread articles by age chart
func PrepareUnreadArticleAgeDistribution(metrics schema.Metrics) template.JS {
	// Define age bucket labels in display order
	bucketLabels := []struct {
		key   string
		label string
	}{
		{"less_than_1_month", "Less than 1 month"},
		{"1_to_3_months", "1-3 months"},
		{"3_to_6_months", "3-6 months"},
		{"6_to_12_months", "6-12 months"},
		{"older_than_1year", "Older than 1 year"},
	}

	labels := make([]string, 0)
	data := make([]int, 0)

	for _, bucket := range bucketLabels {
		labels = append(labels, bucket.label)
		count := metrics.UnreadArticleAgeDistribution[bucket.key]
		data = append(data, count)
	}

	chartData := map[string]interface{}{
		"labels": labels,
		"data":   data,
	}
	jsonData, _ := json.Marshal(chartData)
	return template.JS(jsonData)
}

// PrepareUnreadByYear creates JSON data for unread articles by year chart
func PrepareUnreadByYear(metrics schema.Metrics) template.JS {
	// Get sorted years in descending order (latest first)
	var years []string
	for year := range metrics.UnreadByYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(years)))

	unreadData := make([]int, 0)
	for _, year := range years {
		unreadData = append(unreadData, metrics.UnreadByYear[year])
	}

	data := map[string]interface{}{
		"labels": years,
		"data":   unreadData,
	}
	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}
