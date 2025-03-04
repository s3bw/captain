package cmd

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

type IniConfig struct {
	Profile string `ini:"profile"`
}

type Config struct {
	DBFile       string `ini:"dbname"`
	LookBackDays int    `ini:"lookback_days"`
	LogLength    int    `ini:"log_length"`
	CaptainDir   string
}

func (cfg *Config) Set(key string, value string) error {
	cfgFile := fmt.Sprintf("%s/config.ini", cfg.CaptainDir)

	file, err := ini.LooseLoad(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to save:%w", err)
	}

	file.Section("").Key(key).SetValue(value)

	err = file.SaveTo(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to save:%w", err)
	}
	return nil
}

func LoadConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("failed to get the user's home directory")
	}

	capDir := fmt.Sprintf("%s/.captain", homeDir)
	err = os.MkdirAll(capDir, os.ModePerm)
	if err != nil {
		panic("failed to create directory")
	}

	cfgFile := fmt.Sprintf("%s/config.ini", capDir)

	// Default profile configuration
	iniCfg := IniConfig{Profile: "main"}

	// Default config values
	cfg := Config{
		DBFile:       "testdo.db",
		LookBackDays: 7,
		LogLength:    10,
		CaptainDir:   capDir,
	}

	// If config file exists, try to load it
	var file *ini.File
	if _, err := os.Stat(cfgFile); err == nil {
		// Load existing file to preserve other profiles
		file, err = ini.Load(cfgFile)
		if err != nil {
			panic("Failed to load config file")
		}
	} else {
		// Create new file if it doesn't exist
		file = ini.Empty()
	}

	// Load profile from root section
	err = file.Section("").MapTo(&iniCfg)
	if err == nil {
		// Try to load config from the specified profile
		err = file.Section(iniCfg.Profile).MapTo(&cfg)
		if err != nil {
			// Profile doesn't exist, will create it with defaults
			fmt.Printf("Profile '%s' not found, creating with defaults\n", iniCfg.Profile)
		}
	}

	// Save the profile in root section
	file.Section("").ReflectFrom(&iniCfg)
	// Save the config values in profile section
	err = file.Section(iniCfg.Profile).ReflectFrom(&cfg)
	if err != nil {
		panic("Error reflecting config values")
	}

	err = file.SaveTo(cfgFile)
	if err != nil {
		panic("Error saving config file")
	}

	return cfg
}

// SetProfile sets a value in the current profile section
func (c *Config) SetProfile(key, value string) error {
	cfgFile := fmt.Sprintf("%s/config.ini", c.CaptainDir)

	file, err := ini.LooseLoad(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current profile from root section
	profile := file.Section("").Key("profile").String()
	if profile == "" {
		return fmt.Errorf("no profile selected")
	}

	// Ensure the profile section exists
	if !file.HasSection(profile) {
		file.NewSection(profile)
	}

	// Set the value in the profile section
	file.Section(profile).Key(key).SetValue(value)

	err = file.SaveTo(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
