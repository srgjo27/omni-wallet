package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UserServiceClient is the HTTP adapter that implements ports.UserServiceClient.
// It calls the User Service over internal HTTP for PIN verification and user existence checks.
type UserServiceClient struct {
	baseURL    string
	httpClient *http.Client
	jwtToken   string // static internal service-to-service token
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

// VerifyPIN calls the User Service to validate the transaction PIN.
// Returns nil if the PIN is correct, an error otherwise.
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
