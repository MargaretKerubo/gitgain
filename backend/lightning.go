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
