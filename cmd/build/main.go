package main

import (
	"net/http"
	"os"

	"github.com/buildpacks/libbuildpack/v2/build"

	"github.com/jromero/openfaas-cnb/cmd"
	"github.com/jromero/openfaas-cnb/pkg/config"
	"github.com/jromero/openfaas-cnb/pkg/watchdog"
)

func main() {
	b, err := build.DefaultBuild()
	if err != nil {
		cmd.Exit(cmd.UnexpectedError, err)
	}

	fh, err := os.Open(config.Filename(b.Application.Root))
	if err != nil {
		b.Logger.Info(err.Error())
		os.Exit(b.Failure(cmd.ParseConfigError))
	}
	defer fh.Close()

	conf, err := watchdog.ParseConfig(fh)
	if err != nil {
		b.Logger.Info(err.Error())
		os.Exit(b.Failure(cmd.ParseConfigError))
	}

	layerCreator := watchdog.NewContributor(b.Logger, http.DefaultClient)
	_, err = layerCreator.Contribute(b.Layers, conf.Watchdog)
	if err != nil {
		b.Logger.Info(err.Error())
		os.Exit(b.Failure(cmd.LayerCreationError))
	}
}
