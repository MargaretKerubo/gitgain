package main

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Username         string         `gorm:"uniqueIndex;not null" json:"username"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	GitHubUsername   string         `gorm:"uniqueIndex;not null" json:"github_username"`
	LightningAddress string         `json:"lightning_address"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type Challenge struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"not null" json:"description"`
	RepoOwner   string         `gorm:"not null" json:"repo_owner"`
	RepoName    string         `gorm:"not null" json:"repo_name"`
	RewardSats  int64          `gorm:"not null" json:"reward_sats"`
	Status      string         `gorm:"default:'active'" json:"status"` // active, completed
	CreatorID   uint           `gorm:"not null" json:"creator_id"`
	Creator     User           `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Submission struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	ChallengeID       uint           `gorm:"not null" json:"challenge_id"`
	Challenge         Challenge      `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
	UserID            uint           `gorm:"not null" json:"user_id"`
	User              User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PullRequestURL    string         `gorm:"not null" json:"pull_request_url"`
	PullRequestNumber int            `gorm:"not null" json:"pull_request_number"`
	Status            string         `gorm:"default:'pending'" json:"status"` // pending, verifications_passed, completed, failed
	PaymentHash       string         `json:"payment_hash"`
	ErrorMessage      string         `json:"error_message"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto Migrate the schemas
	err = DB.AutoMigrate(&User{}, &Challenge{}, &Submission{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	log.Println("Database connection established and models migrated.")
}
