package http

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// Opt определяет функцию конфигурации клиента Resty.
type Opt func(*resty.Client) error

// New создаёт новый Resty клиент с базовым URL и применяет указанные опции.
// baseURL — адрес сервера (например, "http://localhost:8080").
// opts — список опций конфигурации клиента.
func New(baseURL string, opts ...Opt) (*resty.Client, error) {
	client := resty.New().SetBaseURL(baseURL)

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// RetryPolicy описывает параметры повторной попытки HTTP-запроса.
type RetryPolicy struct {
	Count   int           // Количество попыток повторного запроса
	Wait    time.Duration // Время ожидания между попытками
	MaxWait time.Duration // Максимальное время ожидания между попытками
}

// WithRetryPolicy возвращает опцию для конфигурации клиента с политикой повторов.
// Можно передать несколько RetryPolicy, но будет использован первый непустой.
func WithRetryPolicy(policies ...RetryPolicy) Opt {
	return func(c *resty.Client) error {
		for _, policy := range policies {
			if policy.Count > 0 || policy.Wait > 0 || policy.MaxWait > 0 {
				if policy.Count > 0 {
					c.SetRetryCount(policy.Count)
				}
				if policy.Wait > 0 {
					c.SetRetryWaitTime(policy.Wait)
				}
				if policy.MaxWait > 0 {
					c.SetRetryMaxWaitTime(policy.MaxWait)
				}
				break
			}
		}
		return nil
	}
}
