package watchdog_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libbuildpack/v2/layers"
	"github.com/buildpacks/libbuildpack/v2/logger"
	"github.com/gojuno/minimock/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jromero/openfaas-cnb/pkg/config"
	"github.com/jromero/openfaas-cnb/pkg/watchdog"
)

func TestLayerCreator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LayerCreator")
}

var _ = Describe("LayerCreator", func() {

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

	Context("when version 0.0.1 used", func() {

		var (
			layersRoot string
		)

		BeforeEach(func() {
			var err error

			httpClient := watchdog.NewHttpClientMock(mc).GetMock.Return(&http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("version 0.0.1"))),
			}, nil)

			layerCreator := watchdog.NewLayerCreator(logger.Logger{}, httpClient)

			layersRoot, err = ioutil.TempDir(tmpDir, "layers")
			Expect(err).To(BeNil())

			_, err = layerCreator.Create(
				layers.Layers{Root: layersRoot},
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

				layerCreator := watchdog.NewLayerCreator(logger.Logger{}, httpClient)

				l, err := layerCreator.Create(
					layers.Layers{Root: layersRoot},
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

				layerCreator := watchdog.NewLayerCreator(logger.Logger{}, httpClient)

				l, err := layerCreator.Create(
					layers.Layers{Root: layersRoot},
					config.Watchdog{Version: "0.0.2"},
				)
				Expect(err).To(BeNil())

				b, err := ioutil.ReadFile(filepath.Join(l.Root, "watchdog"))
				Expect(err).To(BeNil())
				Expect(string(b)).To(Equal("version 0.0.2"))
			})
		})
	})
})
