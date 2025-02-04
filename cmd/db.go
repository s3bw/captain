package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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

type Do struct {
	ID          uint      `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"default:current_timestamp"`
	CompletedAt *time.Time
	Completed   bool           `gorm:"default:false"`
	Description string         `gorm:"not null"`
	Type        DoType         `gorm:"type:TEXT;not null"`
	Docs        sql.NullString `gorm:"type:TEXT"`
	Deleted     bool           `gorm:"default:false"`
}

func (DoType) GormDataType() string {
	return "string"
}

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique:not null"`
}

type DoTag struct {
	DoID  uint `gorm:"not null"`
	TagID uint `gorm:"not null"`
}

func OpenConn() *gorm.DB {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("failed to get the user's home directory")
	}

	dbDir := fmt.Sprintf("%s/.captain", homeDir)
	dbPath := fmt.Sprintf("%s/testdo.db", dbDir)

	err = os.MkdirAll(dbDir, os.ModePerm)
	if err != nil {
		panic("failed to create directory")
	}

	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("could not open database: %v", err)
	}

	err = conn.AutoMigrate(&Do{}, &Tag{}, &DoTag{})
	if err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	return conn
}
