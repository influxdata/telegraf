package http_listener

import (
	"compress/gzip"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/services"
	"github.com/influxdata/telegraf/pubsub"
)

type TimeFunc func() time.Time

type HttpReceiver struct {
	ServiceAddress    string
	Path              string
	Methods           []string
	ReadTimeout       internal.Duration
	WriteTimeout      internal.Duration
	MaxBodySize       internal.Size
	Port              int
	TLSCert           string   `toml:"tls_cert"`
	TLSKey            string   `toml:"tls_key"`
	TLSAllowedCACerts []string `toml:"tls_allowed_cacerts"`

	BasicUsername string
	BasicPassword string

	TimeFunc

	wg sync.WaitGroup

	listener net.Listener
	ps       *pubsub.PubSub
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf(
				"could not read certificate %q: %v", certFile, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse any PEM certificates %q: %v", certFile, err)
		}
	}
	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf(
			"could not load keypair %s:%s: %v", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	config.BuildNameToCertificate()
	return nil
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (h *HttpReceiver) TLSConfig() (*tls.Config, error) {
	if h.TLSCert == "" && h.TLSKey == "" && len(h.TLSAllowedCACerts) == 0 {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if len(h.TLSAllowedCACerts) != 0 {
		pool, err := makeCertPool(h.TLSAllowedCACerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if h.TLSCert != "" && h.TLSKey != "" {
		err := loadCertificate(tlsConfig, h.TLSCert, h.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

func (h *HttpReceiver) Description() string {
	return "Receive http web requests"
}

var sampleConfig = `
`

func (h *HttpReceiver) SampleConfig() string {
	return sampleConfig
}

func (h *HttpReceiver) Connect() error {
	fmt.Println("Http Connect")

	return nil
}

func (h *HttpReceiver) Run(msgbus *pubsub.PubSub) error {
	tlsConf, err := h.TLSConfig()
	if err != nil {
		log.Println("E! TLSConfig", err)
		return err
	}

	server := &http.Server{
		Addr:         h.ServiceAddress,
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", h.ServiceAddress, tlsConf)
	} else {
		listener, err = net.Listen("tcp", h.ServiceAddress)
	}
	if err != nil {
		log.Println("E! Creating http listener", err)
		return err
	}
	fmt.Println("Http Listen")
	h.ps = msgbus
	h.listener = listener
	h.Port = listener.Addr().(*net.TCPAddr).Port

	log.Printf("I! Starting HTTP listener service on %s\n", h.ServiceAddress)
	server.Serve(h.listener)
	log.Printf("I! Started HTTP listener service on %s\n", h.ServiceAddress)
	return nil
}

func (h *HttpReceiver) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.ps.Pub(req)

	if req.URL.Path == h.Path {
		h.AuthenticateIfSet(h.serveWrite, res, req)
	} else {
		h.AuthenticateIfSet(http.NotFound, res, req)
	}
}

func (h *HttpReceiver) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize.Size {
		tooLarge(res)
		return
	}

	// Check if the requested HTTP method was specified in config.
	isAcceptedMethod := false
	for _, method := range h.Methods {
		if req.Method == method {
			isAcceptedMethod = true
			break
		}
	}
	if !isAcceptedMethod {
		methodNotAllowed(res)
		return
	}

	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		if err != nil {
			log.Println("D! " + err.Error())
			badRequest(res)
			return
		}
		defer body.Close()
	}

	body = http.MaxBytesReader(res, body, h.MaxBodySize.Size)
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		tooLarge(res)
		return
	}

	fmt.Println(string(bytes))
	res.WriteHeader(http.StatusNoContent)
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func methodNotAllowed(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusMethodNotAllowed)
	res.Write([]byte(`{"error":"http: method not allowed"}`))
}

func internalServerError(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusInternalServerError)
}

func badRequest(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(`{"error":"http: bad request"}`))
}

func (h *HttpReceiver) AuthenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if h.BasicUsername != "" && h.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {

			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(res, req)
	} else {
		handler(res, req)
	}
}

func (h *HttpReceiver) Close() error {
	fmt.Println("Http Close")
	return nil
}

func init() {
	services.Add("http", func() telegraf.Service {
		return &HttpReceiver{}
	})
}
