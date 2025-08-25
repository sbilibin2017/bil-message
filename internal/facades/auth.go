package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// AuthFacade представляет Auth-сервис
type AuthFacade struct {
	client  *resty.Client
	baseURL string
}

// New создаёт новый Auth Service
func NewAuthFacade(baseURL string) *AuthFacade {
	return &AuthFacade{
		client:  resty.New(),
		baseURL: baseURL,
	}
}

// Decode проверяет и декодирует токен через auth-service
func (s *AuthFacade) Decode(ctx context.Context, token string) (*models.TokenResponse, error) {
	var res models.TokenResponse

	resp, err := s.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&res).
		Post(s.baseURL + "/api/v1/auth/token/decode")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("invalid token: %s", resp.String())
	}

	return &res, nil
}
