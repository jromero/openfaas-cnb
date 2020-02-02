package watchdog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libbuildpack/v2/layers"

	"github.com/jromero/openfaas-cnb/pkg/config"
)

const defaultVersion = "0.7.6"
const executableName = "watchdog"

func DownloadUrl(version string) string {
	return fmt.Sprintf(
		"https://github.com/openfaas-incubator/of-watchdog/releases/download/%s/of-watchdog",
		version,
	)
}

func CreateWatchdogLayer(lyrs layers.Layers, conf config.Watchdog) (*layers.Layer, error) {
	watchdogLayer := lyrs.Layer(executableName)
	err := watchdogLayer.WriteMetadata(nil, layers.Cache, layers.Launch)
	if err != nil {
		return nil, err
	}

	for key, value := range conf.Env {
		err := watchdogLayer.OverrideLaunchEnv(key, value)
		if err != nil {
			return nil, err
		}
	}

	resp, err := http.Get(DownloadUrl(conf.Version))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	watchdogBin, err := os.Create(filepath.Join(watchdogLayer.Root, executableName))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = watchdogBin.Close()
	}()

	_, err = io.Copy(watchdogBin, resp.Body)
	if err != nil {
		return nil, err
	}

	err = os.Chmod(watchdogBin.Name(), os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &watchdogLayer, nil
}

func Process(watchdogLayerDir string) layers.Process {
	return layers.Process{
		Type:    "web",
		Command: filepath.Join(watchdogLayerDir, executableName),
		Args:    nil,
		Direct:  false,
	}
}

func ParseConfig(appDir string) (*config.Config, error) {
	conf := &config.Config{}
	if _, err := toml.DecodeFile(config.Filename(appDir), conf); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if conf.Watchdog.Version == "" {
		conf.Watchdog.Version = defaultVersion
	}

	return conf, nil
}
