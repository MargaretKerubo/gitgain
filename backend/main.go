package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

var lnClient LightningClient

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables.")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "gitgain.db"
	}
	InitDB(dbPath)

	lnMode := os.Getenv("LIGHTNING_MODE")
	if lnMode == "lnd" {
		host := os.Getenv("LND_REST_HOST")
		macaroon := os.Getenv("LND_MACAROON_HEX")
		skipTLSStr := os.Getenv("LND_SKIP_TLS_VERIFY")
		skipTLS := skipTLSStr == "true" || skipTLSStr == "1"

		if host == "" || macaroon == "" {
			log.Fatal("LND_REST_HOST and LND_MACAROON_HEX are required when LIGHTNING_MODE=lnd")
		}
		lnClient = NewLndRestClient(host, macaroon, skipTLS)
		log.Println("Lightning service initialized in LND mode.")
	} else {
		lnClient = &MockLightningClient{}
		log.Println("Lightning service initialized in MOCK mode.")
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	api := app.Group("/api")
	api.Get("/stats", getStats)
	api.Post("/users", createUser)
	api.Get("/users/:username", getUser)
	api.Post("/challenges", createChallenge)
	api.Get("/challenges", listChallenges)
	api.Get("/challenges/:id", getChallenge)
	api.Post("/submissions", createSubmission)
	api.Get("/submissions", listSubmissions)
	api.Post("/admin/payout/:id", triggerPayout)
	api.Post("/webhooks/github", handleGithubWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(app.Listen(":" + port))
}

// Placeholder handlers

func getStats(c *fiber.Ctx) error {
	var totalPaid int64
	var activeChallenges int64
	var completedSubmissions int64

	// Count active challenges
	DB.Model(&Challenge{}).Where("status = ?", "active").Count(&activeChallenges)

	// Count completed submissions
	DB.Model(&Submission{}).Where("status = ?", "completed").Count(&completedSubmissions)

	// Compute total sats paid out
	var challenges []Challenge
	DB.Joins("JOIN submissions ON submissions.challenge_id = challenges.id").
		Where("submissions.status = ?", "completed").
		Find(&challenges)

	for _, ch := range challenges {
		totalPaid += ch.RewardSats
	}

	// Get wallet balance
	walletBalance, err := lnClient.GetBalance()
	if err != nil {
		log.Printf("Error fetching wallet balance: %v", err)
		walletBalance = 0
	}

	return c.JSON(fiber.Map{
		"total_sats_paid":       totalPaid,
		"active_challenges":     activeChallenges,
		"completed_submissions": completedSubmissions,
		"wallet_balance_sats":   walletBalance,
	})
}
func createUser(c *fiber.Ctx) error {
	type Request struct {
		Username         string `json:"username"`
		Email            string `json:"email"`
		GitHubUsername   string `json:"github_username"`
		LightningAddress string `json:"lightning_address"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Username == "" || req.Email == "" || req.GitHubUsername == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username, email, and github_username are required"})
	}

	// Check if user already exists
	var user User
	result := DB.Where("github_username = ?", req.GitHubUsername).First(&user)
	if result.Error == nil {
		// Update user if they exist
		user.Username = req.Username
		user.Email = req.Email
		user.LightningAddress = req.LightningAddress
		DB.Save(&user)
		return c.JSON(user)
	}

	// Create new user
	user = User{
		Username:         req.Username,
		Email:            req.Email,
		GitHubUsername:   req.GitHubUsername,
		LightningAddress: req.LightningAddress,
	}

	if err := DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func getUser(c *fiber.Ctx) error {
	username := c.Params("username")

	var user User
	if err := DB.Where("username = ? OR github_username = ?", username, username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}
func createChallenge(c *fiber.Ctx) error {
	type Request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		RepoOwner   string `json:"repo_owner"`
		RepoName    string `json:"repo_name"`
		RewardSats  int64  `json:"reward_sats"`
		CreatorID   uint   `json:"creator_id"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Title == "" || req.Description == "" || req.RepoOwner == "" || req.RepoName == "" || req.RewardSats <= 0 || req.CreatorID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid challenge parameter inputs"})
	}

	challenge := Challenge{
		Title:       req.Title,
		Description: req.Description,
		RepoOwner:   req.RepoOwner,
		RepoName:    req.RepoName,
		RewardSats:  req.RewardSats,
		CreatorID:   req.CreatorID,
		Status:      "active",
	}

	if err := DB.Create(&challenge).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(challenge)
}

func listChallenges(c *fiber.Ctx) error {
	var challenges []Challenge
	if err := DB.Preload("Creator").Find(&challenges).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(challenges)
}

func getChallenge(c *fiber.Ctx) error {
	id := c.Params("id")

	var challenge Challenge
	if err := DB.Preload("Creator").First(&challenge, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "challenge not found"})
	}

	return c.JSON(challenge)
}
func createSubmission(c *fiber.Ctx) error {
	type Request struct {
		ChallengeID       uint   `json:"challenge_id"`
		UserID            uint   `json:"user_id"`
		PullRequestURL    string `json:"pull_request_url"`
		PullRequestNumber int    `json:"pull_request_number"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.ChallengeID == 0 || req.UserID == 0 || req.PullRequestURL == "" || req.PullRequestNumber == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing submission details"})
	}

	submission := Submission{
		ChallengeID:       req.ChallengeID,
		UserID:            req.UserID,
		PullRequestURL:    req.PullRequestURL,
		PullRequestNumber: req.PullRequestNumber,
		Status:            "pending",
	}

	if err := DB.Create(&submission).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(submission)
}

func listSubmissions(c *fiber.Ctx) error {
	var submissions []Submission
	if err := DB.Preload("Challenge").Preload("User").Find(&submissions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(submissions)
}
func triggerPayout(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid submission id"})
	}

	var sub Submission
	if err := DB.Preload("Challenge").Preload("User").First(&sub, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "submission not found"})
	}

	if sub.Status == "completed" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "submission payout has already been sent"})
	}

	// Trigger payment using user's Lightning Address (Option A)
	if sub.User.LightningAddress == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user does not have a lightning address configured"})
	}

	preimage, err := lnClient.PayToLightningAddress(sub.User.LightningAddress, sub.Challenge.RewardSats)
	if err != nil {
		sub.Status = "failed"
		sub.ErrorMessage = err.Error()
		DB.Save(&sub)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   fmt.Sprintf("lightning payout failed: %v", err),
			"details": sub,
		})
	}

	// Update status on success
	sub.Status = "completed"
	sub.PaymentHash = preimage // Store preimage
	sub.ErrorMessage = ""
	DB.Save(&sub)

	// Optionally mark the challenge status as completed if needed
	sub.Challenge.Status = "completed"
	DB.Save(&sub.Challenge)

	return c.JSON(fiber.Map{
		"message":      "payout sent successfully",
		"preimage":     preimage,
		"submission":   sub,
	})
}


