package api_server_test

import (
	api_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"

	konvoy_version "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/version"
)

var _ = Describe("Index WS", func() {

	var backupBuildInfo konvoy_version.BuildInfo
	BeforeEach(func() {
		backupBuildInfo = konvoy_version.Build
	})
	AfterEach(func() {
		konvoy_version.Build = backupBuildInfo
	})

	It("should return the version of Kuma Control Plane", func(done Done) {
		// given
		konvoy_version.Build = konvoy_version.BuildInfo{
			Version:   "1.2.3",
			GitTag:    "v1.2.3",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}

		// setup
		resourceStore := memory.NewStore()
		apiServer := createTestApiServer(resourceStore, *api_server.DefaultApiServerConfig())

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for the server
		Eventually(func() error {
			_, err := http.Get("http://localhost" + apiServer.Address())
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get("http://localhost" + apiServer.Address())
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		expected := `
		{
			"tagline": "Kuma",
			"version": "1.2.3"
		}
`
		Expect(body).To(MatchJSON(expected))
		close(done)
	}, 5)
})
