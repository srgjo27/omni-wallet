package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type UserServiceClient struct {
	baseURL    string
	httpClient *http.Client
	jwtToken   string
}

func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type verifyPINRequest struct {
	UserID string `json:"user_id"`
	PIN    string `json:"pin"`
}

type verifyPINResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *UserServiceClient) VerifyPIN(userID string, pin string) error {
	payload := verifyPINRequest{UserID: userID, PIN: pin}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling verify-pin request: %w", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/v1/internal/users/verify-pin", c.baseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("calling user-service verify-pin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid transaction PIN")
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result verifyPINResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parsing verify-pin response: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("invalid transaction PIN: %s", result.Message)
	}

	return nil
}

type existsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Exists bool `json:"exists"`
	} `json:"data"`
}

type lookupByEmailResponse struct {
	Success bool `json:"success"`
	Data    struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
	} `json:"data"`
}

func (c *UserServiceClient) FindUserIDByEmail(email string) (string, error) {
	resp, err := c.httpClient.Get(
		fmt.Sprintf("%s/api/v1/internal/users/lookup?email=%s", c.baseURL, url.QueryEscape(email)),
	)
	if err != nil {
		return "", fmt.Errorf("calling user-service lookup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("target user not found")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from user-service lookup", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result lookupByEmailResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parsing lookup response: %w", err)
	}
	return result.Data.UserID, nil
}

func (c *UserServiceClient) ExistsByID(userID string) (bool, error) {
	resp, err := c.httpClient.Get(
		fmt.Sprintf("%s/api/v1/internal/users/%s/exists", c.baseURL, userID),
	)
	if err != nil {
		return false, fmt.Errorf("calling user-service exists endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status %d from user-service", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result existsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, fmt.Errorf("parsing exists response: %w", err)
	}

	return result.Data.Exists, nil
}
