package db

import (
	"backend/config"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents the database connection
type DB struct {
	Pool *pgxpool.Pool
}

// Player represents a player in the leaderboard
type Player struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Wins      int       `json:"wins"`
	Losses    int       `json:"losses"`
	Draws     int       `json:"draws"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		return nil, err
	}

	connString := dbConfig.ConnectionString()
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	db := &DB{Pool: pool}

	// Initialize database tables
	if err := db.initTables(); err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// initTables creates the necessary tables if they don't exist
func (db *DB) initTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS players (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		wins INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		draws INTEGER DEFAULT 0,
		rating INTEGER DEFAULT 1000,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	return nil
}

// GetLeaderboard returns the top players sorted by rating
func (db *DB) GetLeaderboard(limit int) ([]Player, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
	SELECT id, username, wins, losses, draws, rating, created_at, updated_at
	FROM players
	ORDER BY rating DESC
	LIMIT $1
	`

	rows, err := db.Pool.Query(context.Background(), query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %v", err)
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(
			&p.ID, &p.Username, &p.Wins, &p.Losses, &p.Draws,
			&p.Rating, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan player row: %v", err)
		}
		players = append(players, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return players, nil
}

// GetPlayerByUsername retrieves a player by username
func (db *DB) GetPlayerByUsername(username string) (*Player, error) {
	query := `
	SELECT id, username, wins, losses, draws, rating, created_at, updated_at
	FROM players
	WHERE username = $1
	`

	var p Player
	err := db.Pool.QueryRow(context.Background(), query, username).Scan(
		&p.ID, &p.Username, &p.Wins, &p.Losses, &p.Draws,
		&p.Rating, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("player not found: %v", err)
	}

	return &p, nil
}

// CreateOrUpdatePlayer creates a new player or updates an existing one
func (db *DB) CreateOrUpdatePlayer(username string) (*Player, error) {
	query := `
	INSERT INTO players (username)
	VALUES ($1)
	ON CONFLICT (username) DO NOTHING
	RETURNING id, username, wins, losses, draws, rating, created_at, updated_at
	`

	var p Player
	err := db.Pool.QueryRow(context.Background(), query, username).Scan(
		&p.ID, &p.Username, &p.Wins, &p.Losses, &p.Draws,
		&p.Rating, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		// If the player already exists, get their data
		return db.GetPlayerByUsername(username)
	}

	return &p, nil
}

// UpdateGameResult updates player statistics after a game
func (db *DB) UpdateGameResult(winner, loser string) error {
	// Start a transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	// Update winner
	winnerQuery := `
	UPDATE players
	SET wins = wins + 1, rating = rating + 25, updated_at = CURRENT_TIMESTAMP
	WHERE username = $1
	`
	_, err = tx.Exec(context.Background(), winnerQuery, winner)
	if err != nil {
		return fmt.Errorf("failed to update winner: %v", err)
	}

	// Update loser
	loserQuery := `
	UPDATE players
	SET losses = losses + 1, rating = GREATEST(rating - 15, 0), updated_at = CURRENT_TIMESTAMP
	WHERE username = $1
	`
	_, err = tx.Exec(context.Background(), loserQuery, loser)
	if err != nil {
		return fmt.Errorf("failed to update loser: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// UpdateDraw updates player statistics after a draw
func (db *DB) UpdateDraw(player1, player2 string) error {
	// Start a transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	// Update player1
	query := `
	UPDATE players
	SET draws = draws + 1, rating = rating + 5, updated_at = CURRENT_TIMESTAMP
	WHERE username = $1
	`
	_, err = tx.Exec(context.Background(), query, player1)
	if err != nil {
		return fmt.Errorf("failed to update player1: %v", err)
	}

	// Update player2
	_, err = tx.Exec(context.Background(), query, player2)
	if err != nil {
		return fmt.Errorf("failed to update player2: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
