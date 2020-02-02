package cmd

import (
	"fmt"
	"os"

	"github.com/buildpacks/libbuildpack/v2/detect"
	"github.com/buildpacks/libbuildpack/v2/logger"
)

const (
	ParseConfigError   = detect.FailStatusCode + 1
	LayerCreationError = detect.FailStatusCode + 2
	UnexpectedError    = detect.FailStatusCode + 9
)

func Exit(code int, err error) {
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(code)
}

func ExitWithLogger(logger logger.Logger, code int, err error) {
	logger.Info(err.Error())
	os.Exit(code)
}
