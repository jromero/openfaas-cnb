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

	conf := watchdog.DefaultConfig()
	configPath := config.Path(b.Application.Root)
	if fh, err := os.Open(configPath); err != nil {
		if !os.IsNotExist(err) {
			cmd.ExitWithLogger(b.Logger, cmd.UnexpectedError, err)
		}
	} else {
		defer fh.Close()
		conf, err = watchdog.ParseConfig(fh)
		if err != nil {
			cmd.ExitWithLogger(b.Logger, cmd.ParseConfigError, err)
		}
	}

	contributor := watchdog.NewContributor(b.Logger, http.DefaultClient)
	_, err = contributor.Contribute(b.Layers, conf.Watchdog)
	if err != nil {
		b.Logger.Info(err.Error())
		os.Exit(b.Failure(cmd.LayerCreationError))
	}
}
