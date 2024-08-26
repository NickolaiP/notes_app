package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"noteApi/internal/config"

	_ "github.com/lib/pq"
)

// Database определяет интерфейс для взаимодействия с базой данных.
// Все методы интерфейса принимают контекст для управления временем выполнения
// и отмены операций.
type Database interface {
	// Query выполняет запрос к базе данных и возвращает строки результата.
	// Аргументы запроса передаются как ...interface{}.
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow выполняет запрос к базе данных и возвращает одну строку результата.
	// Аргументы запроса передаются как ...interface{}.
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Exec выполняет запрос к базе данных, который не возвращает строки результата,
	// например, команды INSERT, UPDATE, DELETE.
	// Аргументы запроса передаются как ...interface{}.
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Close закрывает соединение с базой данных.
	Close() error
}

// PostgresDB реализует интерфейс Database для работы с базой данных PostgreSQL.
// Внутри него используется встроенное соединение базы данных *sql.DB.
type PostgresDB struct {
	*sql.DB
}

// Query выполняет запрос к базе данных с использованием контекста и возвращает строки результата.
// Этот метод реализует интерфейс Database.
func (db *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

// QueryRow выполняет запрос к базе данных с использованием контекста и возвращает одну строку результата.
// Этот метод реализует интерфейс Database.
func (db *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// Exec выполняет запрос к базе данных, который не возвращает строки результата, с использованием контекста.
// Этот метод реализует интерфейс Database.
func (db *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// Close закрывает соединение с базой данных.
// Этот метод реализует интерфейс Database.
func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

// NewPostgresDB создает и возвращает новый экземпляр PostgresDB, используя настройки из конфигурации.
// Выполняется проверка подключения к базе данных для обеспечения его корректной работы.
// При успешной проверке возвращается объект PostgresDB и nil, иначе возвращается ошибка.
func NewPostgresDB(cfg config.DatabaseConfig) (Database, error) {
	// Формирование строки подключения к базе данных PostgreSQL.
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	// Открытие соединения с базой данных.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверка подключения к базе данных с использованием контекста.
	// Таймаут установлен на 5 секунд для проверки доступности базы данных.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	// Возвращаем объект PostgresDB, который реализует интерфейс Database.
	return &PostgresDB{DB: db}, nil
}
