package test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/kentik/libkflow/api"
	"github.com/kentik/libkflow/chf"
	"github.com/robfig/cron"
	"zombiezen.com/go/capnproto2"
)

type Server struct {
	Host     net.IP
	Port     int
	Email    string
	Token    string
	Device   *api.Device
	Log      *log.Logger
	quiet    bool
	flows    chan chf.PackedCHF
	mux      *mux.Router
	listener net.Listener
}

var (
	flowCounter   uint64
	packetCounter uint64
	byteCounter   uint64
)

const (
	API  = "/api/internal"
	FLOW = "/chf"
	TSDB = "/tsdb"
)

func NewServer(host string, port int, tls, quiet bool) (*Server, error) {
	var listener net.Listener

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}

	listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	if tls {
		listener, err = tlslistener(listener, host, addr)
		if err != nil {
			return nil, err
		}
	}

	addr = listener.Addr().(*net.TCPAddr)

	return &Server{
		Host:     addr.IP,
		Port:     addr.Port,
		Log:      log.New(os.Stderr, "", log.LstdFlags),
		quiet:    quiet,
		flows:    make(chan chf.PackedCHF, 100),
		mux:      mux.NewRouter(),
		listener: listener,
	}, nil
}

func (s *Server) Serve(email, token string, dev *api.Device) error {
	s.Email = email
	s.Token = token
	s.Device = dev

	s.mux.HandleFunc(API+"/device/{did}", s.wrap(s.device))
	s.mux.HandleFunc(API+"/device/{did}/interfaces", s.wrap(s.interfaces))
	s.mux.HandleFunc(API+"/company/{cid}/device/{did}/tags/snmp", s.wrap(s.update))
	s.mux.HandleFunc(FLOW, s.wrap(s.flow))
	s.mux.HandleFunc(TSDB, s.wrap(s.tsdb))

	c := cron.New()
	c.AddFunc("* * * * * *", func() {
		flows := atomic.SwapUint64(&flowCounter, 0)
		packets := atomic.SwapUint64(&packetCounter, 0)
		bytes := atomic.SwapUint64(&byteCounter, 0)
		if flows > 0 || packets > 0 || bytes > 0 {
			s.Log.Printf("flows: %12d, packets: %12d, bytes: %12d", flows, packets, bytes)
		}
	})
	c.Start()

	return http.Serve(s.listener, s.mux)
}

func (s *Server) URL(path string) *url.URL {
	url, _ := url.Parse(fmt.Sprintf("http://%s:%d%s", s.Host, s.Port, path))
	return url
}

func (s *Server) Flows() <-chan chf.PackedCHF {
	return s.flows
}

func (s *Server) device(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["did"]

	switch {
	case id == strconv.Itoa(s.Device.ID):
	case id == s.Device.Name:
	case net.ParseIP(id).Equal(s.Device.IP):
	default:
		panic(http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(&api.DeviceResponse{
		Device: s.Device,
	})

	if err != nil {
		panic(http.StatusInternalServerError)
	}
}

func (s *Server) interfaces(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]api.Interface{})
}

func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	// just ignore it
}

func (s *Server) flow(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("sid") != "0" {
		panic(http.StatusBadRequest)
	}

	if r.FormValue("sender_id") != s.Device.ClientID() {
		panic(http.StatusBadRequest)
	}

	if r.Header.Get("Content-Type") != "application/binary" {
		panic(http.StatusBadRequest)
	}

	cid := [80]byte{}
	n, err := r.Body.Read(cid[:])
	if err != nil || n != len(cid) {
		panic(http.StatusBadRequest)
	}

	msg, err := capnp.NewPackedDecoder(r.Body).Decode()
	defer r.Body.Close()
	if err != nil {
		panic(http.StatusBadRequest)
	}

	root, err := chf.ReadRootPackedCHF(msg)
	if err != nil {
		panic(http.StatusBadRequest)
	}

	select {
	case s.flows <- root:
	default:
	}

	msgs, err := root.Msgs()
	if err != nil {
		panic(http.StatusBadRequest)
	}

	var (
		packetctr uint64
		bytectr   uint64
	)

	for i := 0; i < msgs.Len(); i++ {
		msg := msgs.At(i)

		packetctr += msg.InPkts() + msg.OutPkts()
		bytectr += msg.InBytes() + msg.OutBytes()

		if !s.quiet {
			buf := bytes.Buffer{}
			Print(&buf, i, msg, s.Device)
			s.Log.Output(0, buf.String())
		}
	}

	atomic.AddUint64(&flowCounter, uint64(msgs.Len()))
	atomic.AddUint64(&packetCounter, packetctr)
	atomic.AddUint64(&byteCounter, bytectr)
}

func (s *Server) tsdb(w http.ResponseWriter, r *http.Request) {
	// just ignore it
	io.Copy(ioutil.Discard, r.Body)
}

func (s *Server) wrap(f handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				if code, ok := r.(int); ok {
					http.Error(w, http.StatusText(code), code)
					return
				}
				panic(r)
			}
		}()

		email := r.Header.Get("X-CH-Auth-Email")
		token := r.Header.Get("X-CH-Auth-API-Token")

		if email != s.Email || token != s.Token {
			panic(http.StatusUnauthorized)
		}

		if err := r.ParseForm(); err != nil {
			panic(http.StatusBadRequest)
		}

		f(w, r)
	}
}

func tlslistener(tcp net.Listener, host string, addr *net.TCPAddr) (net.Listener, error) {
	pri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	pub := &pri.PublicKey

	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{Organization: []string{"Kentik"}},
		IPAddresses:           []net.IP{addr.IP},
		DNSNames:              []string{host},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, pri)
	if err != nil {
		return nil, err
	}

	cert := tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  pri,
	}

	cfg := tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.NewListener(tcp, &cfg), nil
}

type handler func(http.ResponseWriter, *http.Request)
