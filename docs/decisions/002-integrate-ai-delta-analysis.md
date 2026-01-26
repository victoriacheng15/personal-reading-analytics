# ADR 002: Integrate AI Delta Analysis

- **Status:** Accepted
- **Date:** 2026-01-23
- **Author:** Victoria Cheng

## Context and Problem Statement

The current analytics dashboard (`analytics.exe`) is stateless regarding history. It loads only the single most recent JSON metrics file (`loadLatestMetrics`) and renders a snapshot of the present moment. There is zero capability for the visitor to see week-over-week changes or trends (e.g., "Read rate increased by 5%"). The data exists on disk, but the product provides no interface to access it.

## Decision Outcome

The project will integrate **Google Gemini (Generative AI)** to perform an **AI Delta Analysis**, comparing raw metrics snapshots to generate a qualitative weekly narrative. This provides historical context that static charts cannot easily convey.

- **Mechanism:** The `metrics.exe` binary supports two distinct operational modes via flags:
  - **`--fetch` (Workflow A):** Connects to Google Sheets, calculates stats, and saves the raw `YYYY-MM-DD.json`.
  - **`--summarize` (Workflow B):** Reads the latest local JSON, compares it with the previous week's file, generates the AI delta analysis, and appends it to the *existing* JSON file.
- **Architecture:**
  - A new package `cmd/internal/ai` isolates external API interactions.
  - The `metrics` package remains the source of truth for data structure.
  - The default model is **`gemini-2.5-flash-lite`**.

## Consequences

- **Positive:**
  - AI-generated delta analyses provide high-signal narrative context without requiring a database or complex frontend charting.
  - `metrics` package remains the single source of truth.
- **Negative/Trade-offs:**
  - Introduces an external API dependency (Gemini).
  - The pipeline cannot fail solely due to the AI step (implemented as non-blocking warning).

## Verification

- [x] **Manual Check:** Run `go run cmd/metrics/main.go --summarize` and verify the `ai_delta_analysis` field in the generated JSON.
- [x] **Automated Tests:** Run `go test ./cmd/internal/metrics/...` to verify prompt construction and mock client interactions.
