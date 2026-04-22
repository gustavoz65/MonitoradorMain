package storage

import (
	"context"

	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type UserRecord struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
}

type Store interface {
	Driver() string
	Enabled() bool
	Close() error
	Ping(context.Context) error
	EnsureSchema(context.Context) error
	CreateUser(context.Context, UserRecord) error
	GetUserByUsername(context.Context, string) (UserRecord, error)
	SaveNotificationEmail(context.Context, string, string) error
	GetNotificationEmail(context.Context, string) (string, error)
	AddHistory(context.Context, shared.HistoryRecord) error
	ListHistory(context.Context, int) ([]shared.HistoryRecord, error)
}
