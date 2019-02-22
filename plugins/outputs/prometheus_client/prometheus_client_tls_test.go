package prometheus_client_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/influxdata/telegraf/plugins/outputs/prometheus_client"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"testing"
)

var ca, _ = filepath.Abs("assets/telegrafCA.crt")
var cert, _ = filepath.Abs("assets/telegraf.crt")
var key, _ = filepath.Abs("assets/telegraf.key")
var configWithTLS = fmt.Sprintf(`
 listen = "127.0.0.1:9090"
 tls_ca = "%s"
 tls_cert = "%s"
 tls_key = "%s"
`, ca, cert, key)

var configWithoutTLS = `
  listen = "127.0.0.1:9090"
`

type PrometheusClientTestContext struct {
	Output      *prometheus_client.PrometheusClient
	Accumulator *testutil.Accumulator
	Client      *http.Client

	*GomegaWithT
}

func init() {
	path, _ := filepath.Abs("./scripts/generate_certs.sh")
	_, err := exec.Command(path).CombinedOutput()
	if err != nil {
		panic(err)
	}
}

func TestWorksWithoutTLS(t *testing.T) {
	tc := buildTestContext(t, []byte(configWithoutTLS))
	err := tc.Output.Connect()
	defer tc.Output.Close()

	if err != nil {
		panic(err)
	}

	var response *http.Response
	tc.Eventually(func() bool {
		response, err = tc.Client.Get("http://localhost:9090/metrics")
		return err == nil
	}, "5s").Should(BeTrue())

	if err != nil {
		panic(err)
	}

	tc.Expect(response.StatusCode).To(Equal(http.StatusOK))
}

func TestWorksWithTLS(t *testing.T) {
	tc := buildTestContext(t, []byte(configWithTLS))
	err := tc.Output.Connect()
	defer tc.Output.Close()

	if err != nil {
		panic(err)
	}

	var response *http.Response
	tc.Eventually(func() bool {
		response, err = tc.Client.Get("https://localhost:9090/metrics")
		return err == nil
	}, "5s").Should(BeTrue())

	if err != nil {
		panic(err)
	}

	tc.Expect(response.StatusCode).To(Equal(http.StatusOK))

	response, err = tc.Client.Get("http://localhost:9090/metrics")

	tc.Expect(err).To(HaveOccurred())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	response, err = client.Get("https://localhost:9090/metrics")

	tc.Expect(err).To(HaveOccurred())
}

func buildTestContext(t *testing.T, config []byte) *PrometheusClientTestContext {
	output := prometheus_client.NewClient()
	err := toml.Unmarshal(config, output)

	if err != nil {
		panic(err)
	}

	var (
		httpClient *http.Client
	)

	if output.TLSCA != "" {
		httpClient = buildClientWithTLS(output)
	} else {
		httpClient = buildClientWithoutTLS()
	}

	return &PrometheusClientTestContext{
		Output:      output,
		Accumulator: &testutil.Accumulator{},
		Client:      httpClient,
		GomegaWithT: NewGomegaWithT(t),
	}
}

func buildClientWithoutTLS() *http.Client {
	return &http.Client{}
}

func buildClientWithTLS(output *prometheus_client.PrometheusClient) *http.Client {
	cert, err := tls.LoadX509KeyPair(output.TLSCert, output.TLSKey)
	if err != nil {
		panic(err)
	}

	caCert, err := ioutil.ReadFile(output.TLSCA)
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		ServerName:   "telegraf",
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}
}
