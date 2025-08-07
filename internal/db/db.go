package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func MustConnect() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("DB_HOST", "postgres"),
		getenv("DB_USER", "postgres"),
		getenv("DB_PASSWORD", "postgres"),
		getenv("DB_NAME", "postgres"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	return db
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
