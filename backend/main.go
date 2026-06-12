package main

import (
	"log"
	"os"

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

func getStats(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func createUser(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func getUser(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func createChallenge(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func listChallenges(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func getChallenge(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func createSubmission(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func listSubmissions(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func triggerPayout(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
func handleGithubWebhook(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNotImplemented) }
