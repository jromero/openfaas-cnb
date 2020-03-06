package watchdog_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libbuildpack/v2/layers"
	"github.com/buildpacks/libbuildpack/v2/logger"
	"github.com/gojuno/minimock/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

	Describe("ParseConfig", func() {
		It("parses a config file", func() {
			conf, err := watchdog.ParseConfig(strings.NewReader(`
[watchdog]
version = "1.2.3"
process_type = "someType"

[watchdog.env]
key1 = "value1"
`))
			Expect(err).To(BeNil())
			Expect(conf).To(Equal(watchdog.Config{
				Version:     "1.2.3",
				ProcessType: "someType",
			}))
		})

		Context("version is not set", func() {
			It("defaults to '0.7.6'", func() {
				conf, err := watchdog.ParseConfig(strings.NewReader(``))
				Expect(err).To(BeNil())
				Expect(conf.Version).To(Equal("0.7.6"))
			})
		})

		Context("process_type is not set", func() {
			It("defaults to 'web'", func() {
				conf, err := watchdog.ParseConfig(strings.NewReader(``))
				Expect(err).To(BeNil())
				Expect(conf.ProcessType).To(Equal("web"))
			})
		})
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
					watchdog.Config{Version: "0.0.1"},
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
						watchdog.Config{Version: "0.0.1"},
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
						watchdog.Config{Version: "0.0.2"},
					)
					Expect(err).To(BeNil())

					b, err := ioutil.ReadFile(filepath.Join(l.Root, "watchdog"))
					Expect(err).To(BeNil())
					Expect(string(b)).To(Equal("version 0.0.2"))
				})
			})
		})

		Context("when 'process_type' is set to 'blah'", func() {
			It("should set function_process to 'web' process type and create 'faas' process type", func() {
				httpClient := watchdog.NewHttpClientMock(mc).GetMock.Return(&http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("version 0.0.2"))),
				}, nil)
				layerCreator := watchdog.NewContributor(logger.Logger{}, httpClient)

				watchdogLayer, err := layerCreator.Contribute(lyrs, watchdog.Config{
					Version:     "0.0.1",
					ProcessType: "blah",
				})
				Expect(err).To(BeNil())

				md := &layers.Metadata{}
				_, err = toml.DecodeFile(filepath.Join(lyrs.Root, "launch.toml"), md)
				Expect(err).To(BeNil())

				b, err := ioutil.ReadFile(filepath.Join(watchdogLayer.Root, "env.launch", "function_process.default"))
				Expect(err).To(BeNil())
				Expect(string(b)).To(Equal("/cnb/lifecycle/launcher blah"))

				Expect(md.Processes[0].Type).To(Equal("faas"))
				Expect(md.Processes[0].Command).To(Equal(filepath.Join(watchdogLayer.Root, "watchdog")))
			})
		})

		Context("when version is not found", func() {
			It("should fail", func() {
				httpClient := watchdog.NewHttpClientMock(mc).GetMock.Return(&http.Response{
					StatusCode: 404,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("not found"))),
				}, nil)
				layerCreator := watchdog.NewContributor(logger.Logger{}, httpClient)

				_, err := layerCreator.Contribute(lyrs, watchdog.Config{
					Version:     "0.0.1",
					ProcessType: "blah",
				})

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("downloading from"))
				Expect(err.Error()).To(ContainSubstring("returned status code '404'"))
			})
		})
	})
})
