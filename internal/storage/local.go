package storage

import (
	"context"
	"errors"

	"github.com/gustavoz65/MoniMaster/internal/shared"
)

var errDisabled = errors.New("storage relacional nao esta habilitado")

type NullStore struct{}

func NewNullStore() *NullStore {
	return &NullStore{}
}

func (n *NullStore) Driver() string { return "disabled" }
func (n *NullStore) Enabled() bool  { return false }
func (n *NullStore) Close() error   { return nil }
func (n *NullStore) Ping(context.Context) error {
	return errDisabled
}
func (n *NullStore) EnsureSchema(context.Context) error {
	return nil
}
func (n *NullStore) CreateUser(context.Context, UserRecord) error {
	return errDisabled
}
func (n *NullStore) GetUserByUsername(context.Context, string) (UserRecord, error) {
	return UserRecord{}, errDisabled
}
func (n *NullStore) SaveNotificationEmail(context.Context, string, string) error {
	return errDisabled
}
func (n *NullStore) GetNotificationEmail(context.Context, string) (string, error) {
	return "", errDisabled
}
func (n *NullStore) AddHistory(context.Context, shared.HistoryRecord) error {
	return nil
}
func (n *NullStore) ListHistory(context.Context, int) ([]shared.HistoryRecord, error) {
	return nil, errDisabled
}
