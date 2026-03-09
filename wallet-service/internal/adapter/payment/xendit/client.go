package xendit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/omni-wallet/wallet-service/internal/core/domain"
)

const apiBaseURL = "https://api.xendit.co"

type Client struct {
	secretKey    string
	webhookToken string
	callbackURL  string
	httpClient   *http.Client
}

func NewClient(secretKey, webhookToken, callbackURL string) *Client {
	return &Client{
		secretKey:    secretKey,
		webhookToken: webhookToken,
		callbackURL:  callbackURL,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) VerifyWebhookToken(token string) bool {
	return c.webhookToken != "" && token == c.webhookToken
}

type createVABody struct {
	ExternalID  string `json:"external_id"`
	BankCode    string `json:"bank_code"`
	Name        string `json:"name"`
	Currency    string `json:"currency"`
	IsClosed    bool   `json:"is_closed"`
	IsSingleUse bool   `json:"is_single_use"`
	CallbackURL string `json:"callback_url,omitempty"`
}

type xenditVAResp struct {
	ID            string `json:"id"`
	ExternalID    string `json:"external_id"`
	BankCode      string `json:"bank_code"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
	MerchantCode  string `json:"merchant_code"`
	Currency      string `json:"currency"`
	IsClosed      bool   `json:"is_closed"`
	Status        string `json:"status"`
}

type xenditErrResp struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

type APIError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("xendit error %d [%s]: %s", e.HTTPStatus, e.Code, e.Message)
}

func (c *Client) CreateFixedVA(externalID, name, bankCode string) (*domain.VirtualAccount, error) {
	payload := createVABody{
		ExternalID:  externalID,
		BankCode:    bankCode,
		Name:        name,
		Currency:    "IDR",
		IsClosed:    false,
		IsSingleUse: false,
		CallbackURL: c.callbackURL,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling VA payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiBaseURL+"/callback_virtual_accounts", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("building Xendit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.secretKey, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Xendit API: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var e xenditErrResp
		_ = json.Unmarshal(respBody, &e)
		return nil, &APIError{
			HTTPStatus: resp.StatusCode,
			Code:       e.ErrorCode,
			Message:    e.Message,
		}
	}

	var va xenditVAResp
	if err := json.Unmarshal(respBody, &va); err != nil {
		return nil, fmt.Errorf("parsing Xendit VA response: %w", err)
	}

	return mapVA(va), nil
}

func mapVA(va xenditVAResp) *domain.VirtualAccount {
	return &domain.VirtualAccount{
		ID:            va.ID,
		ExternalID:    va.ExternalID,
		BankCode:      va.BankCode,
		Name:          va.Name,
		AccountNumber: va.AccountNumber,
		MerchantCode:  va.MerchantCode,
		Currency:      va.Currency,
		IsClosed:      va.IsClosed,
		Status:        va.Status,
	}
}
