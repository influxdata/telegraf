package webserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Webserver struct {
	ServiceAddress string
	Router         *mux.Router
}

func NewWebserver(serviceAddress string) *Webserver {
	return &Webserver{Router: mux.NewRouter(), ServiceAddress: serviceAddress}
}

func (wb *Webserver) Listen() {
	err := http.ListenAndServe(fmt.Sprintf("%s", wb.ServiceAddress), wb.Router)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

func (wb *Webserver) Start() error {
	go wb.Listen()
	log.Printf("Started the webhook server on %s\n", wb.ServiceAddress)
	return nil
}
