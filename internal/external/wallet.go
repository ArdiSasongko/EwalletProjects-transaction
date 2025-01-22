package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/env"
	"github.com/joho/godotenv"
)

type WalletResponse struct {
	UserID    int32     `json:"user_id"`
	Amount    float64   `json:"amount"`
	Reference string    `json:"reference"`
	CreatedAt time.Time `json:"created_at"`
}

type WalletRequest struct {
	Amount    float64 `json:"amount"`
	Reference string  `json:"reference"`
}

type Wallet interface {
	Credit(context.Context, WalletRequest, string) (*WalletResponse, error)
	Debit(context.Context, WalletRequest, string) (*WalletResponse, error)
}

type wallet struct {
	httpClient *http.Client
}

func NewWalletService() Wallet {
	return &wallet{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *wallet) Credit(ctx context.Context, reqData WalletRequest, token string) (*WalletResponse, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	url := env.GetEnvString("WALLET_SERVICE", "") + env.GetEnvString("WALLET_BASE_PATH", "") + "/credit"

	payload := WalletRequest{
		Amount:    reqData.Amount,
		Reference: reqData.Reference,
	}

	log.Println(url)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet payload :%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request ;%w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet :%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body :%w", err)
	}

	log.Println("response body", string(body))
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("wallet service returned error :%s", string(body))
	}

	var walletResp WalletResponse
	if err := json.Unmarshal(body, &walletResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response :%w", err)
	}
	return &walletResp, nil
}

func (w *wallet) Debit(ctx context.Context, reqData WalletRequest, token string) (*WalletResponse, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	url := env.GetEnvString("WALLET_SERVICE", "") + env.GetEnvString("WALLET_BASE_PATH", "") + "/debit"

	payload := WalletRequest{
		Amount:    reqData.Amount,
		Reference: reqData.Reference,
	}

	//log.Println(url)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet payload :%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request ;%w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet :%w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body :%w", err)
	}

	log.Println("response body", string(body))
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("wallet service returned error :%s", string(body))
	}

	var walletResp WalletResponse
	if err := json.Unmarshal(body, &walletResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response :%w", err)
	}
	return &walletResp, nil
}
