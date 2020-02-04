package main

import (
	"net/http"
	"os"

	"github.com/buildpacks/libbuildpack/v2/build"
	"github.com/buildpacks/libbuildpack/v2/layers"

	"github.com/jromero/openfaas-cnb/cmd"
	"github.com/jromero/openfaas-cnb/pkg/watchdog"
)

func main() {
	b, err := build.DefaultBuild()
	if err != nil {
		cmd.Exit(cmd.UnexpectedError, err)
	}

	conf, err := watchdog.ParseConfig(b.Application.Root)
	if err != nil {
		b.Logger.Info(err.Error())
		os.Exit(b.Failure(cmd.ParseConfigError))
	}

	layerCreator := watchdog.NewLayerCreator(b.Logger, http.DefaultClient)

	watchdogLayer, err := layerCreator.Create(b.Layers, conf.Watchdog)
	if err != nil {
		b.Logger.Info(err.Error())
		os.Exit(b.Failure(cmd.LayerCreationError))
	}

	err = b.Layers.WriteApplicationMetadata(layers.Metadata{
		Processes: []layers.Process{watchdog.Process(watchdogLayer.Root)},
	})
	if err != nil {
		b.Logger.Info(err.Error())
		os.Exit(b.Failure(cmd.UnexpectedError))
	}
}
