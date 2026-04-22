package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/sijms/go-ora/v2"
	"github.com/gustavoz65/MoniMaster/internal/config"
	"github.com/gustavoz65/MoniMaster/internal/shared"
	_ "modernc.org/sqlite"
)

type SQLStore struct {
	db     *sql.DB
	driver string
}

func New(cfg config.StorageConfig) (Store, error) {
	if !cfg.Enabled || strings.TrimSpace(cfg.Driver) == "" || strings.TrimSpace(cfg.DSN) == "" {
		return NewNullStore(), nil
	}
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	db, err := sql.Open(driver, cfg.DSN)
	if err != nil {
		return nil, err
	}
	store := &SQLStore{db: db, driver: driver}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.EnsureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLStore) Driver() string { return s.driver }
func (s *SQLStore) Enabled() bool  { return true }
func (s *SQLStore) Close() error   { return s.db.Close() }
func (s *SQLStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLStore) EnsureSchema(ctx context.Context) error {
	for _, query := range schemaQueries(s.driver) {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("schema %s: %w", s.driver, err)
		}
	}
	return nil
}

func (s *SQLStore) CreateUser(ctx context.Context, user UserRecord) error {
	query := bind(s.driver, "INSERT INTO users (id, username, email, password_hash, created_at) VALUES (?, ?, ?, ?, ?)")
	_, err := s.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.PasswordHash, time.Now().UTC())
	return err
}

func (s *SQLStore) GetUserByUsername(ctx context.Context, username string) (UserRecord, error) {
	query := bind(s.driver, "SELECT id, username, email, password_hash FROM users WHERE username = ?")
	var user UserRecord
	err := s.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	return user, err
}

func (s *SQLStore) SaveNotificationEmail(ctx context.Context, userID, email string) error {
	switch s.driver {
	case "postgres", "sqlite", "mysql":
		query := bind(s.driver, "INSERT INTO notification_targets (user_id, email, updated_at) VALUES (?, ?, ?) ON CONFLICT(user_id) DO UPDATE SET email = excluded.email, updated_at = excluded.updated_at")
		if s.driver == "mysql" {
			query = "INSERT INTO notification_targets (user_id, email, updated_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE email = VALUES(email), updated_at = VALUES(updated_at)"
		}
		_, err := s.db.ExecContext(ctx, query, userID, email, time.Now().UTC())
		return err
	case "oracle":
		query := "MERGE INTO notification_targets t USING (SELECT :1 AS user_id, :2 AS email FROM dual) src ON (t.user_id = src.user_id) WHEN MATCHED THEN UPDATE SET t.email = src.email, t.updated_at = :3 WHEN NOT MATCHED THEN INSERT (user_id, email, updated_at) VALUES (:1, :2, :3)"
		_, err := s.db.ExecContext(ctx, query, userID, email, time.Now().UTC())
		return err
	default:
		return fmt.Errorf("driver nao suportado: %s", s.driver)
	}
}

func (s *SQLStore) GetNotificationEmail(ctx context.Context, userID string) (string, error) {
	query := bind(s.driver, "SELECT email FROM notification_targets WHERE user_id = ?")
	var email string
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&email)
	return email, err
}

func (s *SQLStore) AddHistory(ctx context.Context, record shared.HistoryRecord) error {
	query := bind(s.driver, "INSERT INTO history (id, actor, mode, action, target, success, details, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	successValue := 0
	if record.Success {
		successValue = 1
	}
	_, err := s.db.ExecContext(ctx, query, record.ID, record.Actor, record.Mode, record.Action, record.Target, successValue, record.Details, record.CreatedAt.UTC())
	return err
}

func (s *SQLStore) ListHistory(ctx context.Context, limit int) ([]shared.HistoryRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	query := bind(s.driver, "SELECT id, actor, mode, action, target, success, details, created_at FROM history ORDER BY created_at DESC LIMIT ?")
	if s.driver == "oracle" {
		query = "SELECT id, actor, mode, action, target, success, details, created_at FROM history ORDER BY created_at DESC FETCH FIRST :1 ROWS ONLY"
	}
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []shared.HistoryRecord
	for rows.Next() {
		var (
			record  shared.HistoryRecord
			success int
		)
		if err := rows.Scan(&record.ID, &record.Actor, &record.Mode, &record.Action, &record.Target, &success, &record.Details, &record.CreatedAt); err != nil {
			return nil, err
		}
		record.Success = success == 1
		records = append(records, record)
	}
	return records, rows.Err()
}

func bind(driver, query string) string {
	switch driver {
	case "postgres":
		index := 1
		for strings.Contains(query, "?") {
			query = strings.Replace(query, "?", fmt.Sprintf("$%d", index), 1)
			index++
		}
	case "oracle":
		index := 1
		for strings.Contains(query, "?") {
			query = strings.Replace(query, "?", fmt.Sprintf(":%d", index), 1)
			index++
		}
	}
	return query
}

func schemaQueries(driver string) []string {
	textType := "TEXT"
	switch driver {
	case "oracle":
		textType = "CLOB"
	}
	return []string{
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS users (id VARCHAR(64) PRIMARY KEY, username VARCHAR(191) UNIQUE NOT NULL, email VARCHAR(191) UNIQUE NOT NULL, password_hash VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL)"),
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS notification_targets (user_id VARCHAR(64) PRIMARY KEY, email VARCHAR(191) NOT NULL, updated_at TIMESTAMP NOT NULL)"),
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS history (id VARCHAR(64) PRIMARY KEY, actor VARCHAR(191) NOT NULL, mode VARCHAR(32) NOT NULL, action VARCHAR(64) NOT NULL, target %s, success INTEGER NOT NULL, details %s, created_at TIMESTAMP NOT NULL)", textType, textType),
	}
}
