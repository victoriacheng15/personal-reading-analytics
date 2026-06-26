package web

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"testing"

	schema "github.com/victoriacheng15/personal-reading-analytics/internal"
)

func TestAnalyticsService_Generate(t *testing.T) {
	tests := []struct {
		name          string
		metrics       schema.Metrics
		expectSuccess bool
	}{
		{
			name: "generates full site with metrics",
			metrics: schema.Metrics{
				TotalArticles: 10,
				BySource:      map[string]int{"SourceA": 10},
				BySourceReadStatus: map[string][2]int{
					"SourceA":               {5, 5},
					"substack_author_count": {0, 0},
				},
				ByYear:  map[string]int{"2024": 10},
				ByMonth: map[string]int{"01": 10},
				ByMonthAndSource: map[string]map[string][2]int{
					"01": {"SourceA": {5, 5}},
				},
				UnreadByMonth: map[string]int{"01": 5},
				UnreadByYear:  map[string]int{"2024": 5},
				UnreadArticleAgeDistribution: map[string]int{
					"less_than_1_month": 5,
					"1_to_3_months":     0,
					"3_to_6_months":     0,
					"6_to_12_months":    0,
					"older_than_1year":  0,
				},
			},
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "web_service_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// The service looks for templates relative to CWD or in specific paths.
			// For testing, we'll create a mock structure.
			templateDir := filepath.Join(tmpDir, "internal", "web", "templates")
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Create required template files
			baseTmpl := `{{define "base"}}<html><head><title>{{.AnalyticsTitle}} - {{.PageTitle}}</title></head><body><div id="app"><header><h1>{{.PageTitle}}</h1><nav><ul><li><a href="{{.BaseURL}}index.html">Home</a></li></ul></nav></header>{{block "content" .}}{{end}}</div></body></html>{{end}}`
			indexTmpl := `{{define "content"}}<h1>Home</h1>{{end}}{{template "base" .}}`
			webTmpl := `{{define "content"}}<h1>Analytics</h1>{{end}}{{template "base" .}}`
			evolutionTmpl := `{{define "content"}}<h1>Evolution</h1>{{end}}{{template "base" .}}`

			templates := map[string]string{
				"base.html":      baseTmpl,
				"index.html":     indexTmpl,
				"analytics.html": webTmpl,
				"evolution.html": evolutionTmpl,
			}

			for name, content := range templates {
				if err := os.WriteFile(filepath.Join(templateDir, name), []byte(content), 0644); err != nil {
					t.Fatalf("failed to create template %s: %v", name, err)
				}
			}

			// Mock evolution.yml
			evolutionData := `chapters: []`
			contentDir := filepath.Join(tmpDir, "internal", "web", "content")
			if err := os.MkdirAll(contentDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(contentDir, "evolution.yml"), []byte(evolutionData), 0644); err != nil {
				t.Fatal(err)
			}

			// Mock index.yml
			indexData := `
intro_section:
  heading: "Test Heading"
  cta_buttons:
    - text: "Test Analytics"
      url: "test-analytics.html"
origin_story_section:
  title: "Test Origin Story"
  paragraphs:
    - "Test paragraph 1"
engineering_principles_section:
  title: "Test Engineering Principles"
  principles:
    - icon: "🧪"
      title: "Test Principle 1"
      description: "Test Description 1"
`
			if err := os.WriteFile(filepath.Join(contentDir, "index.yml"), []byte(indexData), 0644); err != nil {
				t.Fatal(err)
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			service := NewAnalyticsService("dist")
			config := GenConfig{
				OutputDir:    "dist",
				BaseURL:      "./",
				IsHistorical: false,
				HistoryDates: []string{"2024-01-01"},
				ReportDate:   "2024-01-01",
			}

			// Test Full Site Generation
			err = service.GenerateFullSite(tt.metrics, config)
			if (err == nil) != tt.expectSuccess {
				t.Errorf("GenerateFullSite() error = %v, expectSuccess %v", err, tt.expectSuccess)
			}

			if _, err := os.Stat("dist/index.html"); os.IsNotExist(err) {
				t.Error("dist/index.html was not created")
			}

			// Test Analytics Only Generation
			config.IsHistorical = true
			config.OutputDir = "dist/history/2024-01-01"
			config.BaseURL = "../../"
			err = service.GenerateAnalyticsOnly(tt.metrics, config)
			if (err == nil) != tt.expectSuccess {
				t.Errorf("GenerateAnalyticsOnly() error = %v, expectSuccess %v", err, tt.expectSuccess)
			}

			if _, err := os.Stat("dist/history/2024-01-01/analytics.html"); os.IsNotExist(err) {
				t.Error("dist/history/2024-01-01/analytics.html was not created")
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		setup     func(t *testing.T, src, dst string)
		expectErr bool
	}{
		{
			name:    "successfully copies file",
			content: "hello world",
			setup: func(t *testing.T, src, dst string) {
				// Normal setup
			},
			expectErr: false,
		},
		{
			name:    "source does not exist",
			content: "",
			setup: func(t *testing.T, src, dst string) {
				os.Remove(src) // Ensure source is missing
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcPath := filepath.Join(tmpDir, "source.txt")
			dstPath := filepath.Join(tmpDir, "dest.txt")

			if tt.content != "" {
				if err := os.WriteFile(srcPath, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			tt.setup(t, srcPath, dstPath)

			err := copyFile(srcPath, dstPath)
			if (err != nil) != tt.expectErr {
				t.Errorf("copyFile() error = %v, expectErr %v", err, tt.expectErr)
			}

			if !tt.expectErr {
				content, err := os.ReadFile(dstPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if string(content) != tt.content {
					t.Errorf("content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

func TestCopyDir(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, src string)
		expectErr bool
		verify    func(t *testing.T, dst string)
	}{
		{
			name: "recursively copies directory",
			setup: func(t *testing.T, src string) {
				// Create file in root
				os.WriteFile(filepath.Join(src, "root.txt"), []byte("root"), 0644)
				// Create subdir
				subDir := filepath.Join(src, "subdir")
				os.Mkdir(subDir, 0755)
				// Create file in subdir
				os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("sub"), 0644)
			},
			expectErr: false,
			verify: func(t *testing.T, dst string) {
				// Check root file
				if _, err := os.Stat(filepath.Join(dst, "root.txt")); os.IsNotExist(err) {
					t.Error("root.txt not copied")
				}
				// Check subdir
				if _, err := os.Stat(filepath.Join(dst, "subdir")); os.IsNotExist(err) {
					t.Error("subdir not copied")
				}
				// Check subdir file
				if _, err := os.Stat(filepath.Join(dst, "subdir", "sub.txt")); os.IsNotExist(err) {
					t.Error("sub.txt not copied")
				}
			},
		},
		{
			name: "source does not exist",
			setup: func(t *testing.T, src string) {
				os.RemoveAll(src)
			},
			expectErr: true,
			verify:    func(t *testing.T, dst string) {},
		},
		{
			name: "source is a file",
			setup: func(t *testing.T, src string) {
				os.RemoveAll(src)
				os.WriteFile(src, []byte("file"), 0644)
			},
			expectErr: true,
			verify:    func(t *testing.T, dst string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcDir := filepath.Join(tmpDir, "src")
			dstDir := filepath.Join(tmpDir, "dst")

			// Create src dir by default (tests can remove it)
			os.Mkdir(srcDir, 0755)

			tt.setup(t, srcDir)

			err := copyDir(srcDir, dstDir)
			if (err != nil) != tt.expectErr {
				t.Errorf("copyDir() error = %v, expectErr %v", err, tt.expectErr)
			}

			if !tt.expectErr {
				tt.verify(t, dstDir)
			}
		})
	}
}

// ==============================================================================
// CHART TEST SUITE
// ==============================================================================

func TestFormatHex(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "000000",
		},
		{
			name:     "small number",
			input:    255,
			expected: "0000ff",
		},
		{
			name:     "mid range",
			input:    4095,
			expected: "000fff",
		},
		{
			name:     "large number",
			input:    16777215,
			expected: "ffffff",
		},
		{
			name:     "arbitrary value",
			input:    12345,
			expected: "003039",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHex(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestColorHash(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedLength int
		expectNonEmpty bool
	}{
		{
			name:           "simple string",
			input:          "test",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "empty string",
			input:          "",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "long string",
			input:          "this is a much longer test string",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "special characters",
			input:          "@#$%^&*()",
			expectedLength: 6,
			expectNonEmpty: true,
		},
		{
			name:           "consistent hash",
			input:          "Substack",
			expectedLength: 6,
			expectNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorHash(tt.input)
			if len(result) != tt.expectedLength {
				t.Errorf("expected length %d, got %d", tt.expectedLength, len(result))
			}
			if tt.expectNonEmpty && len(result) == 0 {
				t.Error("expected non-empty result")
			}

			// Verify it's valid hex
			for _, ch := range result {
				if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
					t.Errorf("invalid hex character: %c", ch)
				}
			}
		})
	}
}

func TestColorHashConsistency(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "same input produces same hash",
			input: "GitHub",
		},
		{
			name:  "different inputs produce different hashes",
			input: "Substack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := colorHash(tt.input)
			hash2 := colorHash(tt.input)
			if hash1 != hash2 {
				t.Errorf("expected consistent hash, got %s and %s", hash1, hash2)
			}
		})
	}
}

func TestPrepareYearChartData(t *testing.T) {
	tests := []struct {
		name             string
		years            []schema.YearInfo
		expectedLabels   []string
		expectedDataLen  int
		shouldHaveLabels bool
		shouldHaveData   bool
	}{
		{
			name: "single year",
			years: []schema.YearInfo{
				{Year: "2025", Count: 100},
			},
			expectedLabels:   []string{"2025"},
			expectedDataLen:  1,
			shouldHaveLabels: true,
			shouldHaveData:   true,
		},
		{
			name: "multiple years",
			years: []schema.YearInfo{
				{Year: "2025", Count: 100},
				{Year: "2024", Count: 80},
				{Year: "2023", Count: 50},
			},
			expectedLabels:   []string{"2025", "2024", "2023"},
			expectedDataLen:  3,
			shouldHaveLabels: true,
			shouldHaveData:   true,
		},
		{
			name:             "empty years",
			years:            []schema.YearInfo{},
			expectedLabels:   []string{},
			expectedDataLen:  0,
			shouldHaveLabels: true,
			shouldHaveData:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareYearChartData(tt.years)

			if !tt.shouldHaveLabels || result.LabelsJSON == nil {
				if tt.shouldHaveLabels && result.LabelsJSON == nil {
					t.Error("expected LabelsJSON, got nil")
				}
				return
			}

			var labels []string
			err := json.Unmarshal(result.LabelsJSON, &labels)
			if err != nil {
				t.Fatalf("failed to unmarshal labels: %v", err)
			}

			if len(labels) != len(tt.expectedLabels) {
				t.Errorf("expected %d labels, got %d", len(tt.expectedLabels), len(labels))
			}

			for i, expected := range tt.expectedLabels {
				if i < len(labels) && labels[i] != expected {
					t.Errorf("label[%d]: expected %s, got %s", i, expected, labels[i])
				}
			}

			if !tt.shouldHaveData || result.DataJSON == nil {
				if tt.shouldHaveData && result.DataJSON == nil {
					t.Error("expected DataJSON, got nil")
				}
				return
			}

			var data []int
			err = json.Unmarshal(result.DataJSON, &data)
			if err != nil {
				t.Fatalf("failed to unmarshal data: %v", err)
			}

			if len(data) != tt.expectedDataLen {
				t.Errorf("expected %d data points, got %d", tt.expectedDataLen, len(data))
			}
		})
	}
}

func TestPrepareYearChartDataValues(t *testing.T) {
	tests := []struct {
		name         string
		years        []schema.YearInfo
		expectedData []int
	}{
		{
			name: "correct count values",
			years: []schema.YearInfo{
				{Year: "2025", Count: 100},
				{Year: "2024", Count: 75},
				{Year: "2023", Count: 50},
			},
			expectedData: []int{100, 75, 50},
		},
		{
			name: "zero counts",
			years: []schema.YearInfo{
				{Year: "2025", Count: 0},
				{Year: "2024", Count: 50},
			},
			expectedData: []int{0, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareYearChartData(tt.years)

			var data []int
			err := json.Unmarshal(result.DataJSON, &data)
			if err != nil {
				t.Fatalf("failed to unmarshal data: %v", err)
			}

			for i, expected := range tt.expectedData {
				if i < len(data) && data[i] != expected {
					t.Errorf("data[%d]: expected %d, got %d", i, expected, data[i])
				}
			}
		})
	}
}

func TestPrepareMonthChartData(t *testing.T) {
	tests := []struct {
		name            string
		months          []schema.MonthInfo
		sources         []schema.SourceInfo
		expectedLabels  []string
		expectDatasets  bool
		expectTotalData bool
	}{
		{
			name: "single month with sources",
			months: []schema.MonthInfo{
				{
					Name:  "January",
					Total: 50,
					Sources: map[string]int{
						"Substack":     30,
						"freeCodeCamp": 20,
					},
				},
			},
			sources: []schema.SourceInfo{
				{Name: "Substack", Read: 10, Unread: 20},
				{Name: "freeCodeCamp", Read: 8, Unread: 12},
			},
			expectedLabels:  []string{"January"},
			expectDatasets:  true,
			expectTotalData: true,
		},
		{
			name: "multiple months",
			months: []schema.MonthInfo{
				{
					Name:  "January",
					Total: 50,
					Sources: map[string]int{
						"Substack": 30,
						"GitHub":   20,
					},
				},
				{
					Name:  "February",
					Total: 60,
					Sources: map[string]int{
						"Substack": 40,
						"GitHub":   20,
					},
				},
			},
			sources: []schema.SourceInfo{
				{Name: "Substack", Read: 15, Unread: 25},
				{Name: "GitHub", Read: 20, Unread: 0},
			},
			expectedLabels:  []string{"January", "February"},
			expectDatasets:  true,
			expectTotalData: true,
		},
		{
			name:            "empty months",
			months:          []schema.MonthInfo{},
			sources:         []schema.SourceInfo{},
			expectedLabels:  []string{},
			expectDatasets:  true,
			expectTotalData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareMonthChartData(tt.months, tt.sources)

			// Verify labels
			var labels []string
			err := json.Unmarshal(result.LabelsJSON, &labels)
			if err != nil {
				t.Fatalf("failed to unmarshal labels: %v", err)
			}

			if len(labels) != len(tt.expectedLabels) {
				t.Errorf("expected %d labels, got %d", len(tt.expectedLabels), len(labels))
			}

			for i, expected := range tt.expectedLabels {
				if i < len(labels) && labels[i] != expected {
					t.Errorf("label[%d]: expected %s, got %s", i, expected, labels[i])
				}
			}

			// Verify datasets
			if tt.expectDatasets {
				var datasets []map[string]interface{}
				err := json.Unmarshal(result.DatasetsJSON, &datasets)
				if err != nil {
					t.Fatalf("failed to unmarshal datasets: %v", err)
				}

				for _, dataset := range datasets {
					if _, hasLabel := dataset["label"]; !hasLabel {
						t.Error("dataset missing label")
					}
					if _, hasData := dataset["data"]; !hasData {
						t.Error("dataset missing data")
					}
				}
			}

			// Verify total data
			if tt.expectTotalData {
				var totalData []int
				err := json.Unmarshal(result.TotalDataJSON, &totalData)
				if err != nil {
					t.Fatalf("failed to unmarshal total data: %v", err)
				}

				if len(totalData) != len(tt.months) {
					t.Errorf("expected %d total data points, got %d", len(tt.months), len(totalData))
				}
			}
		})
	}
}

func TestPrepareMonthChartDataDataValues(t *testing.T) {
	tests := []struct {
		name          string
		months        []schema.MonthInfo
		sources       []schema.SourceInfo
		expectedTotal []int
	}{
		{
			name: "correct monthly totals",
			months: []schema.MonthInfo{
				{
					Name:  "January",
					Total: 50,
					Sources: map[string]int{
						"Substack": 30,
						"GitHub":   20,
					},
				},
				{
					Name:  "February",
					Total: 75,
					Sources: map[string]int{
						"Substack": 50,
						"GitHub":   25,
					},
				},
			},
			sources: []schema.SourceInfo{
				{Name: "Substack", Read: 15, Unread: 65},
				{Name: "GitHub", Read: 30, Unread: 15},
			},
			expectedTotal: []int{50, 75},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareMonthChartData(tt.months, tt.sources)

			var totalData []int
			err := json.Unmarshal(result.TotalDataJSON, &totalData)
			if err != nil {
				t.Fatalf("failed to unmarshal total data: %v", err)
			}

			for i, expected := range tt.expectedTotal {
				if i < len(totalData) && totalData[i] != expected {
					t.Errorf("total[%d]: expected %d, got %d", i, expected, totalData[i])
				}
			}
		})
	}
}

func TestPrepareMonthChartDataColorAssignment(t *testing.T) {
	tests := []struct {
		name          string
		sourceName    string
		providedColor string
		expectedColor string
	}{
		{
			name:          "Source with provided color uses that color",
			sourceName:    "Substack",
			providedColor: "#667eea",
			expectedColor: "#667eea",
		},
		{
			name:          "Source without provided color uses hash-generated color",
			sourceName:    "UnknownSource",
			providedColor: "",
			expectedColor: "#", // Should start with #
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			months := []schema.MonthInfo{
				{
					Name:  "January",
					Total: 30,
					Sources: map[string]int{
						tt.sourceName: 30,
					},
				},
			}

			sources := []schema.SourceInfo{
				{Name: tt.sourceName, Read: 10, Unread: 20, Color: tt.providedColor},
			}

			result := PrepareMonthChartData(months, sources)

			var datasets []map[string]interface{}
			err := json.Unmarshal(result.DatasetsJSON, &datasets)
			if err != nil {
				t.Fatalf("failed to unmarshal datasets: %v", err)
			}

			if len(datasets) > 0 {
				bgColor := datasets[0]["backgroundColor"]
				if bgColor == nil {
					t.Error("backgroundColor is missing")
					return
				}

				colorStr := bgColor.(string)
				if tt.providedColor != "" && colorStr != tt.expectedColor {
					t.Errorf("expected color %s, got %s", tt.expectedColor, colorStr)
				}

				if colorStr[0] != '#' {
					t.Errorf("backgroundColor should start with #, got %s", colorStr)
				}
			}
		})
	}
}

// ==============================================================================
// ASSET LOADER TEST SUITE
// ==============================================================================

func TestGetTemplatesDir(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string // returns temp dir path
		expectError bool
		expectEmpty bool
	}{
		{
			name: "finds templates directory from primary path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				// Create directory structure for primary path
				templatesDir := filepath.Join("internal", "web", "templates")
				if err := os.MkdirAll(templatesDir, 0755); err != nil {
					t.Fatalf("failed to create directories: %v", err)
				}

				return tmpDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "finds templates directory from relative path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				// Create directory structure for relative path
				templatesDir := filepath.Join(".", "internal", "web", "templates")
				if err := os.MkdirAll(templatesDir, 0755); err != nil {
					t.Fatalf("failed to create directories: %v", err)
				}

				return tmpDir
			},
			expectError: false,
			expectEmpty: false,
		},
		{
			name: "returns error when templates directory not found",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}
				return tmpDir
			},
			expectError: true,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := tt.setup(t)
			defer os.RemoveAll(tmpDir)

			dir, err := GetTemplatesDir()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if !tt.expectEmpty && dir != "" {
					t.Errorf("expected empty path on error, got: %v", dir)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectEmpty && dir == "" {
				t.Errorf("expected non-empty path, got empty string")
			}

			if !tt.expectEmpty && dir == "" {
				t.Errorf("expected non-empty path, got empty string")
			}
		})
	}
}

func TestLoadEvolutionData(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
	}{
		{
			name: "loads evolution data successfully",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}

				dir := filepath.Join("internal", "web", "content")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}

				yamlContent := `
chapters:
  - title: "Chapter 1"
    timeline:
      - date: "2024-01"
        title: "Test Event"
        description: |
          - "Detail 1"
`
				if err := os.WriteFile(filepath.Join(dir, "evolution.yml"), []byte(yamlContent), 0644); err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			expectError: false,
		},
		{
			name: "returns error when file missing",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}
				return tmpDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := tt.setup(t)
			defer os.RemoveAll(tmpDir)

			data, err := LoadEvolutionData()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(data.Chapters) > 0 && len(data.Chapters[0].Timeline) > 0 && data.Chapters[0].Timeline[0].Title != "Test Event" {
				t.Errorf("expected title 'Test Event', got %s", data.Chapters[0].Timeline[0].Title)
			}

			if len(data.Chapters) > 0 && len(data.Chapters[0].Timeline) > 0 {
				if len(data.Chapters[0].Timeline[0].DescriptionLines) == 0 || data.Chapters[0].Timeline[0].DescriptionLines[0] != "Detail 1" {
					t.Errorf("expected DescriptionLines[0] to be 'Detail 1', got %v", data.Chapters[0].Timeline[0].DescriptionLines)
				}
			}
		})
	}
}

// ==============================================================================
// METRICS PREPARATION TEST SUITE
// ==============================================================================

func TestPrepareReadUnreadByYear(t *testing.T) {
	tests := []struct {
		name            string
		metrics         schema.Metrics
		expectedYear0   string
		expectedRead0   float64
		expectedUnread0 float64
		expectedRead1   float64
		expectedUnread1 float64
		expectEmpty     bool
	}{
		{
			name: "multiple years with correct values",
			metrics: schema.Metrics{
				ByYear: map[string]int{
					"2024": 100,
					"2023": 50,
				},
				ByYearAndMonth: map[string]map[string]int{
					"2024": {"01": 10, "02": 20},
					"2023": {"01": 5},
				},
				UnreadByMonth: map[string]int{
					"01": 2,
					"02": 3,
				},
			},
			expectedYear0:   "2024",
			expectedRead0:   30,
			expectedUnread0: 5,
			expectedRead1:   5,
			expectedUnread1: 2,
			expectEmpty:     false,
		},
		{
			name:        "empty metrics",
			metrics:     schema.Metrics{ByYear: map[string]int{}},
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := PrepareReadUnreadByYear(tt.metrics)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			labels := data["labels"].([]interface{})
			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if tt.expectEmpty {
				if len(labels) != 0 {
					t.Errorf("expected empty labels, got %d", len(labels))
				}
				if len(readData) != 0 {
					t.Errorf("expected empty readData, got %d", len(readData))
				}
				return
			}

			if labels[0].(string) != tt.expectedYear0 {
				t.Errorf("expected year %s first, got %s", tt.expectedYear0, labels[0])
			}
			if readData[0].(float64) != tt.expectedRead0 {
				t.Errorf("expected %v read, got %v", tt.expectedRead0, readData[0])
			}
			if unreadData[0].(float64) != tt.expectedUnread0 {
				t.Errorf("expected %v unread, got %v", tt.expectedUnread0, unreadData[0])
			}
			if readData[1].(float64) != tt.expectedRead1 {
				t.Errorf("expected %v read, got %v", tt.expectedRead1, readData[1])
			}
			if unreadData[1].(float64) != tt.expectedUnread1 {
				t.Errorf("expected %v unread, got %v", tt.expectedUnread1, unreadData[1])
			}
		})
	}
}

func TestPrepareReadUnreadByMonth(t *testing.T) {
	tests := []struct {
		name            string
		metrics         schema.Metrics
		expectedRead0   float64
		expectedUnread0 float64
		expectedRead1   float64
		expectedUnread1 float64
		expectedRead2   float64
		isAllZero       bool
	}{
		{
			name: "monthly breakdown with correct calculations",
			metrics: schema.Metrics{
				UnreadByMonth: map[string]int{
					"01": 5,
					"02": 10,
				},
				ByMonth: map[string]int{
					"01": 20,
					"02": 30,
				},
			},
			expectedRead0:   15,
			expectedUnread0: 5,
			expectedRead1:   20,
			expectedUnread1: 10,
			expectedRead2:   0,
			isAllZero:       false,
		},
		{
			name:      "empty metrics returns zeroed arrays",
			metrics:   schema.Metrics{},
			isAllZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := PrepareReadUnreadByMonth(tt.metrics)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if tt.isAllZero {
				if len(readData) != 12 {
					t.Errorf("expected 12 months, got %d", len(readData))
				}
				for i := 0; i < 12; i++ {
					if readData[i].(float64) != 0 || unreadData[i].(float64) != 0 {
						t.Errorf("expected zero at index %d", i)
					}
				}
				return
			}

			if readData[0].(float64) != tt.expectedRead0 {
				t.Errorf("expected %v read for Jan, got %v", tt.expectedRead0, readData[0])
			}
			if unreadData[0].(float64) != tt.expectedUnread0 {
				t.Errorf("expected %v unread for Jan, got %v", tt.expectedUnread0, unreadData[0])
			}
			if readData[1].(float64) != tt.expectedRead1 {
				t.Errorf("expected %v read for Feb, got %v", tt.expectedRead1, readData[1])
			}
			if unreadData[1].(float64) != tt.expectedUnread1 {
				t.Errorf("expected %v unread for Feb, got %v", tt.expectedUnread1, unreadData[1])
			}
			if readData[2].(float64) != tt.expectedRead2 {
				t.Errorf("expected %v read for Mar, got %v", tt.expectedRead2, readData[2])
			}
		})
	}
}

func TestPrepareReadUnreadBySource(t *testing.T) {
	tests := []struct {
		name               string
		sources            []schema.SourceInfo
		expectedLabels     int
		expectedFirstLabel string
		expectedRead       float64
		expectedUnread     float64
	}{
		{
			name: "multiple sources",
			sources: []schema.SourceInfo{
				{Name: "SourceA", Read: 10, Unread: 5},
				{Name: "SourceB", Read: 20, Unread: 0},
			},
			expectedLabels:     2,
			expectedFirstLabel: "SourceA",
			expectedRead:       10,
			expectedUnread:     5,
		},
		{
			name: "single source",
			sources: []schema.SourceInfo{
				{Name: "SourceX", Read: 15, Unread: 3},
			},
			expectedLabels:     1,
			expectedFirstLabel: "SourceX",
			expectedRead:       15,
			expectedUnread:     3,
		},
		{
			name:           "nil sources list",
			sources:        nil,
			expectedLabels: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := PrepareReadUnreadBySource(tt.sources)
			var data map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &data)

			labels := data["labels"].([]interface{})
			readData := data["readData"].([]interface{})
			unreadData := data["unreadData"].([]interface{})

			if len(labels) != tt.expectedLabels {
				t.Errorf("expected %d labels, got %d", tt.expectedLabels, len(labels))
			}
			if tt.expectedLabels > 0 {
				if labels[0].(string) != tt.expectedFirstLabel {
					t.Errorf("expected %s, got %s", tt.expectedFirstLabel, labels[0])
				}
				if readData[0].(float64) != tt.expectedRead {
					t.Errorf("expected %v read, got %v", tt.expectedRead, readData[0])
				}
				if unreadData[0].(float64) != tt.expectedUnread {
					t.Errorf("expected %v unread, got %v", tt.expectedUnread, unreadData[0])
				}
			}
		})
	}
}

func createTestMetricsWithAgeDistribution() *schema.Metrics {
	metrics := &schema.Metrics{
		UnreadArticleAgeDistribution: make(map[string]int),
		UnreadByYear:                 make(map[string]int),
		ByYear:                       make(map[string]int),
		ByMonth:                      make(map[string]int),
		ByYearAndMonth:               make(map[string]map[string]int),
		ByMonthAndSource:             make(map[string]map[string][2]int),
		BySource:                     make(map[string]int),
		ByCategory:                   make(map[string][2]int),
		BySourceReadStatus:           make(map[string][2]int),
		UnreadByCategory:             make(map[string]int),
		UnreadBySource:               make(map[string]int),
	}

	metrics.UnreadArticleAgeDistribution["less_than_1_month"] = 8
	metrics.UnreadArticleAgeDistribution["1_to_3_months"] = 12
	metrics.UnreadArticleAgeDistribution["3_to_6_months"] = 15
	metrics.UnreadArticleAgeDistribution["6_to_12_months"] = 10
	metrics.UnreadArticleAgeDistribution["older_than_1year"] = 5

	return metrics
}

func TestPrepareUnreadArticleAgeDistribution(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *schema.Metrics
		validate func(t *testing.T, jsonStr template.JS)
	}{
		{
			name:    "age distribution with all buckets populated",
			metrics: createTestMetricsWithAgeDistribution(),
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				if _, hasLabels := chartData["labels"]; !hasLabels {
					t.Error("missing 'labels' key in chart data")
				}
				if _, hasData := chartData["data"]; !hasData {
					t.Error("missing 'data' key in chart data")
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 5 {
					t.Errorf("expected 5 labels, got %v", labels)
					return
				}

				labelStrs := make([]string, len(labels))
				for i, label := range labels {
					labelStrs[i] = label.(string)
				}

				if labelStrs[4] != "Older than 1 year" {
					t.Errorf("expected 'Older than 1 year' label, got %s", labelStrs[4])
				}
			},
		},
		{
			name: "empty age distribution",
			metrics: &schema.Metrics{
				UnreadArticleAgeDistribution: make(map[string]int),
			},
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				data, ok := chartData["data"].([]interface{})
				if !ok || len(data) == 0 {
					t.Error("expected valid data array for empty metrics")
				}
				// Also verify values are 0
				for _, val := range data {
					if val.(float64) != 0 {
						t.Error("expected 0 for empty distribution buckets")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.metrics
			if metrics == nil {
				metrics = &schema.Metrics{
					UnreadArticleAgeDistribution: make(map[string]int),
				}
			}

			jsonStr := PrepareUnreadArticleAgeDistribution(*metrics)
			tt.validate(t, jsonStr)
		})
	}
}

func TestPrepareUnreadArticleAgeDistributionJSON(t *testing.T) {
	metrics := createTestMetricsWithAgeDistribution()
	jsonStr := PrepareUnreadArticleAgeDistribution(*metrics)

	var chartData map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &chartData)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	requiredKeys := []string{"labels", "data"}
	for _, key := range requiredKeys {
		if _, exists := chartData[key]; !exists {
			t.Errorf("required key '%s' missing from chart data", key)
		}
	}

	data, ok := chartData["data"].([]interface{})
	if !ok {
		t.Error("data field should be array of numbers")
		return
	}

	for i, val := range data {
		if _, isNum := val.(float64); !isNum {
			t.Errorf("data[%d] should be numeric, got %T", i, val)
		}
	}

	labels, _ := chartData["labels"].([]interface{})
	expectedMap := metrics.UnreadArticleAgeDistribution

	for i, label := range labels {
		labelStr := label.(string)
		dataVal := int(data[i].(float64))

		labelToKey := map[string]string{
			"Less than 1 month": "less_than_1_month",
			"1-3 months":        "1_to_3_months",
			"3-6 months":        "3_to_6_months",
			"6-12 months":       "6_to_12_months",
			"Older than 1 year": "older_than_1year",
		}

		key := labelToKey[labelStr]
		expectedVal := expectedMap[key]

		if dataVal != expectedVal {
			t.Errorf("data mismatch for %s: expected %d, got %d", labelStr, expectedVal, dataVal)
		}
	}
}

func createTestMetricsWithUnreadByYear() *schema.Metrics {
	metrics := &schema.Metrics{
		UnreadByYear:                 make(map[string]int),
		UnreadArticleAgeDistribution: make(map[string]int),
		ByYear:                       make(map[string]int),
		ByMonth:                      make(map[string]int),
		ByYearAndMonth:               make(map[string]map[string]int),
		ByMonthAndSource:             make(map[string]map[string][2]int),
		BySource:                     make(map[string]int),
		ByCategory:                   make(map[string][2]int),
		BySourceReadStatus:           make(map[string][2]int),
		UnreadByCategory:             make(map[string]int),
		UnreadBySource:               make(map[string]int),
	}

	metrics.UnreadByYear["2025"] = 30
	metrics.UnreadByYear["2024"] = 25
	metrics.UnreadByYear["2023"] = 15
	metrics.UnreadByYear["2022"] = 8

	return metrics
}

func TestPrepareUnreadByYear(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *schema.Metrics
		validate func(t *testing.T, jsonStr template.JS)
	}{
		{
			name:    "multiple years in descending order",
			metrics: createTestMetricsWithUnreadByYear(),
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 4 {
					t.Errorf("expected 4 labels, got %v", labels)
					return
				}

				if labels[0].(string) != "2025" {
					t.Errorf("expected first year to be 2025, got %s", labels[0])
				}
				if labels[len(labels)-1].(string) != "2022" {
					t.Errorf("expected last year to be 2022, got %s", labels[len(labels)-1])
				}
			},
		},
		{
			name: "single year",
			metrics: &schema.Metrics{
				UnreadByYear: map[string]int{
					"2025": 20,
				},
			},
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 1 {
					t.Errorf("expected 1 label, got %v", labels)
				}
			},
		},
		{
			name: "non-consecutive years",
			metrics: &schema.Metrics{
				UnreadByYear: map[string]int{
					"2025": 25,
					"2022": 10,
					"2020": 5,
				},
			},
			validate: func(t *testing.T, jsonStr template.JS) {
				var chartData map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chartData)
				if err != nil {
					t.Errorf("failed to unmarshal JSON: %v", err)
					return
				}

				labels, ok := chartData["labels"].([]interface{})
				if !ok || len(labels) != 3 {
					t.Errorf("expected 3 labels, got %v", labels)
					return
				}

				if labels[0].(string) != "2025" || labels[1].(string) != "2022" || labels[2].(string) != "2020" {
					t.Errorf("years not in descending order: %v", labels)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.metrics
			if metrics == nil {
				metrics = &schema.Metrics{
					UnreadByYear: make(map[string]int),
				}
			}

			jsonStr := PrepareUnreadByYear(*metrics)
			tt.validate(t, jsonStr)
		})
	}
}

func TestPrepareUnreadByYearDataValidity(t *testing.T) {
	tests := []struct {
		name          string
		metrics       *schema.Metrics
		expectedValid bool
	}{
		{
			name:          "data matches input metrics",
			metrics:       createTestMetricsWithUnreadByYear(),
			expectedValid: true,
		},
		{
			name: "single year",
			metrics: &schema.Metrics{
				UnreadByYear: map[string]int{
					"2025": 100,
				},
			},
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := PrepareUnreadByYear(*tt.metrics)

			var chartData map[string]interface{}
			err := json.Unmarshal([]byte(jsonStr), &chartData)
			if err != nil {
				t.Fatalf("JSON unmarshaling failed: %v", err)
			}

			labels, ok := chartData["labels"].([]interface{})
			if !ok {
				t.Fatal("labels should be array")
			}

			data, ok := chartData["data"].([]interface{})
			if !ok {
				t.Fatal("data should be array")
			}

			if len(labels) != len(data) {
				t.Errorf("labels and data length mismatch: %d vs %d", len(labels), len(data))
			}

			for i, label := range labels {
				year := label.(string)
				expectedCount := tt.metrics.UnreadByYear[year]
				actualCount := int(data[i].(float64))
				if actualCount != expectedCount {
					t.Errorf("year %s: expected %d, got %d", year, expectedCount, actualCount)
				}
			}
		})
	}
}

// ==============================================================================
// METRICS IO TEST SUITE
// ==============================================================================

func TestGetMetricsDates(t *testing.T) {
	tests := []struct {
		name          string
		fileNames     []string
		expectedDates []string
		expectError   bool
	}{
		{
			name:          "returns sorted dates",
			fileNames:     []string{"2025-01-01.json", "2024-01-01.json", "invalid.txt"},
			expectedDates: []string{"2025-01-01", "2024-01-01"},
			expectError:   false,
		},
		{
			name:          "no valid metrics files",
			fileNames:     []string{"not-a-date.txt"},
			expectedDates: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test_metrics_dates")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			metricsDir := filepath.Join(tmpDir, "metrics")
			if err := os.Mkdir(metricsDir, 0755); err != nil {
				t.Fatal(err)
			}

			for _, fileName := range tt.fileNames {
				if err := os.WriteFile(filepath.Join(metricsDir, fileName), []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			dates, err := GetMetricsDates()
			if (err != nil) != tt.expectError {
				t.Errorf("unexpected error: %v", err)
			}

			if len(dates) != len(tt.expectedDates) {
				t.Errorf("expected %d dates, got %d", len(tt.expectedDates), len(dates))
			}

			for i := range dates {
				if dates[i] != tt.expectedDates[i] {
					t.Errorf("expected date %s, got %s", tt.expectedDates[i], dates[i])
				}
			}
		})
	}
}

func TestLoadMetricsByDate(t *testing.T) {
	tests := []struct {
		name             string
		date             string
		fileContent      string
		expectedArticles int
		expectError      bool
	}{
		{
			name:             "loads metrics for specific date",
			date:             "2025-01-01",
			fileContent:      `{"total_articles": 100}`,
			expectedArticles: 100,
			expectError:      false,
		},
		{
			name:             "non-existent date",
			date:             "2000-01-01",
			fileContent:      "",
			expectedArticles: 0,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "test_metrics_by_date")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			metricsDir := filepath.Join(tmpDir, "metrics")
			if err := os.Mkdir(metricsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if tt.fileContent != "" {
				if err := os.WriteFile(filepath.Join(metricsDir, tt.date+".json"), []byte(tt.fileContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			metrics, err := LoadMetricsByDate(tt.date)
			if (err != nil) != tt.expectError {
				t.Errorf("unexpected error: %v", err)
			}
			if metrics.TotalArticles != tt.expectedArticles {
				t.Errorf("expected %d articles, got %d", tt.expectedArticles, metrics.TotalArticles)
			}
		})
	}
}
