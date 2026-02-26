package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()

	// Temporarily change HOME to point to temp directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()

	if cfg.DBFile == "" {
		t.Error("Expected DBFile to be set")
	}

	if cfg.LookBackDays == 0 {
		t.Error("Expected LookBackDays to be set")
	}

	if cfg.LogLength == 0 {
		t.Error("Expected LogLength to be set")
	}

	if cfg.CaptainDir == "" {
		t.Error("Expected CaptainDir to be set")
	}

	// Verify config file was created
	cfgPath := filepath.Join(tmpDir, ".captain", "config.ini")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

func TestConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := LoadConfig()

	expectedDefaults := map[string]interface{}{
		"DBFile":       "testdo.db",
		"LookBackDays": 7,
		"LogLength":    10,
	}

	if cfg.DBFile != expectedDefaults["DBFile"] {
		t.Errorf("Expected DBFile to be %s, got %s", expectedDefaults["DBFile"], cfg.DBFile)
	}

	if cfg.LookBackDays != expectedDefaults["LookBackDays"] {
		t.Errorf("Expected LookBackDays to be %d, got %d", expectedDefaults["LookBackDays"], cfg.LookBackDays)
	}

	if cfg.LogLength != expectedDefaults["LogLength"] {
		t.Errorf("Expected LogLength to be %d, got %d", expectedDefaults["LogLength"], cfg.LogLength)
	}
}

func TestConfigSet(t *testing.T) {
	tmpDir := t.TempDir()
	capDir := filepath.Join(tmpDir, ".captain")
	os.MkdirAll(capDir, os.ModePerm)

	cfg := Config{
		DBFile:       "test.db",
		LookBackDays: 7,
		LogLength:    10,
		CaptainDir:   capDir,
	}

	// Create initial config file
	cfgPath := filepath.Join(capDir, "config.ini")
	file := ini.Empty()
	file.Section("").Key("profile").SetValue("main")
	file.SaveTo(cfgPath)

	// Test setting a value
	err := cfg.Set("profile", "production")
	if err != nil {
		t.Errorf("Failed to set config value: %v", err)
	}

	// Verify the value was set
	loadedFile, err := ini.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	profile := loadedFile.Section("").Key("profile").String()
	if profile != "production" {
		t.Errorf("Expected profile to be 'production', got %s", profile)
	}
}

func TestConfigSetProfile(t *testing.T) {
	tmpDir := t.TempDir()
	capDir := filepath.Join(tmpDir, ".captain")
	os.MkdirAll(capDir, os.ModePerm)

	cfg := Config{
		DBFile:       "test.db",
		LookBackDays: 7,
		LogLength:    10,
		CaptainDir:   capDir,
	}

	// Create initial config file with profile
	cfgPath := filepath.Join(capDir, "config.ini")
	file := ini.Empty()
	file.Section("").Key("profile").SetValue("main")
	file.NewSection("main")
	file.SaveTo(cfgPath)

	// Test setting a profile value
	err := cfg.SetProfile("log_length", "25")
	if err != nil {
		t.Errorf("Failed to set profile value: %v", err)
	}

	// Verify the value was set
	loadedFile, err := ini.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	logLength := loadedFile.Section("main").Key("log_length").String()
	if logLength != "25" {
		t.Errorf("Expected log_length to be '25', got %s", logLength)
	}
}

func TestCaptainDirCreation(t *testing.T) {
	tmpDir := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	LoadConfig()

	capDir := filepath.Join(tmpDir, ".captain")
	if _, err := os.Stat(capDir); os.IsNotExist(err) {
		t.Error("Expected .captain directory to be created")
	}
}
