package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/truckguard/core/src/api/dtos"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type AuthClient struct {
	BaseURL      string
	HTTPClient   *http.Client
	AuthHeader   string
	ApiKeyHeader string
}

func NewAuthClient(authHeader, apiKeyHeader string) *AuthClient {
	baseURL := os.Getenv("AUTH_SERVICE_URL")
	if baseURL == "" {
		return nil
	}
	return &AuthClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
		AuthHeader:   authHeader,
		ApiKeyHeader: apiKeyHeader,
	}
}

type CreateKeyRequest struct {
	Name          string   `json:"name"`
	PermissionIDs []string `json:"permission_ids"`
}

type CreateKeyResponse struct {
	ID     interface{} `json:"id"`
	APIKey string      `json:"api_key"`
}

type RegisterUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func (c *AuthClient) CreateApiKey(ctx context.Context, name string, permissions []string) (*CreateKeyResponse, error) {
	url := fmt.Sprintf("%s/admin/keys", c.BaseURL)

	reqBody := CreateKeyRequest{
		Name:          name,
		PermissionIDs: permissions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	}
	if c.ApiKeyHeader != "" {
		req.Header.Set("X-Api-Key", c.ApiKeyHeader)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	var result CreateKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *AuthClient) DeleteApiKey(ctx context.Context, keyID string) error {
	url := fmt.Sprintf("%s/admin/keys/%s", c.BaseURL, keyID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	}
	if c.ApiKeyHeader != "" {
		req.Header.Set("X-Api-Key", c.ApiKeyHeader)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *AuthClient) RegisterUser(ctx context.Context, username, password, role string) (*RegisterUserResponse, error) {
	url := fmt.Sprintf("%s/register", c.BaseURL)

	reqBody := RegisterUserRequest{
		Username: username,
		Password: password,
		Role:     role,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	}
	if c.ApiKeyHeader != "" {
		req.Header.Set("X-Api-Key", c.ApiKeyHeader)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth service registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("auth service registration returned status: %d", resp.StatusCode)
	}

	var result RegisterUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	return &result, nil
}

func (c *AuthClient) DeleteUser(ctx context.Context, userID uint) error {
	url := fmt.Sprintf("%s/admin/users/%d", c.BaseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete user request: %w", err)
	}

	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	}
	if c.ApiKeyHeader != "" {
		req.Header.Set("X-Api-Key", c.ApiKeyHeader)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth service delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth service delete user returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *AuthClient) GetUser(ctx context.Context, userID string) (*dtos.AuthUser, error) {
	url := fmt.Sprintf("%s/admin/users/%s", c.BaseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get user request: %w", err)
	}

	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth service get user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service get user returned status: %d", resp.StatusCode)
	}

	var user dtos.AuthUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode get user response: %w", err)
	}

	return &user, nil
}
