package main

import (
	"os"

	"github.com/buildpacks/libbuildpack/v2/detect"

	"github.com/jromero/openfaas-cnb/cmd"
	"github.com/jromero/openfaas-cnb/pkg/config"
)

func main() {
	d, err := detect.DefaultDetect()
	if err != nil {
		os.Exit(detect.FailStatusCode)
	}

	if exists, err := config.Exists(config.Filename(d.Application.Root)); err != nil {
		d.Logger.Info(err.Error())
		os.Exit(d.Error(cmd.UnexpectedError))
	} else if !exists {
		os.Exit(d.Fail())
	}

	exitCode, err := d.Pass()
	if err != nil {
		d.Logger.Info(err.Error())
		os.Exit(d.Error(cmd.UnexpectedError))
	}

	os.Exit(exitCode)
}
