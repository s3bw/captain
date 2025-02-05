package cmd

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

type Config struct {
	DBFile     string `ini:"dbname"`
	CaptainDir string
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

	// Default values
	cfg := Config{DBFile: "testdo.db", CaptainDir: capDir}

	cfgFile := fmt.Sprintf("%s/config.ini", capDir)
	if _, err := os.Stat(cfgFile); err == nil {
		file, _ := ini.Load(cfgFile)
		err = file.Section("").MapTo(&cfg)
		if err != nil {
			panic("Failed mapping config values")
		}
	}

	file := ini.Empty()
	section := file.Section("")

	err = section.ReflectFrom(&cfg)
	if err != nil {
		panic("Error reflecting config values")
	}

	err = file.SaveTo(cfgFile)
	if err != nil {
		panic("Error saving config file")
	}

	return cfg
}
