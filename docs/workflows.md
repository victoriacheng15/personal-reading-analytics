# Platform Workflows

This document details the CI/CD and automation paths that validate the Personal Reading Analytics pipeline.

---

## 📂 Core Workflows

### 🚢 [Continuous Integration](../.github/workflows/ci.yml)

The central pipeline that coordinates the parallel validation and linting of the analytics applications.

- **Trigger**: Push or Pull Request targeting the main branch.
- **Responsibility**: Detects modified files and triggers Go, Python, or Markdown checks conditionally.
- **Key Feature**: Leverages path-filtering to execute only the jobs corresponding to the modified components.

#### 🧪 Go Lint & Test

Ensures code quality and functional correctness across the Go metrics processor and static site generator packages.

- **Trigger**: File changes detected in the Go codebase.
- **Responsibility**: Validates syntax via `go vet`, checks code formatting, and executes the full suite of unit tests.
- **Key Feature**: Verification of Go formatting without writing files via custom checking script.

#### 🐍 Python Lint & Test

Validates the asynchronous configuration-driven extraction engine.

- **Trigger**: File changes detected in the Python codebase.
- **Responsibility**: Checks formatting and style conventions using Ruff, and executes unit tests via Pytest.
- **Key Feature**: Mocking of third-party APIs (Google Sheets, MongoDB) to verify offline extraction logic.

#### 📝 Markdown Linting

Enforces syntax and format consistency across all Markdown documentation files.

- **Trigger**: File changes detected in Markdown files.
- **Responsibility**: Scans documentation directories using `markdownlint-cli` via `npx` to enforce formatting rules.
- **Key Feature**: Automated layout verification protecting project operational memory documents.

### 📅 [Daily Ingestion](../.github/workflows/extraction.yml)

Coordinates daily content extraction from RSS/HTML feeds and logs them to the MongoDB event store.

- **Trigger**: Daily schedule at 6:00 AM UTC or manual trigger.
- **Responsibility**: Resolves feed updates and publishes event telemetry.
- **Key Feature**: Asynchronous concurrency handling multiple providers without blocking.

### 📊 [Weekly Metrics Generation](../.github/workflows/metrics_generation.yml)

Performs quantitative processing and generates delta narratives powered by Gemini.

- **Trigger**: Weekly schedule on Fridays at 1:00 AM UTC or manual trigger.
- **Responsibility**: Computes metrics and opens an automated pull request with the new historical dataset.
- **Key Feature**: Asynchronous integration with the Gemini API to write weekly narrative summaries.

### 🚀 [GitHub Pages Deployment](../.github/workflows/deployment.yml)

Compiles the latest static asset distributions and uploads them to the GitHub Pages environment.

- **Trigger**: Push to the main branch targeting the metrics dataset or web templates.
- **Responsibility**: Invokes the Go template generator and transpiles CSS assets before deployment.
- **Key Feature**: Uses GitHub Actions pages deployment role for zero-downtime hosting.
