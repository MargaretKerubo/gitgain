package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// VerifySignature validates GitHub webhook HMAC-SHA256 signature
func VerifySignature(payload []byte, sigHeader string, secret string) bool {
	if secret == "" {
		return true // Skip verification if secret is not set
	}
	if !strings.HasPrefix(sigHeader, "sha256=") {
		return false
	}
	actualSig, err := hex.DecodeString(sigHeader[7:])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSig := mac.Sum(nil)
	return hmac.Equal(actualSig, expectedSig)
}
