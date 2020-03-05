// +build e2e

package test_e2e

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	// Env var to provide a specific pack binary
	packBinary = "PACK_BIN"
	// Env var to provide a specific buildpack path
	buildpackPath = "BUILDPACK_PATH"
)

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E")
}

var _ = Describe("E2E", func() {
	t := GinkgoT()

	It("runs successfully", func() {
		packExec := findExec(t, "pack", os.Getenv(packBinary))

		t.Log("> Test with heroku/buildpacks:18")

		t.Log("Building buildpack...")
		buildpackPath := resolveBuildpack(t)

		t.Log("Building app...")
		imageName := "test-app-" + strconv.Itoa(rand.Int())
		packCmd := exec.Command(
			packExec,
			"build", imageName,
			"--builder", "heroku/buildpacks:18",
			"--buildpack", "from=builder",
			"--buildpack", buildpackPath,
			"--path", filepath.Join("testdata", "app"),
			"--verbose",
		)

		t.Log("Running: ", packCmd.String())

		output, err := packCmd.CombinedOutput()
		Expect(err).To(BeNil(), "failed build app.\n %s", string(output))

		t.Log("Run application...")
		dockerExec := findExec(t, "docker", "")

		hostPort, err := getFreePort()
		Expect(err).To(BeNil(), "failed to get free port")

		dockerCmd := exec.Command(dockerExec, "run", "--rm", "-d", "-p", strconv.Itoa(hostPort)+":8080", imageName, "faas")
		output, err = dockerCmd.CombinedOutput()
		Expect(err).To(BeNil(), "failed to start container")

		containerId := string(output)
		defer func() {
			exec.Command(dockerExec, "stop", containerId)
		}()

		// TODO: Do this better
		time.Sleep(3 * time.Second)

		t.Log("Ensure it's running as expected...")
		Eventually(func() (string, error) {
			resp, err := http.Get("http://localhost:" + strconv.Itoa(hostPort))
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			buf := &bytes.Buffer{}
			if _, err := io.Copy(buf, resp.Body); err != nil {
				return "", fmt.Errorf("failed to read response: %s", err.Error())
			}

			return strings.TrimSpace(buf.String()), nil
		}).Should(Equal("it works"))
	})
})

func resolveBuildpack(t Tester) string {
	if envBuildpackPath := os.Getenv(buildpackPath); envBuildpackPath != "" {
		t.Log("Prebuilt buildpack path provided, NOT building...")
		if !filepath.IsAbs(envBuildpackPath) {
			envBuildpackPath = filepath.Join(projectDir(t), envBuildpackPath)
		}

		return envBuildpackPath
	}

	t.Log("No prebuilt buildpack path provided, building...")

	makeExec := findExec(t, "make", "")

	cmd := exec.Command(makeExec, "build")
	cmd.Dir = projectDir(t)
	cmd.Env = append(os.Environ(), "GOOS=linux", "CGO_ENABLED=0", "GOARCH=amd64")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Log("output: ", string(output))
		t.Fatalf("failed build buildpack: %s", err)
	}

	return filepath.Join(cmd.Dir, "build/")
}

func projectDir(t Tester) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %s", err)
	}

	if filepath.Base(wd) == "test_e2e" {
		return filepath.Dir(wd)
	}

	return wd
}

func findExec(t Tester, command, override string) string {
	if override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			t.Fatalf("could not resolve command override '%s': %s", override, err)
		}

		return abs
	}

	output, err := exec.Command("which", command).CombinedOutput()
	if err != nil {
		t.Log("output: ", string(output))
		t.Fatalf("could not find command '%s': %s", command, err)
	}

	return command
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type Tester interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Error(...interface{})
	Errorf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	FailNow()
}
