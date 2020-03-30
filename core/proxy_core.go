package core

import (
	"github.com/fatcat22/ssctrl/config"
)

var isOnTest = false

type ProxyCore struct {
	pacSrv *PACServer
	ss     *ShadowSocks
	op     *OSOperator

	pacPort   string
	pacFile   string
	localPort string
	localAddr string
	mode      string
	srvCfg    config.ServerConfig

	isStartup bool
}

func NewProxyCore(pacPort, pacFile, localPort, mode string, srv config.ServerConfig, ssPath string) (*ProxyCore, error) {
	const localAddr = "127.0.0.1"

	pacSrv, err := NewPACServer(pacPort, pacFile, localAddr, localPort)
	if err != nil {
		return nil, err
	}
	ss, err := NewShadowSocks(ssPath, localAddr, localPort, srv)
	if err != nil {
		return nil, err
	}
	op, err := NewOSOperator(mode, pacSrv.GetPACURL(), localAddr, localPort)
	if err != nil {
		return nil, err
	}

	return &ProxyCore{
		pacSrv: pacSrv,
		ss:     ss,
		op:     op,

		pacPort:   pacPort,
		pacFile:   pacFile,
		localPort: localPort,
		localAddr: localAddr,
		mode:      mode,
		srvCfg:    srv,

		isStartup: false,
	}, nil
}

func (pc *ProxyCore) Startup() error {
	if pc.isStartup {
		return nil
	}

	if err := pc.pacSrv.Startup(); err != nil {
		return err
	}
	defer func() {
		if !pc.isStartup {
			pc.pacSrv.Shutdown()
		}
	}()

	if err := pc.ss.Startup(); err != nil {
		return err
	}
	defer func() {
		if !pc.isStartup {
			pc.ss.Shutdown()
		}
	}()

	if err := pc.op.Startup(); err != nil {
		return err
	}
	defer func() {
		if !pc.isStartup {
			pc.op.Shutdown()
		}
	}()

	pc.isStartup = true
	return nil
}

func (pc *ProxyCore) Shutdown() {
	if !pc.isStartup {
		return
	}

	pc.op.Shutdown()
	pc.ss.Shutdown()
	pc.pacSrv.Shutdown()

	pc.isStartup = false
}

func (pc *ProxyCore) ChangeMode(newMode string) error {
	if newMode == pc.mode {
		return nil
	}

	if err := pc.op.ChangeMode(newMode); err != nil {
		return err
	}

	pc.mode = newMode
	return nil
}

func (pc *ProxyCore) ChangeLocalPort(newPort string) (result error) {
	oldPort := pc.localPort
	if newPort == oldPort {
		return nil
	}

	if err := pc.ss.ChangeLocalPort(newPort); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			pc.ss.ChangeLocalPort(oldPort)
		}
	}()

	if err := pc.op.ChangeLocalPort(newPort); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			pc.op.ChangeLocalPort(oldPort)
		}
	}()

	if err := pc.pacSrv.ChangeLocalPort(newPort); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			pc.pacSrv.ChangeLocalPort(oldPort)
		}
	}()

	pc.localPort = newPort
	return nil
}

func (pc *ProxyCore) ChangePACPort(newPort string) (result error) {
	if newPort == pc.pacPort {
		return nil
	}

	newPACSrv, err := NewPACServer(newPort, pc.pacFile, pc.localAddr, pc.localPort)
	if err != nil {
		return err
	}
	if err := newPACSrv.Startup(); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			newPACSrv.Shutdown()
		}
	}()

	if err := pc.op.ChangePACURL(newPACSrv.GetPACURL()); err != nil {
		return err
	}

	pc.pacSrv.Shutdown()
	pc.pacSrv = newPACSrv
	return nil
}

func (pc *ProxyCore) ChangeServerConfig(newSrvCfg config.ServerConfig) error {
	if newSrvCfg == pc.srvCfg {
		return nil
	}

	if err := pc.ss.ChangeServerConfig(newSrvCfg); err != nil {
		return err
	}

	pc.srvCfg = newSrvCfg
	return nil
}
