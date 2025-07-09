package db

import (
	"backend/config"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

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

func NewDB() (*DB, error) {
	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		return nil, err
	}

	connString := dbConfig.DatabaseURL
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	db := &DB{Pool: pool}

	if err := db.initTables(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

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

func (db *DB) GetLeaderboard(limit int) ([]Player, error) {
	if limit <= 0 {
		limit = 10
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
		return db.GetPlayerByUsername(username)
	}

	return &p, nil
}

func (db *DB) UpdateGameResult(winner, loser string) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	winnerQuery := `
	UPDATE players
	SET wins = wins + 1, rating = rating + 25, updated_at = CURRENT_TIMESTAMP
	WHERE username = $1
	`
	_, err = tx.Exec(context.Background(), winnerQuery, winner)
	if err != nil {
		return fmt.Errorf("failed to update winner: %v", err)
	}

	loserQuery := `
	UPDATE players
	SET losses = losses + 1, rating = GREATEST(rating - 15, 0), updated_at = CURRENT_TIMESTAMP
	WHERE username = $1
	`
	_, err = tx.Exec(context.Background(), loserQuery, loser)
	if err != nil {
		return fmt.Errorf("failed to update loser: %v", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (db *DB) UpdateDraw(player1, player2 string) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	query := `
	UPDATE players
	SET draws = draws + 1, rating = rating + 5, updated_at = CURRENT_TIMESTAMP
	WHERE username = $1
	`
	_, err = tx.Exec(context.Background(), query, player1)
	if err != nil {
		return fmt.Errorf("failed to update player1: %v", err)
	}

	_, err = tx.Exec(context.Background(), query, player2)
	if err != nil {
		return fmt.Errorf("failed to update player2: %v", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
