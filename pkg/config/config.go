package config

import (
	"os"
	"path/filepath"
)

const filename = "watchdog.toml"

type Config struct {
	Watchdog Watchdog `toml:"watchdog"`
}

type Watchdog struct {
	Version     string `toml:"version"`
	ProcessType string `toml:"process_type"`
	Env         Env    `toml:"env"`
}

type Env map[string]string

func Filename(appDir string) string {
	return filepath.Join(appDir, filename)
}

func Exists(configPath string) (exists bool, err error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
