package enterprise

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/influxdata/enterprise-client/v2"
	"github.com/influxdata/enterprise-client/v2/admin"
)

type Config struct {
	AdminPort uint16
	Hosts     []*client.Host
}

type Service struct {
	hosts     []*client.Host
	logger    *log.Logger
	hostname  string
	adminPort string

	shutdown chan struct{}
}

func NewEnterprise(c Config, hostname string, shutdown chan struct{}) *Service {
	return &Service{
		hosts:     c.Hosts,
		hostname:  hostname,
		logger:    log.New(os.Stdout, "[enterprise]", log.Ldate|log.Ltime),
		adminPort: fmt.Sprintf(":%d", c.AdminPort),
		shutdown:  shutdown,
	}
}

func (s *Service) Open() {
	if len(s.hosts) == 0 {
		return
	}

	cl, err := client.New(s.hosts)
	if err != nil {
		s.logger.Printf("Unable to contact one or more Enterprise hosts. err: %s", err.Error())
		return
	}
	go func() {
		token, secret, err := s.registerProduct(cl)
		if err == nil {
			s.startAdminInterface(token, secret)
		}
	}()
}

func (s *Service) registerProduct(cl *client.Client) (token string, secret string, err error) {
	p := client.Product{
		ProductID: "4815162342",
		Host:      s.hostname,
		ClusterID: "8675309",
		Name:      "telegraf",
		Version:   "0.10.1.dev",
		AdminURL:  "http://" + s.hostname + s.adminPort,
	}

	_, err = cl.Register(&p)
	if err != nil {
		s.logger.Println("Unable to register Telegraf with Enterprise")
		return
	}

	for _, host := range cl.Hosts {
		if host.Primary {
			token = host.Token
			secret = host.SecretKey
		}
	}
	return
}

func (s *Service) startAdminInterface(token, secret string) {
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      admin.App(token, []byte(secret)),
	}
	l, err := net.Listen("tcp", s.adminPort)
	if err != nil {
		s.logger.Printf("Unable to bind to admin interface port: err: %s", err.Error())
		return
	}
	go srv.Serve(l)
	select {
	case <-s.shutdown:
		s.logger.Printf("Shutting down enterprise admin interface")
		l.Close()
	}
	return
}
