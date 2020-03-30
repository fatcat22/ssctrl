package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kardianos/service"
)

type SSService struct {
	s service.Service

	exitCh chan os.Signal
}

func NewSSService() (*SSService, error) {
	prg := &program{}
	svcConfig := &service.Config{
		Name:        "ssctrl",
		DisplayName: "ssctrl service",
		Option: service.KeyValue{
			"UserService": true,
		},
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, err
	}

	exitCh := make(chan os.Signal)
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	return &SSService{
		s: s,

		exitCh: exitCh,
	}, nil
}

func (svc *SSService) Install() error {
	return svc.s.Install()
}

func (svc *SSService) Uninstall() error {
	return svc.s.Uninstall()
}

func (svc *SSService) Exit() error {
	if service.Interactive() {
		svc.exitCh <- syscall.SIGINT
		return nil
	} else {
		return svc.s.Stop()
	}
}

func (svc *SSService) Wait() {
	<-svc.exitCh
}

type program struct{}

func (p *program) Start(s service.Service) error {
	panic("should not execute")
}
func (p *program) Stop(s service.Service) error {
	panic("should not execute")
}
