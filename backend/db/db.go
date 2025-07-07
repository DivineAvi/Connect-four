package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	Conn *pgx.Conn
}

var Conn *pgx.Conn

func Connect(url string) {
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
		return
	}
	Conn = conn
}

func (db *DB) Close() error {
	if db.Conn != nil {
		return db.Conn.Close(context.Background())
	}
	return nil
}
