package watchdog

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libbuildpack/v2/layers"
	"github.com/buildpacks/libbuildpack/v2/logger"

	"github.com/jromero/openfaas-cnb/pkg/config"
)

const defaultVersion = "0.7.6"
const executableName = "watchdog"

type metadata struct {
	Version string
}

type HttpClient interface {
	Get(url string) (*http.Response, error)
}

type LayerCreator struct {
	log        logger.Logger
	httpClient HttpClient
}

func NewLayerCreator(log logger.Logger, httpClient HttpClient) *LayerCreator {
	return &LayerCreator{
		log:        log,
		httpClient: httpClient,
	}
}

func (l *LayerCreator) Create(lyrs layers.Layers, conf config.Watchdog) (*layers.Layer, error) {
	watchdogLayer := lyrs.Layer(executableName)

	md := &metadata{}
	if err := watchdogLayer.ReadMetadata(md); err != nil {
		return nil, errors.New("read metadata: " + err.Error())
	}

	switch {
	case md.Version == conf.Version:
		l.log.Debug("using cache")
		return &watchdogLayer, nil
	case md.Version != "":
		if err := watchdogLayer.RemoveMetadata(); err != nil {
			return nil, errors.New("removing old metadata: " + err.Error())
		}
	}

	md.Version = conf.Version
	if err := watchdogLayer.WriteMetadata(&md, layers.Cache, layers.Launch); err != nil {
		return nil, errors.New("writing metadata: " + err.Error())
	}

	for key, value := range conf.Env {
		err := watchdogLayer.OverrideLaunchEnv(key, value)
		if err != nil {
			return nil, err
		}
	}

	downloadUrl := downloadUrl(conf.Version)
	l.log.Debug("downloading from: %s", downloadUrl)
	resp, err := l.httpClient.Get(downloadUrl)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	err = os.MkdirAll(watchdogLayer.Root, os.ModePerm)
	if err != nil {
		return nil, errors.New("creating layer dir: " + err.Error())
	}

	watchdogBin, err := os.Create(filepath.Join(watchdogLayer.Root, executableName))
	if err != nil {
		return nil, errors.New("creating binary: " + err.Error())
	}
	defer func() {
		_ = watchdogBin.Close()
	}()

	_, err = io.Copy(watchdogBin, resp.Body)
	if err != nil {
		return nil, errors.New("downloading watchdog: " + err.Error())
	}

	if err := os.Chmod(watchdogBin.Name(), os.ModePerm); err != nil {
		return nil, err
	}

	return &watchdogLayer, nil
}

func downloadUrl(version string) string {
	return fmt.Sprintf(
		"https://github.com/openfaas-incubator/of-watchdog/releases/download/%s/of-watchdog",
		version,
	)
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
