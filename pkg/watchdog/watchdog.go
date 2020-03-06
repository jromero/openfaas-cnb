package watchdog

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/buildpacks/libbuildpack/v2/layers"
	"github.com/buildpacks/libbuildpack/v2/logger"
)

const (
	executableName = "watchdog"
)

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

func (l *Contributor) Contribute(lyrs layers.Layers, conf Config) (*layers.Layer, error) {
	watchdogLayer := lyrs.Layer(executableName)

	if err := l.installBinaries(watchdogLayer, conf.Version); err != nil {
		return nil, err
	}

	if err := l.configureApp(lyrs, watchdogLayer, conf.ProcessType); err != nil {
		return nil, err
	}

	return &watchdogLayer, nil
}

func (l *Contributor) installBinaries(watchdogLayer layers.Layer, version string) error {
	wdMD := &metadata{}
	if err := watchdogLayer.ReadMetadata(wdMD); err != nil {
		return errors.New("read metadata: " + err.Error())
	}

	switch {
	case wdMD.Version == version:
		l.log.Debug("using cache")
	case wdMD.Version != "":
		if err := watchdogLayer.RemoveMetadata(); err != nil {
			return errors.New("removing old metadata: " + err.Error())
		}
		fallthrough
	default:
		if err := l.downloadWatchdog(version, watchdogLayer.Root); err != nil {
			return errors.New("downloading binary: " + err.Error())
		}
	}

	wdMD.Version = version
	if err := watchdogLayer.WriteMetadata(&wdMD, layers.Cache, layers.Launch); err != nil {
		return errors.New("writing metadata: " + err.Error())
	}

	return nil
}

// configureApp configures the application
func (l *Contributor) configureApp(lyrs layers.Layers, watchdogLayer layers.Layer, processType string) error {
	err := watchdogLayer.DefaultLaunchEnv("function_process", fmt.Sprintf("/cnb/lifecycle/launcher %s", processType))
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

	if resp.StatusCode != 200 {
		return fmt.Errorf("downloading from '%s' returned status code '%d'", downloadUrl, resp.StatusCode)
	}

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

