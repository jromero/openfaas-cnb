package watchdog

import (
	"io"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	configName         = "watchdog.toml"
	defaultProcessType = "web"
	defaultVersion     = "0.7.6"
)

type configTOML struct {
	Watchdog Config `toml:"watchdog"`
}

func ParseConfig(reader io.Reader) (Config, error) {
	cTOML := &configTOML{}
	if _, err := toml.DecodeReader(reader, &cTOML); err != nil {
		return cTOML.Watchdog, err
	}

	if cTOML.Watchdog.Version == "" {
		cTOML.Watchdog.Version = defaultVersion
	}

	if cTOML.Watchdog.ProcessType == "" {
		cTOML.Watchdog.ProcessType = defaultProcessType
	}

	return cTOML.Watchdog, nil
}

func DefaultConfig() Config {
	return Config{
		Version:     defaultVersion,
		ProcessType: defaultProcessType,
	}
}

type Config struct {
	Version     string `toml:"version"`
	ProcessType string `toml:"process_type"`
}

func ConfigPath(appDir string) string {
	return filepath.Join(appDir, configName)
}
