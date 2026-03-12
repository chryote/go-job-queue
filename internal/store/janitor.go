package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/chryote/go-job-queue/internal/queue"
	"github.com/redis/go-redis/v9"
)

type Janitor struct {
	DB    *sql.DB
	Redis *redis.Client
}

func NewJanitor(db *sql.DB, rdb *redis.Client) *Janitor {
	return &Janitor{DB: db, Redis: rdb}
}

// RecoverOrphanedJobs finds jobs stuck in 'pending' or 'processing' for too long
func (j *Janitor) RecoverOrphanedJobs(ctx context.Context, timeout time.Duration) error {
	// 1. Start a transaction
	tx, err := j.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2. Find jobs that haven't been updated recently.
	// We use SKIP LOCKED so other janitor instances won't collide.
	query := `
		SELECT id, job_type, payload 
		FROM jobs 
		WHERE (status = 'pending' OR status = 'processing' OR status = 'retrying')
		AND created_at < $1
		FOR UPDATE SKIP LOCKED`

	// Define "too old" (e.g., created more than 'timeout' ago)
	threshold := time.Now().Add(-timeout)

	rows, err := tx.QueryContext(ctx, query, threshold)
	if err != nil {
		return err
	}
	defer rows.Close()

	var recoveredCount int
	for rows.Next() {
		var job queue.JobPayload
		if err := rows.Scan(&job.ID, &job.Type, &job.Payload); err != nil {
			continue
		}

		// 3. Re-queue back into Redis
		payload, _ := json.Marshal(job)
		if err := queue.Push(ctx, j.Redis, payload); err != nil {
			log.Printf("Failed to re-queue job %s: %v", job.ID, err)
			continue
		}

		// 4. Update status back to 'pending' to reset the clock
		_, err = tx.ExecContext(ctx, "UPDATE jobs SET status = 'pending', created_at = NOW() WHERE id = $1", job.ID)
		if err != nil {
			return err
		}
		recoveredCount++
	}

	if recoveredCount > 0 {
		log.Printf("🧹 Janitor: Recovered %d orphaned jobs", recoveredCount)
	}

	return tx.Commit()
}
