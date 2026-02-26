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
	Task  DoType = "task"
	Ask   DoType = "ask"
	Tell  DoType = "tell"
	Brag  DoType = "brag"
	Learn DoType = "learn"
	PR    DoType = "PR"
	Meta  DoType = "meta"
)

type DoPrio string

const (
	Low    DoPrio = "low"
	Medium DoPrio = "medium"
	High   DoPrio = "high"
)

type Do struct {
	ID          uint      `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"default:current_timestamp"`
	CompletedAt *time.Time
	Completed   bool   `gorm:"default:false"`
	Pinned      bool   `gorm:"default:false"`
	Sensitive   bool   `gorm:"default:false"`
	Description string `gorm:"not null"`
	Type        DoType `gorm:"type:TEXT;not null"`
	Priority    DoPrio `gorm:"type:TEXT;not null;default:medium"`
	Deleted     bool   `gorm:"default:false"`
	Reason      string `gorm:"type:TEXT"`
	Doc         DoDoc  `gorm:"foreignKey:DoID"`
	Tags        []Tag  `gorm:"many2many:do_tags;"`
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
	DoID  uint `gorm:"primaryKey;not null"`
	TagID uint `gorm:"primaryKey;not null"`
	Do    Do   `gorm:"foreignKey:DoID"`
	Tag   Tag  `gorm:"foreignKey:TagID"`
}

type Template struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"unique;not null"`
	Content   string    `gorm:"type:TEXT;not null"`
	Deleted   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"default:current_timestamp"`
	UpdatedAt time.Time `gorm:"default:current_timestamp"`
}

func OpenConn(cfg *Config) *gorm.DB {
	dbPath := fmt.Sprintf("%s/%s", cfg.CaptainDir, cfg.DBFile)
	conn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("could not open database: %v", err)
	}

	err = conn.AutoMigrate(&Do{}, &Tag{}, &DoTag{}, &DoDoc{}, &Template{})
	if err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	return conn
}
