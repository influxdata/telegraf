package enterprise

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
}

func NewEnterprise(c Config, hostname string) *Service {
	return &Service{
		hosts:     c.Hosts,
		hostname:  hostname,
		logger:    log.New(os.Stdout, "[enterprise]", log.Ldate|log.Ltime),
		adminPort: fmt.Sprintf(":%d", c.AdminPort),
	}
}

func (s *Service) Open() {
	cl, err := client.New(s.hosts)
	if err != nil {
		s.logger.Printf("Unable to contact one or more Enterprise hosts. err: %s", err.Error())
		return
	}
	go s.registerProduct(cl)
	go s.startAdminInterface()
}

func (s *Service) registerProduct(cl *client.Client) {
	p := client.Product{
		ProductID: "telegraf",
		Host:      s.hostname,
	}

	_, err := cl.Register(&p)
	if err != nil {
		s.logger.Println("Unable to register Telegraf with Enterprise")
	}
}

func (s *Service) startAdminInterface() {
	go http.ListenAndServe(s.adminPort, admin.App("foo", []byte("bar")))
}
