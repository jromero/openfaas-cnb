package main

import (
	"os"

	"github.com/buildpacks/libbuildpack/v2/detect"
)

func main() {
	os.Exit(detect.PassStatusCode)
}
