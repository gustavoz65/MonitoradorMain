package notify

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/config"
	"github.com/gustavoz65/MoniMaster/internal/shared"
	"github.com/gustavoz65/MoniMaster/internal/storage"
)

type alertJob struct {
	cfg     config.AppConfig
	to      string
	subject string
	body    string
}

type Service struct {
	store    storage.Store
	provider Provider
	alerts   chan alertJob
}

func NewService(store storage.Store) *Service {
	s := &Service{
		store:    store,
		provider: &SMTPProvider{},
		alerts:   make(chan alertJob, 64),
	}
	go s.dispatchLoop()
	return s
}

func (s *Service) SetProvider(p Provider) { s.provider = p }

func (s *Service) dispatchLoop() {
	for job := range s.alerts {
		_ = s.provider.Send(job.cfg, job.to, job.subject, job.body)
	}
}

func (s *Service) ResolveTarget(cfg config.AppConfig, identity *shared.Identity) string {
	if identity != nil && s.store.Enabled() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		email, err := s.store.GetNotificationEmail(ctx, identity.ID)
		if err == nil && strings.TrimSpace(email) != "" {
			return email
		}
	}
	return cfg.Notification.DefaultEmail
}

func (s *Service) SetTarget(cfg *config.AppConfig, identity *shared.Identity, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email vazio")
	}
	if identity != nil && s.store.Enabled() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.store.SaveNotificationEmail(ctx, identity.ID, email); err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	cfg.Notification.DefaultEmail = email
	return nil
}

func (s *Service) Send(cfg config.AppConfig, to, subject, body string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("destinatario nao configurado")
	}
	select {
	case s.alerts <- alertJob{cfg, to, subject, body}:
		return nil
	default:
		return fmt.Errorf("fila de alertas cheia")
	}
}

func (s *Service) SendSync(cfg config.AppConfig, to, subject, body string) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("destinatario nao configurado")
	}
	return s.provider.Send(cfg, to, subject, body)
}
