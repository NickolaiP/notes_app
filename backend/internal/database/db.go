package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"noteApi/internal/config"

	_ "github.com/lib/pq"
)

// Database определяет интерфейс для взаимодействия с базой данных
type Database interface {
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Close() error
}

// PostgresDB реализует интерфейс Database
type PostgresDB struct {
	*sql.DB
}

// Query реализует метод интерфейса Database
func (db *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

// QueryRow реализует метод интерфейса Database
func (db *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// Exec реализует метод интерфейса Database
func (db *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// Close реализует метод интерфейса Database
func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

func NewPostgresDB(cfg config.DatabaseConfig) (Database, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверка подключения к базе данных
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &PostgresDB{DB: db}, nil
}
