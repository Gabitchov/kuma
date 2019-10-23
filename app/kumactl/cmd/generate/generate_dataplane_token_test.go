package generate_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Kong/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	config_kumactl "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/Kong/kuma/pkg/coordinates"
	"github.com/Kong/kuma/pkg/coordinates/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

type staticDataplaneTokenGenerator struct {
	err error
}

var _ tokens.DataplaneTokenClient = &staticDataplaneTokenGenerator{}

func (s *staticDataplaneTokenGenerator) Generate(name string, mesh string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return fmt.Sprintf("token-for-%s-%s", name, mesh), nil
}

type staticCoordinatesClient struct {
	resp coordinates.Coordinates
}

var _ client.CoordinatesClient = &staticCoordinatesClient{}

func (s *staticCoordinatesClient) Coordinates() (coordinates.Coordinates, error) {
	return s.resp, nil
}

var _ = Describe("kumactl generate dataplane-token", func() {

	var rootCmd *cobra.Command
	var buf *bytes.Buffer
	var generator *staticDataplaneTokenGenerator

	BeforeEach(func() {
		generator = &staticDataplaneTokenGenerator{}
		ctx := kumactl_cmd.RootContext{
			Runtime: kumactl_cmd.RootRuntime{
				NewDataplaneTokenClient: func(string, *config_kumactl.DataplaneToken) (tokens.DataplaneTokenClient, error) {
					return generator, nil
				},
				NewCoordinatesClient: func(s string) (client.CoordinatesClient, error) {
					return &staticCoordinatesClient{
						resp: coordinates.Coordinates{},
					}, nil
				},
			},
		}

		rootCmd = cmd.NewRootCmd(&ctx)
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	It("should generate a token", func() {
		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example", "--mesh=pilot"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-pilot"))
	})

	It("should generate a token for default mesh when it is not specified", func() {
		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("token-for-example-default"))
	})

	It("should write error when generating token fails", func() {
		// setup
		generator.err = errors.New("could not connect to API")

		// when
		rootCmd.SetArgs([]string{"generate", "dataplane-token", "--dataplane=example"})
		err := rootCmd.Execute()

		// then
		Expect(err).To(HaveOccurred())

		// and
		Expect(buf.String()).To(Equal("Error: failed to generate a dataplane token: could not connect to API\n"))
	})
})
