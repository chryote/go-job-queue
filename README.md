# Go-Redis-Postgres Distributed Job Queue

A robust, production-grade distributed job processing system written in Go. This project demonstrates how to handle asynchronous tasks with high reliability, observability, and self-healing capabilities.

## 🏗 Architecture Overview

- **API Server**: A RESTful entry point that validates jobs and persists them to PostgreSQL before enqueuing them in Redis.
- **Worker**: A scalable consumer that pulls jobs from Redis, processes them with built-in retry logic, and updates the final status in PostgreSQL.
- **Janitor Service**: A background process using `SKIP LOCKED` logic to recover "orphaned" jobs (jobs stuck in processing due to worker crashes).
- **Observability Stack**: Real-time metrics via Prometheus and visual dashboards via Grafana.

## 🚀 Key Features

- **At-Least-Once Delivery**: Jobs are persisted in Postgres before processing.
- **Self-Healing**: Automatic recovery of stalled tasks.
- **Idempotency**: Prevents duplicate job processing using unique task IDs.
- **Observability**: Tracks job throughput, success/failure rates, and processing latency.
- **Scalability**: Designed to run with multiple worker instances sharing the same Redis/DB backends.

## 🛠 Tech Stack

- **Language**: Go (Golang)
- **Database**: PostgreSQL (State Persistence)
- **Message Broker**: Redis (Task Queue)
- **Monitoring**: Prometheus & Grafana
- **Containerization**: Docker & Docker Compose

## 🚦 Getting Started

### Prerequisites
- Docker and Docker Compose
- Go 1.22+ (for local development)

### 1. Start the Stack
```bash
docker-compose up --build
```
### 2. Enqueue a Sample Job
curl -X POST http://localhost:8080/enqueue \
  -H "Content-Type: application/json" \
  -d '{
    "id": "job_001",
    "type": "EMAIL",
    "payload": "hello@example.com"
  }'

### 3. Access Monitoring Dashboards
- Grafana: http://localhost:3000 (Login: admin / admin)
- Prometheus: http://localhost:9090
- Metrics Endpoint: http://localhost:2112/metrics

### 4. Monitoring Queries (PromQL)
- Throughput: sum(rate(job_queue_processed_total[1m])) by (type)
- Avg Latency: rate(job_duration_seconds_sum[5m]) / rate(job_duration_seconds_count[5m])
- Error Rate: job_queue_processed_total{status="failed"}

### 5. Project Structure
.
├── cmd/
│   ├── api/            # API Server (Enqueuer)
│   └── worker/         # Task Processor
├── internal/
│   ├── metrics/        # Prometheus instrumentation
│   ├── queue/          # Redis client & models
│   └── store/          # Postgres & Janitor logic
├── prometheus.yml      # Scrape configuration
├── docker-compose.yml  # Infrastructure orchestration
└── README.md