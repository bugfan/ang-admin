package store

import (
	"github.com/bugfan/empty015/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// Init initializes the SQLite database
func Init(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	return DB.AutoMigrate(
		&models.Book{},
		&models.Chunk{},
		&models.Config{},
	)
}

// GetConfig returns the current LLM configuration
func GetConfig() (*models.Config, error) {
	var cfg models.Config
	result := DB.First(&cfg)
	if result.Error == gorm.ErrRecordNotFound {
		// Return default config
		return &models.Config{
			Provider:    "openai",
			ChatModel:   "gpt-4o-mini",
			EmbedModel:  "text-embedding-3-small",
			TopK:        5,
			ScoreThresh: 0.3,
		}, nil
	}
	return &cfg, result.Error
}

// SaveConfig saves the LLM configuration
func SaveConfig(cfg *models.Config) error {
	var existing models.Config
	result := DB.First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		cfg.ID = 1
		return DB.Create(cfg).Error
	}
	cfg.ID = existing.ID
	return DB.Save(cfg).Error
}
