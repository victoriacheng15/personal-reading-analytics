package web

import (
	"encoding/json"
	"html/template"
	"time"

	schema "github.com/victoriacheng15/personal-reading-analytics/internal"
)

// ==============================================================================
// CONFIGURATION MODELS
// ==============================================================================

// GenConfig holds configuration for a specific generation pass
type GenConfig struct {
	OutputDir    string
	BaseURL      string
	IsHistorical bool
	HistoryDates []string
	ReportDate   string
}

// ==============================================================================
// CHART DATA MODELS
// ==============================================================================

// ChartDataset represents a single dataset for Chart.js
type ChartDataset struct {
	Label           string      `json:"label"`
	Data            interface{} `json:"data"`
	BackgroundColor interface{} `json:"backgroundColor,omitempty"`
	BorderColor     string      `json:"borderColor,omitempty"`
	BorderWidth     int         `json:"borderWidth,omitempty"`
}

// YearChartData holds prepared year chart data
type YearChartData struct {
	LabelsJSON json.RawMessage
	DataJSON   json.RawMessage
}

// MonthChartData holds prepared month chart data
type MonthChartData struct {
	LabelsJSON    json.RawMessage
	DatasetsJSON  json.RawMessage
	TotalDataJSON json.RawMessage
}

// ==============================================================================
// VIEW MODELS
// ==============================================================================

// ViewModel represents the data structure passed to HTML templates
type ViewModel struct {
	AnalyticsTitle                   string
	PageTitle                        string
	KeyMetrics                       []schema.KeyMetric
	HighlightMetrics                 []schema.HightlightMetric
	TotalArticles                    int
	ReadCount                        int
	UnreadCount                      int
	ReadRate                         float64
	AvgArticlesPerMonth              float64
	LastUpdated                      time.Time
	AIDeltaAnalysis                  string
	Sources                          []schema.SourceInfo
	Months                           []schema.MonthInfo
	Years                            []schema.YearInfo
	AllYears                         []string
	AllSources                       []string
	AllYearsJSON                     template.JS
	AllSourcesJSON                   template.JS
	YearChartLabels                  template.JS
	YearChartData                    template.JS
	MonthChartLabels                 template.JS
	MonthChartDatasets               template.JS
	MonthTotalData                   template.JS
	ReadUnreadByMonthJSON            template.JS
	ReadUnreadBySourceJSON           template.JS
	ReadUnreadByYearJSON             template.JS
	UnreadArticleAgeDistributionJSON template.JS
	UnreadByYearJSON                 template.JS
	TopOldestUnreadArticles          []schema.ArticleMeta
	EvolutionData                    schema.EvolutionData
	Landing                          schema.Landing

	// Historical Metrics context
	BaseURL      string
	IsHistorical bool
	HistoryDates []string
	ReportDate   string
}
