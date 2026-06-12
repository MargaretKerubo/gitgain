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

type WebhookPayload struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest *struct {
		HTMLURL string `json:"html_url"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		Head    struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
	WorkflowRun *struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		HeadBranch string `json:"head_branch"`
		HeadSHA    string `json:"head_sha"`
	} `json:"workflow_run"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

func handleGithubWebhook(c *fiber.Ctx) error {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	sigHeader := c.Get("X-Hub-Signature-256")
	payload := c.Body()

	if !VerifySignature(payload, sigHeader, secret) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
	}

	var event WebhookPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to parse payload"})
	}

	eventType := c.Get("X-GitHub-Event")
	log.Printf("Received GitHub webhook event: %s (action: %s)", eventType, event.Action)

	switch eventType {
	case "pull_request":
		return handlePullRequest(event)
	case "workflow_run":
		return handleWorkflowRun(event)
	}

	return c.SendStatus(fiber.StatusOK)
}

func handlePullRequest(event WebhookPayload) error {
	// Implemented in next commit
	return nil
}

func handleWorkflowRun(event WebhookPayload) error {
	// Implemented in next commit
	return nil
}
