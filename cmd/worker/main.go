package main

// This service pulls from Redis and updates Postgres.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/chryote/go-job-queue/internal/metrics"
	"github.com/chryote/go-job-queue/internal/queue"
	"github.com/chryote/go-job-queue/internal/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type Worker struct {
	DB    *sql.DB
	Redis *redis.Client
}

func main() {
	db, _ := store.NewPostgres("postgres://user:password@postgres:5432/jobs_db?sslmode=disable")
	rdb := queue.NewRedis("redis:6379")

	// START THE METRICS SERVER IN A GOROUTINE
	go func() {
		log.Println("📊 Metrics server starting on :2112/metrics")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Printf("❌ Metrics server failed: %v", err)
		}
	}()

	worker := &Worker{DB: db, Redis: rdb}
	worker.Start()
}

func (w *Worker) Start() {
	log.Println("Worker listening for various job types...")

	for {
		res, err := queue.Pop(context.Background(), w.Redis)
		if err != nil {
			continue
		}

		startTime := time.Now()

		var job queue.JobPayload
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
			log.Printf("❌ JSON Error: %v", err)
			continue
		}

		log.Printf("🔄 Processing %s: %s", job.Type, job.ID)

		var workErr error
		switch job.Type {
		case "EMAIL":
			workErr = w.sendEmail(job.Payload)
		case "IMAGE_RESIZE":
			workErr = w.processImage(job.Payload)
		default:
			log.Printf("⚠️ Unknown type: %s", job.Type)
			workErr = fmt.Errorf("unknown task type")
		}

		metrics.JobDuration.WithLabelValues(job.Type).Observe(time.Since(startTime).Seconds())

		// Pass the results to the update function
		w.updateJobsStatus(job.ID, job, workErr)
	}
}

func (w *Worker) processImage(payload string) error {
	log.Printf("[IMAGE_RESIZE] Starting work on: %s", payload)

	// 1. Simulate "Processing Time" (2-5 seconds)
	duration := time.Duration(2+rand.Intn(3)) * time.Second
	time.Sleep(duration)

	// 2. Simulate an occasional failure (10% chance)
	if rand.Float32() < 0.1 {
		return errors.New("image format not supported or file corrupted")
	}

	log.Printf("[IMAGE_RESIZE] Success: %s", payload)
	return nil
}

func (w *Worker) sendEmail(payload string) error {
	log.Printf("[SEND_EMAIL] Sending to: %s", payload)

	// 1. Simulate Network Latency (1 second)
	time.Sleep(1 * time.Second)

	// 2. Simulate a "hard" failure for specific dummy data
	if payload == "fail@example.com" {
		return errors.New("SMTP server timeout")
	}

	log.Printf("[SEND_EMAIL] Sent successfully to: %s", payload)
	return nil
}

// Update the signature to accept jobID and the error result
func (w *Worker) updateJobsStatus(jobID string, job queue.JobPayload, workErr error) {
	finalStatus := "completed"
	if workErr != nil {
		log.Printf("❌ Job %s failed: %v", jobID, workErr)
		finalStatus = "failed"
	} else {
		log.Printf("✅ Job %s completed", jobID)
	}

	// FIX: Only increment the specific label that matches the result
	metrics.JobsCounter.WithLabelValues(job.Type, finalStatus).Inc()

	_, err := w.DB.Exec("UPDATE jobs SET status = $1 WHERE id = $2", finalStatus, jobID)
	if err != nil {
		log.Printf("❌ DB Update Error: %v", err)
	}
}
