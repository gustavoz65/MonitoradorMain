package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gustavoz65/MoniMaster/internal/config"
)

type ResendProvider struct{}

func (p *ResendProvider) Name() string { return "resend" }

func (p *ResendProvider) Send(cfg config.AppConfig, to, subject, body string) error {
	apiKey := strings.TrimSpace(cfg.Notify.APIKey)
	from := strings.TrimSpace(cfg.Notify.From)
	if apiKey == "" {
		return fmt.Errorf("resend: api key nao configurada; use config notify provider set resend --api-key xxx")
	}
	if from == "" {
		return fmt.Errorf("resend: remetente nao configurado; use config notify provider set resend --from noreply@dominio.com")
	}
	payload, _ := json.Marshal(map[string]any{
		"from": from, "to": []string{to}, "subject": subject, "text": body,
	})
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend retornou status %d", resp.StatusCode)
	}
	return nil
}
