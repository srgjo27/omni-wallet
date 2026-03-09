package walletclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type createWalletRequest struct {
	UserID string `json:"user_id"`
}

type apiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

func (c *Client) ProvisionWallet(ctx context.Context, userID string) error {
	body, err := json.Marshal(createWalletRequest{UserID: userID})
	if err != nil {
		return fmt.Errorf("walletclient: marshalling request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/wallets",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("walletclient: building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("walletclient: calling wallet-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}

	var apiResp apiResponse
	if decErr := json.NewDecoder(resp.Body).Decode(&apiResp); decErr != nil {
		return fmt.Errorf("walletclient: unexpected status %d", resp.StatusCode)
	}
	return fmt.Errorf("walletclient: wallet-service returned %d: %s", resp.StatusCode, apiResp.Message)
}
