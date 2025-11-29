package metrics

import (
	"testing"
	"time"
)

// Test NormalizeSourceName
func TestNormalizeSourceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"substack", "Substack"},
		{"SUBSTACK", "Substack"},
		{"Substack", "Substack"},
		{"freecodecamp", "freeCodeCamp"},
		{"FREECODECAMP", "freeCodeCamp"},
		{"github", "GitHub"},
		{"GITHUB", "GitHub"},
		{"shopify", "Shopify"},
		{"stripe", "Stripe"},
		{"Unknown", "Unknown"},
		{"medium", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeSourceName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSourceName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test calculateMonthsDifference
func TestCalculateMonthsDifference(t *testing.T) {
	tests := []struct {
		name     string
		earliest time.Time
		latest   time.Time
		expected int
	}{
		{
			name:     "same month",
			earliest: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "one month difference",
			earliest: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "multiple months",
			earliest: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			expected: 5,
		},
		{
			name:     "one year difference",
			earliest: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 12,
		},
		{
			name:     "multiple years",
			earliest: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			latest:   time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			expected: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMonthsDifference(tt.earliest, tt.latest)
			if result != tt.expected {
				t.Errorf("calculateMonthsDifference(%v, %v) = %d, want %d", tt.earliest, tt.latest, result, tt.expected)
			}
		})
	}
}

// Test parseArticleRow
func TestParseArticleRow(t *testing.T) {
	tests := []struct {
		name      string
		row       []interface{}
		expectErr bool
		validate  func(*ParsedArticle) bool
	}{
		{
			name: "valid article",
			row: []interface{}{
				"2025-11-28",
				"Article Title",
				"https://example.com",
				"Substack",
				"FALSE",
			},
			expectErr: false,
			validate: func(p *ParsedArticle) bool {
				return p.Date.Format("2006-01-02") == "2025-11-28" &&
					p.Category == "Substack" &&
					p.IsRead == false
			},
		},
		{
			name: "read article",
			row: []interface{}{
				"2025-11-27",
				"Article Title",
				"https://example.com",
				"GitHub",
				"TRUE",
			},
			expectErr: false,
			validate: func(p *ParsedArticle) bool {
				return p.Date.Format("2006-01-02") == "2025-11-27" &&
					p.Category == "GitHub" &&
					p.IsRead == true
			},
		},
		{
			name: "normalized source",
			row: []interface{}{
				"2025-11-26",
				"Article",
				"https://example.com",
				"freecodecamp",
				"TRUE",
			},
			expectErr: false,
			validate: func(p *ParsedArticle) bool {
				return p.Category == "freeCodeCamp"
			},
		},
		{
			name:      "incomplete row",
			row:       []interface{}{"2025-11-28", "Title"},
			expectErr: true,
			validate:  func(p *ParsedArticle) bool { return true },
		},
		{
			name: "invalid date",
			row: []interface{}{
				"invalid-date",
				"Title",
				"https://example.com",
				"Substack",
				"FALSE",
			},
			expectErr: true,
			validate:  func(p *ParsedArticle) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseArticleRow(tt.row)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseArticleRow() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if err == nil && !tt.validate(result) {
				t.Errorf("parseArticleRow() validation failed for result: %+v", result)
			}
		})
	}
}
