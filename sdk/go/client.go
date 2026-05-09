// Package owa provides a small service-side client for Open Wallet Auth.
//
// Package owa 提供认证中台服务端接入客户端，封装登录、注册和 profile 校验。
package owa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrInvalidCredentials indicates that the identity center rejected credentials or token.
	// ErrInvalidCredentials 表示认证中台拒绝了账号密码或 access token。
	ErrInvalidCredentials = errors.New("open_wallet_auth_invalid_credentials")
	// ErrEmailExists indicates that the email already exists in the identity center.
	// ErrEmailExists 表示邮箱已在认证中台存在。
	ErrEmailExists = errors.New("open_wallet_auth_email_exists")
)

// Config contains the identity-center address and current application client id.
// Config 保存认证中台地址和当前业务系统 client_id。
type Config struct {
	BaseURL    string
	ClientID   string
	HTTPClient *http.Client
}

// Client wraps the Open Wallet Auth HTTP API used by business services.
// Client 封装业务服务接入认证中台所需的 HTTP API。
type Client struct {
	baseURL    string
	clientID   string
	httpClient *http.Client
}

// User is the normalized identity user returned by the auth center.
// User 是认证中台返回的标准身份用户。
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"`
}

type tokenEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		User User `json:"user"`
	} `json:"data"`
}

type profileEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    User   `json:"data"`
}

// NewClient creates a service-side Open Wallet Auth client.
// NewClient 创建服务端认证中台客户端。
func NewClient(cfg Config) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	clientID := strings.TrimSpace(cfg.ClientID)
	if clientID == "" {
		clientID = "default"
	}
	return &Client{baseURL: baseURL, clientID: clientID, httpClient: httpClient}
}

// Login signs in through the identity center and returns the identity user.
// Login 通过认证中台登录，并返回中台身份用户。
func (c *Client) Login(ctx context.Context, email string, password string) (*User, error) {
	return c.authRequest(ctx, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
}

// Register creates an identity-center user and returns the created identity.
// Register 在认证中台创建用户，并返回创建后的身份用户。
func (c *Client) Register(ctx context.Context, username string, email string, password string) (*User, error) {
	return c.authRequest(ctx, "/api/v1/auth/register", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	})
}

// Profile validates an access token and returns the normalized identity profile.
// Profile 校验认证中台 access token，并返回标准身份资料。
func (c *Client) Profile(ctx context.Context, accessToken string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/profile", nil)
	if err != nil {
		return nil, fmt.Errorf("create profile request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("X-Client-ID", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call profile endpoint: %w", err)
	}
	defer resp.Body.Close()

	var envelope profileEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode profile response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
		return nil, ErrInvalidCredentials
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || envelope.Code != "OK" || envelope.Data.ID == "" {
		return nil, fmt.Errorf("profile response is invalid: %s", envelope.Message)
	}
	return &envelope.Data, nil
}

func (c *Client) authRequest(ctx context.Context, path string, payload map[string]string) (*User, error) {
	payload["client_id"] = c.clientID
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode auth request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call auth endpoint: %w", err)
	}
	defer resp.Body.Close()

	var envelope tokenEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode auth response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
		return nil, ErrInvalidCredentials
	}
	if resp.StatusCode == http.StatusConflict || envelope.Code == "AUTH_EMAIL_ALREADY_EXISTS" {
		return nil, ErrEmailExists
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || envelope.Code != "OK" || envelope.Data.User.ID == "" {
		return nil, fmt.Errorf("auth response is invalid: %s", envelope.Message)
	}
	return &envelope.Data.User, nil
}
