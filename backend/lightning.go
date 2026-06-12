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
