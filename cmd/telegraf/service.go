// Run the telegraf data collector
package main

import (
	"log"

	"github.com/kardianos/service"
)

var logger service.Logger

var stop chan struct{}

var srvc service.Service

type program struct{}

func (p *program) Start(s service.Service) error {
	srvc = s
	go p.run()
	return nil
}
func (p *program) run() {
	stop = make(chan struct{})
	reloadLoop(stop, srvc)
}
func (p *program) Stop(s service.Service) error {
	close(stop)
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "telegraf",
		DisplayName: "Telegraf Data Collector Service",
		Description: "Collects data using a series of plugins and publishes it to" +
			"another series of plugins.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
