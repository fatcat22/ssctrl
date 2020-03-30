package main

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/fatcat22/ssctrl/config"
)

type CoreInterface interface {
	Startup() error
	Shutdown()
	ChangeMode(newMode string) error
	ChangeLocalPort(newPort string) error
	ChangePACPort(newPort string) error
	ChangeServerConfig(newSrvCfg config.ServerConfig) error
}

type Controler struct {
	cfg *config.AppConfig

	apiSrv *apiServer
	core   CoreInterface
	svc    *SSService

	isRunning bool
	lock      sync.Mutex

	exitCh chan<- os.Signal
}

func NewControler(cfg *config.AppConfig, core CoreInterface, svc *SSService) (*Controler, error) {
	ctrl := &Controler{
		cfg: cfg,

		core: core,
		svc:  svc,

		isRunning: false,
	}

	apiSrv, err := NewAPIServer(cfg.GetAPIPort(), ctrl)
	if err != nil {
		return nil, err
	}

	ctrl.apiSrv = apiSrv
	return ctrl, nil
}

func (ctrl *Controler) Startup() error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	return ctrl.startup()
}

func (ctrl *Controler) Shutdown() {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	ctrl.shutdown()
}

func (ctrl *Controler) EnableProxy() error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	if !ctrl.isRunning {
		return errors.New("ssctrl not running")
	}

	if ctrl.cfg.IsEnabled() {
		return nil
	}

	if err := ctrl.core.Startup(); err != nil {
		return err
	}

	ctrl.cfg.SetEnabled(true)
	return nil
}

func (ctrl *Controler) DisableProxy() error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	if !ctrl.isRunning {
		return errors.New("ssctrl not running")
	}

	if !ctrl.cfg.IsEnabled() {
		return nil
	}

	ctrl.core.Shutdown()

	ctrl.cfg.SetEnabled(false)
	return nil
}

func (ctrl *Controler) ChangeMode(newMode string) error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	if !ctrl.isRunning {
		return errors.New("ssctrl not running")
	}
	if ctrl.cfg.GetMode() == newMode {
		return nil
	}
	if !config.IsValidMode(newMode) {
		return fmt.Errorf("unknown mode name '%s'", newMode)
	}

	if err := ctrl.core.ChangeMode(newMode); err != nil {
		return err
	}

	ctrl.cfg.SetModeMust(newMode)
	return nil
}

func (ctrl *Controler) ChangeLocalPort(newPort string) error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	if newPort == ctrl.cfg.GetLocalPort() {
		return nil
	}
	if err := ctrl.cfg.CheckLocalPort(newPort); err != nil {
		return err
	}

	if err := ctrl.core.ChangeLocalPort(newPort); err != nil {
		return err
	}

	ctrl.cfg.SetLocalPortMust(newPort)
	return nil
}

func (ctrl *Controler) ChangePACPort(newPort string) error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	if newPort == ctrl.cfg.GetPACPort() {
		return nil
	}
	if err := ctrl.cfg.CheckPACPort(newPort); err != nil {
		return err
	}

	if err := ctrl.core.ChangePACPort(newPort); err != nil {
		return err
	}

	ctrl.cfg.SetPACPortMust(newPort)
	return nil
}

func (ctrl *Controler) ChangeAPIPort(newPort string) error {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	if newPort == ctrl.cfg.GetAPIPort() {
		return nil
	}
	if err := ctrl.cfg.CheckAPIPort(newPort); err != nil {
		return err
	}

	newSrv, err := NewAPIServer(newPort, ctrl)
	if err != nil {
		return err
	}
	if err := newSrv.Startup(); err != nil {
		return err
	}

	oldSrv := ctrl.apiSrv
	ctrl.apiSrv = newSrv
	go oldSrv.Shutdown()
	ctrl.cfg.SetAPIPortMust(newPort)
	return nil
}

func (ctrl *Controler) ChangeCurrentServer(newSrvName string) error {
	srvCfg, err := ctrl.cfg.GetServerConfig(newSrvName)
	if err != nil {
		return err
	}

	if err := ctrl.core.ChangeServerConfig(srvCfg); err != nil {
		return err
	}

	ctrl.cfg.SetCurrentServerMust(newSrvName)
	return nil
}

func (ctrl *Controler) UpdateServers(servers map[string]config.ServerConfig) (result error) {
	if err := ctrl.checkServersConfig(servers); err != nil {
		return err
	}

	currentSrvName, _ := ctrl.cfg.GetCurrentServerConfig()
	if newCurrentSrv, ok := servers[currentSrvName]; ok {
		config.SetServerConfigDefault(&newCurrentSrv)
		if err := ctrl.core.ChangeServerConfig(newCurrentSrv); err != nil {
			return err
		}
	}

	for name, srv := range servers {
		ctrl.cfg.UpdateServerMust(name, srv)
	}
	return nil
}

func (ctrl *Controler) RemoveServers(names []string) error {
	for _, name := range names {
		if err := ctrl.cfg.CheckServerBeRemoved(name); err != nil {
			return err
		}
	}

	for _, name := range names {
		ctrl.cfg.RemoveServerMust(name)
	}
	return nil
}

func (ctrl *Controler) Autorun(enable bool) error {
	if enable {
		if err := ctrl.svc.Install(); err != nil {
			return err
		}
	} else {
		if err := ctrl.svc.Uninstall(); err != nil {
			return err
		}
	}

	ctrl.cfg.SetAutorun(enable)
	return nil
}

func (ctrl *Controler) Exit() {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	go ctrl.svc.Exit()
}

func (ctrl *Controler) MarshalConfig(marshalFunc func(v interface{}) ([]byte, error)) ([]byte, error) {
	ctrl.lock.Lock()
	defer ctrl.lock.Unlock()

	return ctrl.cfg.Marshal(marshalFunc)
}

func (ctrl *Controler) startup() error {
	if ctrl.isRunning {
		return nil
	}

	if ctrl.cfg.IsEnabled() {
		if err := ctrl.core.Startup(); err != nil {
			return err
		}
		defer func() {
			if !ctrl.isRunning {
				ctrl.core.Shutdown()
			}
		}()
	}

	if err := ctrl.apiSrv.Startup(); err != nil {
		return err
	}
	defer func() {
		if !ctrl.isRunning {
			ctrl.apiSrv.Shutdown()
		}
	}()

	ctrl.isRunning = true
	return nil
}

func (ctrl *Controler) shutdown() {
	if !ctrl.isRunning {
		return
	}

	ctrl.apiSrv.Shutdown()
	ctrl.core.Shutdown()

	ctrl.isRunning = false
}

func (ctrl *Controler) checkServersConfig(servers map[string]config.ServerConfig) error {
	for name, srv := range servers {
		if len(name) == 0 {
			return errors.New("server name is empty")
		}

		if err := ctrl.cfg.CheckServerConfig(srv); err != nil {
			return fmt.Errorf("invalid config for '%s': %v", name, err)
		}
	}

	return nil
}
