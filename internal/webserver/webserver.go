package webserver

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Webserver struct {
	ServiceAddress string
	Router         *mux.Router
	onceStart      sync.Once
}

func NewWebserver(serviceAddress string) *Webserver {
	return &Webserver{Router: mux.NewRouter(), ServiceAddress: serviceAddress}
}

func (wb *Webserver) listen() {
	log.Printf("Started the webhook server on %s\n", wb.ServiceAddress)
	err := http.ListenAndServe(fmt.Sprintf("%s", wb.ServiceAddress), wb.Router)
	if err != nil {
		log.Printf("Error starting webhook server: %v", err)
	}
}

func (wb *Webserver) StartOnce() {
	wb.onceStart.Do(func() {
		go wb.listen()
	})
}
