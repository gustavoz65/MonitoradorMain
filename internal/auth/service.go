package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/shared"
	"github.com/gustavoz65/MoniMaster/internal/storage"
)

type Service struct {
	store storage.Store
}

func NewService(store storage.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Register(username, email, password string) (shared.Identity, error) {
	if !s.store.Enabled() {
		return shared.Identity{}, fmt.Errorf("cadastro exige banco configurado")
	}
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if err := ValidateCredentials(username, email, password); err != nil {
		return shared.Identity{}, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return shared.Identity{}, err
	}
	identity := shared.Identity{
		ID:       shared.NewID("user"),
		Username: username,
		Email:    email,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.store.CreateUser(ctx, storage.UserRecord{
		ID:           identity.ID,
		Username:     identity.Username,
		Email:        identity.Email,
		PasswordHash: hash,
	})
	return identity, err
}

func (s *Service) Login(username, password string) (shared.Identity, error) {
	if !s.store.Enabled() {
		return shared.Identity{}, fmt.Errorf("login exige banco configurado")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	record, err := s.store.GetUserByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if err == sql.ErrNoRows {
			return shared.Identity{}, fmt.Errorf("usuario nao encontrado")
		}
		return shared.Identity{}, err
	}
	if err := ComparePassword(record.PasswordHash, password); err != nil {
		return shared.Identity{}, fmt.Errorf("senha incorreta")
	}
	return shared.Identity{ID: record.ID, Username: record.Username, Email: record.Email}, nil
}
