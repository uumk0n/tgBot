package main

import (
	"database/sql"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/google/uuid"
)

func writeToClickHouse(logMessage string) error {
	dsn := "tcp://localhost:9000"
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create a table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS logs (
            id UUID DEFAULT generateUUIDv4(),
            timestamp DateTime DEFAULT now(),
            message String
        ) ENGINE = MergeTree()
        ORDER BY id
    `)
	if err != nil {
		return err
	}

	// Generate unique ID and current time
	uniqueID := uuid.New().String()
	currentTime := time.Now()

	// Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback if not committed

	// Prepare the insert statement
	stmt, err := tx.Prepare("INSERT INTO logs (id, message, timestamp) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert data into the transaction
	_, err = stmt.Exec(uniqueID, logMessage, currentTime)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
