package watchdog_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libbuildpack/v2/layers"
	"github.com/buildpacks/libbuildpack/v2/logger"
	"github.com/gojuno/minimock/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jromero/openfaas-cnb/pkg/config"
	"github.com/jromero/openfaas-cnb/pkg/watchdog"
)

func TestWatchdog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Watchdog")
}

var _ = Describe("Watchdog", func() {
	var (
		tmpDir string
		mc     minimock.MockController
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).To(BeNil())

		mc = minimock.NewController(GinkgoT())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(BeNil())
	})

	Describe("Contributor", func() {
		var (
			lyrs layers.Layers
		)

		BeforeEach(func() {
			layersRoot, err := ioutil.TempDir(tmpDir, "layers")
			Expect(err).To(BeNil())

			lyrs = layers.NewLayers(layersRoot, logger.Logger{})
		})

		Context("when version 0.0.1 used", func() {

			BeforeEach(func() {
				httpClient := watchdog.NewHttpClientMock(mc).GetMock.Return(&http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("version 0.0.1"))),
				}, nil)
				layerCreator := watchdog.NewContributor(logger.Logger{}, httpClient)
				_, err := layerCreator.Contribute(
					lyrs,
					config.Watchdog{Version: "0.0.1"},
				)
				Expect(err).To(BeNil())
			})

			Context("and version 0.0.1 is used again", func() {
				It("doesn't download again", func() {
					httpClient := watchdog.NewHttpClientMock(mc).GetMock.Set(func(url string) (_ *http.Response, _ error) {
						Fail("tried to download: " + url)
						return nil, nil
					})

					layerCreator := watchdog.NewContributor(logger.Logger{}, httpClient)

					l, err := layerCreator.Contribute(
						lyrs,
						config.Watchdog{Version: "0.0.1"},
					)
					Expect(err).To(BeNil())

					b, err := ioutil.ReadFile(filepath.Join(l.Root, "watchdog"))
					Expect(err).To(BeNil())
					Expect(string(b)).To(Equal("version 0.0.1"))
				})
			})

			Context("and version 0.0.2 is used", func() {
				It("downloads new version", func() {
					httpClient := watchdog.NewHttpClientMock(mc).GetMock.Return(&http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("version 0.0.2"))),
					}, nil)
					layerCreator := watchdog.NewContributor(logger.Logger{}, httpClient)

					l, err := layerCreator.Contribute(
						lyrs,
						config.Watchdog{Version: "0.0.2"},
					)
					Expect(err).To(BeNil())

					b, err := ioutil.ReadFile(filepath.Join(l.Root, "watchdog"))
					Expect(err).To(BeNil())
					Expect(string(b)).To(Equal("version 0.0.2"))
				})
			})
		})

		Context("another buildpack provides 'web' process 'ruby app.rb'", func() {
			BeforeEach(func() {
				err := lyrs.WriteApplicationMetadata(layers.Metadata{
					Processes: []layers.Process{{
						Type:    "web",
						Command: "ruby app.rb",
					}},
				})
				Expect(err).To(BeNil())
			})

			It("should set function_process to 'web' process and create 'faas' process", func() {
				httpClient := watchdog.NewHttpClientMock(mc).GetMock.Return(&http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("version 0.0.2"))),
				}, nil)
				layerCreator := watchdog.NewContributor(logger.Logger{}, httpClient)

				watchdogLayer, err := layerCreator.Contribute(lyrs, config.Watchdog{
					Version: "0.0.1",
					Env:     nil,
				})
				Expect(err).To(BeNil())

				md := &layers.Metadata{}
				_, err = toml.DecodeFile(filepath.Join(lyrs.Root, "launch.toml"), md)
				Expect(err).To(BeNil())

				b, err := ioutil.ReadFile(filepath.Join(watchdogLayer.Root, "env.launch", "function_process.default"))
				Expect(err).To(BeNil())
				Expect(string(b)).To(Equal("/cnb/lifecycle/launcher web"))

				Expect(md.Processes[0].Type).To(Equal("faas"))
				Expect(md.Processes[0].Command).To(Equal(filepath.Join(watchdogLayer.Root, "watchdog")))
			})
		})
	})
})
