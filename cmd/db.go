package cmd

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DoType string

const (
	Task DoType = "task"
	Ask  DoType = "ask"
	Tell DoType = "tell"
	Brag DoType = "brag"
)

type DoPrio string

const (
	Low    DoPrio = "low"
	Medium DoPrio = "medium"
	High   DoPrio = "high"
)

type Do struct {
	ID          uint           `gorm:"primaryKey"`
	CreatedAt   time.Time      `gorm:"default:current_timestamp"`
	CompletedAt *time.Time
	Completed   bool           `gorm:"default:false"`
	Description string         `gorm:"not null"`
	Type        DoType         `gorm:"type:TEXT;not null"`
	Priority    DoPrio         `gorm:"type:TEXT;not null;default:medium"`
	Deleted     bool           `gorm:"default:false"`
	Doc         DoDoc          `gorm:"foreignKey:DoID"`
}

func (DoType) GormDataType() string {
	return "string"
}

type DoDoc struct {
	ID   uint   `gorm:"primaryKey"`
	DoID uint   `gorm:"not null"`
	Text string `gorm:"type:TEXT;not null"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique:not null"`
}

type DoTag struct {
	DoID  uint `gorm:"not null"`
	TagID uint `gorm:"not null"`
}

func OpenConn(cfg *Config) *gorm.DB {
	dbPath := fmt.Sprintf("%s/%s", cfg.CaptainDir, cfg.DBFile)
	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("could not open database: %v", err)
	}

	err = conn.AutoMigrate(&Do{}, &Tag{}, &DoTag{}, &DoDoc{})
	if err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	return conn
}
