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

type Contributor struct {
	log        logger.Logger
	httpClient HttpClient
}

func NewContributor(log logger.Logger, httpClient HttpClient) *Contributor {
	return &Contributor{
		log:        log,
		httpClient: httpClient,
	}
}

func (l *Contributor) Contribute(lyrs layers.Layers, conf config.Watchdog) (*layers.Layer, error) {
	watchdogLayer := lyrs.Layer(executableName)

	if err := l.installBinaries(watchdogLayer, conf); err != nil {
		return nil, err
	}

	if err := l.configureApp(watchdogLayer, lyrs); err != nil {
		return nil, err
	}

	if err := l.addEnvVars(watchdogLayer, conf.Env); err != nil {
		return nil, err
	}

	return &watchdogLayer, nil
}

func (l *Contributor) installBinaries(watchdogLayer layers.Layer, conf config.Watchdog) error {
	wdMD := &metadata{}
	if err := watchdogLayer.ReadMetadata(wdMD); err != nil {
		return errors.New("read metadata: " + err.Error())
	}

	switch {
	case wdMD.Version == conf.Version:
		l.log.Debug("using cache")
	case wdMD.Version != "":
		if err := watchdogLayer.RemoveMetadata(); err != nil {
			return errors.New("removing old metadata: " + err.Error())
		}
		fallthrough
	default:
		if err := l.downloadWatchdog(conf.Version, watchdogLayer.Root); err != nil {
			return errors.New("downloading binary: " + err.Error())
		}
	}

	wdMD.Version = conf.Version
	if err := watchdogLayer.WriteMetadata(&wdMD, layers.Cache, layers.Launch); err != nil {
		return errors.New("writing metadata: " + err.Error())
	}

	return nil
}

func (l *Contributor) addEnvVars(watchdogLayer layers.Layer, env config.Env) error {
	for key, value := range env {
		err := watchdogLayer.DefaultLaunchEnv(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// configureApp configures the application
func (l *Contributor) configureApp(watchdogLayer layers.Layer, lyrs layers.Layers) error {
	err := watchdogLayer.DefaultLaunchEnv("function_process", "/cnb/lifecycle/launcher web")
	if err != nil {
		return errors.New("writing function_process env var: " + err.Error())
	}

	err = lyrs.WriteApplicationMetadata(layers.Metadata{
		Processes: []layers.Process{{
			Type:    "faas",
			Command: filepath.Join(watchdogLayer.Root, executableName),
			Args:    nil,
			Direct:  false,
		}},
		Slices: nil,
	})
	if err != nil {
		return errors.New("writing app metadata file: " + err.Error())
	}

	return nil
}

func (l *Contributor) downloadWatchdog(version string, layerDir string) error {
	downloadUrl := fmt.Sprintf(
		"https://github.com/openfaas-incubator/of-watchdog/releases/download/%s/of-watchdog",
		version,
	)
	l.log.Debug("downloading from: %s", downloadUrl)
	resp, err := l.httpClient.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	err = os.MkdirAll(layerDir, os.ModePerm)
	if err != nil {
		return errors.New("creating layer dir: " + err.Error())
	}

	watchdogBin, err := os.Create(filepath.Join(layerDir, executableName))
	if err != nil {
		return errors.New("creating binary: " + err.Error())
	}
	defer func() {
		_ = watchdogBin.Close()
	}()

	_, err = io.Copy(watchdogBin, resp.Body)
	if err != nil {
		return errors.New("downloading watchdog: " + err.Error())
	}

	if err := os.Chmod(watchdogBin.Name(), os.ModePerm); err != nil {
		return err
	}

	return nil
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
