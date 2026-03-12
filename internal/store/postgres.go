package store

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgres(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Ping the database until it responds (up to 10 times)
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Println("Waiting for database...")
		time.Sleep(2 * time.Second)
	}

	// Table initialization for kickstart
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		job_type TEXT NOT NULL,
		status TEXT NOT NULL,
		payload TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)

	return db, err

}
