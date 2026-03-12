package main

// This service accepts HTTP requests and drops them into Redis.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chryote/go-job-queue/internal/queue"
	"github.com/chryote/go-job-queue/internal/store"
	"github.com/redis/go-redis/v9"
)

// App holds dependencies
type App struct {
	DB    *sql.DB
	Redis *redis.Client
}

func main() {
	// 1. initialize DB & Redis
	db, _ := store.NewPostgres("postgres://user:password@postgres:5432/jobs_db?sslmode=disable")
	rdb := queue.NewRedis("redis:6379")

	app := &App{
		DB:    db,
		Redis: rdb,
	}

	// 2. setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/enqueue", app.HandleEnqueue)
	mux.HandleFunc("/status", app.HandleStatus)

	fmt.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

	janitor := store.NewJanitor(db, rdb)

	// Run Janitor every 1 minute in a goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			// Recover jobs older than 5 minutes
			err := janitor.RecoverOrphanedJobs(context.Background(), 5*time.Minute)
			if err != nil {
				log.Printf("Janitor error: %v", err)
			}
		}
	}()

}

func (a *App) HandleEnqueue(w http.ResponseWriter, r *http.Request) {
	// 1. decode the request
	var job queue.JobPayload
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request", 400)
		return
	}

	// 2. save to Postgres with the Type

	// Using "ON CONFLICT DO NOTHING" ensures it don't create two records for the same ID
	query := `INSERT INTO jobs (id, job_type, status, payload) 
          VALUES ($1, $2, 'pending', $3) 
          ON CONFLICT (id) DO NOTHING`

	res, err := a.DB.Exec(query, job.ID, job.Type, job.Payload)
	rows, _ := res.RowsAffected()

	if rows == 0 {
		// If 0 rows were affected, the ID already exists!
		w.Write([]byte("Job ID already exists, skipping..."))
		return
	}

	if err != nil {
		log.Printf("DB Error: %v", err)
		http.Error(w, "Persistence failed", 500)
		return
	}

	// 3. marshal and Push to Redis
	msg, _ := json.Marshal(job)
	queue.Push(r.Context(), a.Redis, msg)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Job accepted"))
}

func (a *App) HandleStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	var status string
	err := a.DB.QueryRow("SELECT status FROM jobs WHERE id = $1", id).Scan(&status)
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	fmt.Fprintf(w, "Job %s is %s", id, status)
}
