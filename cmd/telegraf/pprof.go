package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server interface {
	Start(string)
	ErrChan() <-chan error
}

type PprofServer struct {
	err chan error
}

func NewPprofServer() *PprofServer {
	return &PprofServer{
		err: make(chan error),
	}
}

func (p *PprofServer) Start(address string) {
	go func() {
		pprofHostPort := address
		parts := strings.Split(pprofHostPort, ":")
		if len(parts) == 2 && parts[0] == "" {
			pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
		}
		pprofHostPort = "http://" + pprofHostPort + "/debug/pprof"

		log.Printf("I! Starting pprof HTTP server at: %s", pprofHostPort)

		server := &http.Server{
			Addr:         address,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil {
			p.err <- fmt.Errorf("E! %w", err)
		}
		close(p.err)
	}()
}

func (p *PprofServer) ErrChan() <-chan error {
	return p.err
}
