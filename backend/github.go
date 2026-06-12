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
	"strconv"
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
	if event.PullRequest == nil {
		return nil
	}

	// Only process opened, synchronized, or reopened pull requests
	action := event.Action
	if action != "opened" && action != "reopened" && action != "synchronize" {
		return nil
	}

	// Parse challenge ID from PR body
	re := regexp.MustCompile(`(?:gitgain\s+#|challenge\s+#|#challenge-)(\d+)`)
	matches := re.FindStringSubmatch(event.PullRequest.Body)
	if len(matches) < 2 {
		log.Println("PR does not link a challenge ID in its body.")
		return nil
	}

	challengeID, err := strconv.Atoi(matches[1])
	if err != nil {
		return err
	}

	var challenge Challenge
	if err := DB.First(&challenge, challengeID).Error; err != nil {
		log.Printf("Challenge #%d not found in database: %v", challengeID, err)
		return nil
	}

	// Validate repo details match challenge (case-insensitive check)
	if strings.ToLower(event.Repository.Owner.Login) != strings.ToLower(challenge.RepoOwner) ||
		strings.ToLower(event.Repository.Name) != strings.ToLower(challenge.RepoName) {
		log.Printf("Repository name mismatch for challenge #%d: expected %s/%s, got %s/%s",
			challengeID, challenge.RepoOwner, challenge.RepoName, event.Repository.Owner.Login, event.Repository.Name)
		return nil
	}

	// Find the user who created the PR by their GitHub username (stored in git_hub_username)
	var user User
	if err := DB.Where("git_hub_username = ?", event.PullRequest.User.Login).First(&user).Error; err != nil {
		log.Printf("User with GitHub username %s not found, submission cannot be registered", event.PullRequest.User.Login)
		return nil
	}

	// Check if a submission already exists
	var sub Submission
	result := DB.Where("challenge_id = ? AND user_id = ?", challenge.ID, user.ID).First(&sub)
	if result.Error == nil {
		// Update existing submission details
		sub.PullRequestURL = event.PullRequest.HTMLURL
		sub.PullRequestNumber = event.PullRequest.Number
		sub.Status = "pending" // reset status for verification
		sub.ErrorMessage = ""
		DB.Save(&sub)
		log.Printf("Updated existing submission ID %d to pending.", sub.ID)
		return nil
	}

	// Create new submission
	sub = Submission{
		ChallengeID:       challenge.ID,
		UserID:            user.ID,
		PullRequestURL:    event.PullRequest.HTMLURL,
		PullRequestNumber: event.PullRequest.Number,
		Status:            "pending",
	}

	if err := DB.Create(&sub).Error; err != nil {
		return err
	}

	log.Printf("Registered new submission ID %d for challenge #%d by %s", sub.ID, challenge.ID, user.Username)
	return nil
}

func handleWorkflowRun(event WebhookPayload) error {
	if event.WorkflowRun == nil {
		return nil
	}

	// Only process completed successful runs
	if event.WorkflowRun.Status != "completed" || event.WorkflowRun.Conclusion != "success" {
		return nil
	}

	// Find pending/failed submissions for this repository
	var subs []Submission
	err := DB.Preload("Challenge").Preload("User").
		Joins("JOIN challenges ON challenges.id = submissions.challenge_id").
		Where("challenges.repo_owner = ? AND challenges.repo_name = ? AND submissions.status IN (?, ?)",
			event.Repository.Owner.Login, event.Repository.Name, "pending", "failed").
		Find(&subs).Error

	if err != nil {
		return err
	}

	for _, sub := range subs {
		// Fetch PR head branch to verify it matches this workflow head branch
		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d",
			sub.Challenge.RepoOwner, sub.Challenge.RepoName, sub.PullRequestNumber)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			continue
		}

		if pat := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN"); pat != "" {
			req.Header.Set("Authorization", "token "+pat)
		}
		req.Header.Set("User-Agent", "gitgain-backend")

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}

		var prDetails struct {
			Head struct {
				Ref string `json:"ref"`
			} `json:"head"`
		}

		json.NewDecoder(resp.Body).Decode(&prDetails)
		resp.Body.Close()

		if prDetails.Head.Ref != event.WorkflowRun.HeadBranch {
			continue // Not the PR that triggered the workflow run
		}

		log.Printf("Verification PASSED for submission ID %d (PR #%d)", sub.ID, sub.PullRequestNumber)

		// Set status to verifications_passed
		sub.Status = "verifications_passed"
		DB.Save(&sub)

		// Trigger automatic Lightning payment!
		if sub.User.LightningAddress == "" {
			sub.Status = "failed"
			sub.ErrorMessage = "user lightning address is not configured"
			DB.Save(&sub)
			continue
		}

		preimage, err := lnClient.PayToLightningAddress(sub.User.LightningAddress, sub.Challenge.RewardSats)
		if err != nil {
			sub.Status = "failed"
			sub.ErrorMessage = fmt.Sprintf("auto-payout failed: %v", err)
			DB.Save(&sub)
			log.Printf("Auto-payout failed for submission ID %d: %v", sub.ID, err)
			continue
		}

		// Success payout update
		sub.Status = "completed"
		sub.PaymentHash = preimage
		sub.ErrorMessage = ""
		DB.Save(&sub)

		// Mark challenge as completed
		sub.Challenge.Status = "completed"
		DB.Save(&sub.Challenge)

		log.Printf("Auto-payout SUCCESSFUL for submission ID %d. Paid %d sats to %s. Preimage: %s",
			sub.ID, sub.Challenge.RewardSats, sub.User.LightningAddress, preimage)
	}

	return nil
}
