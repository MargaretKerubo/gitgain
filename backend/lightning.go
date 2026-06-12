package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LightningClient interface {
	SendPayment(invoice string) (string, error)
	GetBalance() (int64, error)
	PayToLightningAddress(address string, amountSats int64) (string, error)
}

// MockLightningClient simulates Lightning node
type MockLightningClient struct{}

func (m *MockLightningClient) GetBalance() (int64, error) {
	return 21000000, nil // 21M sats mock balance
}

func (m *MockLightningClient) SendPayment(invoice string) (string, error) {
	log.Printf("[MOCK LIGHTNING] Simulating payment of invoice: %s", invoice)
	time.Sleep(1 * time.Second) // Simulate network delay
	return "mock_preimage_settled_successfully_12345", nil
}

func (m *MockLightningClient) PayToLightningAddress(address string, amountSats int64) (string, error) {
	log.Printf("[MOCK LIGHTNING] Simulating payment of %d sats to %s", amountSats, address)
	time.Sleep(1 * time.Second)
	return fmt.Sprintf("mock_preimage_paid_to_%s_%d_sats", address, amountSats), nil
}

// LndRestClient connects to Polar LND REST API
type LndRestClient struct {
	Host           string
	MacaroonHex    string
	SkipTLSVerify  bool
	httpClient     *http.Client
}

func NewLndRestClient(host, macaroonHex string, skipTLS bool) *LndRestClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	// Clean host URL
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	return &LndRestClient{
		Host:          host,
		MacaroonHex:   macaroonHex,
		SkipTLSVerify: skipTLS,
		httpClient:    client,
	}
}

func (l *LndRestClient) GetBalance() (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/balance/channels", l.Host), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Grpc-Metadata-macaroon", l.MacaroonHex)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to get LND balance (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		Balance            string `json:"balance"`
		PendingOpenBalance string `json:"pending_open_balance"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Convert balance from string to int64
	var balance int64
	_, err = fmt.Sscanf(data.Balance, "%d", &balance)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %v", err)
	}

	return balance, nil
}

func (l *LndRestClient) SendPayment(invoice string) (string, error) {
	payload := map[string]interface{}{
		"payment_request": invoice,
		"fee_limit": map[string]interface{}{
			"fixed": 1000, // max fee 1000 sats
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/router/send", l.Host), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Grpc-Metadata-macaroon", l.MacaroonHex)
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to send payment (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Response is a stream of updates. Read chunks until we find "SUCCEEDED" or "FAILED".
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		var update struct {
			Result struct {
				Status          string `json:"status"`
				PaymentHash     string `json:"payment_hash"`
				PaymentPreimage string `json:"payment_preimage"`
				FailureReason   string `json:"failure_reason"`
			} `json:"result"`
		}

		if err := json.Unmarshal(line, &update); err != nil {
			// Some LND versions output a flatter structure, check that too
			var flatUpdate struct {
				Status          string `json:"status"`
				PaymentHash     string `json:"payment_hash"`
				PaymentPreimage string `json:"payment_preimage"`
				FailureReason   string `json:"failure_reason"`
			}
			if errFlat := json.Unmarshal(line, &flatUpdate); errFlat == nil && flatUpdate.Status != "" {
				update.Result = flatUpdate
			} else {
				continue // Skip malformed lines
			}
		}

		if update.Result.Status == "SUCCEEDED" {
			return update.Result.PaymentPreimage, nil
		} else if update.Result.Status == "FAILED" {
			return "", fmt.Errorf("payment failed: %s", update.Result.FailureReason)
		}
	}

	return "", fmt.Errorf("payment stream ended without confirmation status")
}
