# 📚 Personal Reading Analytics

## What is this?

This is a fully automated data pipeline that tracks and analyzes personal reading data.

It demonstrates how a real data + DevOps system would:
- collect and process data automatically
- generate analytics and visualizations
- run on CI/CD without servers
- produce both quantitative and AI-driven insights

The goal is to show how to build a zero-infrastructure analytics system that is automated, observable, and cost-efficient.

👉 [Live Dashboard](https://victoriacheng15.github.io/personal-reading-analytics/)  
📚 [Full Documentation](docs/README.md)

---

## 🔍 What I Built (Quick Proof)

- Fully automated data pipeline using GitHub Actions (no servers)
- Event-driven data ingestion using MongoDB (event sourcing pattern)
- Data extraction from Google Sheets API
- Interactive analytics dashboard using Chart.js
- Weekly AI-generated insights (Velocity, Backlog, Chronology)
- Historical snapshot tracking for trend analysis
- CI/CD pipeline with scheduled workflows
- Integrated with external observability platform (Observability Hub)
- Zero-cost infrastructure using free-tier cloud services

---

## 📦 Platform Projects

This system is built as a collection of smaller data and platform projects:

1. **Data Ingestion Pipeline**
   - Extracts reading data from Google Sheets

2. **Event Sourcing System**
   - Stores events in MongoDB for auditability and decoupling

3. **CI/CD Automation**
   - Scheduled GitHub Actions for data processing and updates

4. **Analytics Engine**
   - Computes reading statistics and trends

5. **Visualization Layer**
   - Interactive dashboards using Chart.js

6. **AI Insight Engine**
   - Generates weekly narrative analysis (velocity, backlog, chronology)

7. **Historical Tracking System**
   - Stores snapshots for long-term trend comparison

8. **Zero-Infrastructure Deployment**
   - Fully hosted on GitHub Pages (no backend servers)

9. **Observability Integration**
   - Sends events to external observability system for monitoring

10. **Cost-Optimized Architecture**
   - Uses only free-tier services (GitHub, MongoDB, Google API)

---

## 🧠 Problems I Solved

- Manual tracking → automated data pipeline
- No historical visibility → added snapshot tracking
- Raw data only → added analytics + visualizations
- Hard to interpret trends → added AI-generated insights
- Infrastructure overhead → built zero-server architecture
- Tight coupling → used event sourcing for flexibility

---

## 🛠️ Tech Stack

**Languages**
- Go, Python

**Data & APIs**
- MongoDB (event storage)
- Google Sheets API

**Frontend**
- Chart.js

**DevOps**
- GitHub Actions (CI/CD)
- Docker

**AI**
- Google Gemini (insight generation)

---

## 🏗️ System Architecture (High-Level)

Flow:

- Data source (Google Sheets)
- CI/CD pipeline (GitHub Actions)
- Event storage (MongoDB)
- Analytics processing
- Dashboard (GitHub Pages)
- Optional observability integration

---

## 🔎 Example: Data Flow

### Step 1 — Data Collection
- Reading data stored in Google Sheets

### Step 2 — Pipeline Execution
- GitHub Actions runs on schedule

### Step 3 — Processing
- Data transformed into metrics and events

### Step 4 — Output
- Dashboard updated with new analytics
- AI generates weekly insights

---

## ⚠️ Challenges

I read blogs from many different sites, and checking each one manually for new content was time-consuming.

To solve this, I built a data pipeline that centralizes content extraction into one place using automated scraping and scheduled workflows.

This allows me to track all reading data in a single system instead of visiting multiple sources manually.
 
---

## 🚀 Project Evolution

This project evolved through multiple stages:

- Local scripts → automated pipeline  
- Manual tracking → structured data ingestion  
- Static data → interactive dashboards  
- Raw metrics → AI-generated insights  
- Standalone system → integrated observability  

👉 [Read Full Evolution](docs/README.md)

---

## 📌 Summary

This project demonstrates how to build a data analytics system using:

- CI/CD-driven automation (GitHub Actions)
- event sourcing architecture (MongoDB)
- interactive visualization (Chart.js)
- AI-powered insights
- zero-infrastructure deployment

It reflects how modern data pipelines can be built with minimal cost and high automation.
