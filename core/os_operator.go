package core

import (
	"errors"
	"fmt"

	"github.com/fatcat22/ssctrl/config"
)

const bypass = "127.0.0.1、192.168.0.0/16、10.0.0.0/8、localhost"

type opConfig struct {
	mode      string
	pacURL    string
	localAddr string
	localPort string
}

type OSOperator struct {
	cfg opConfig

	isStartup bool
}

func NewOSOperator(mode, pacURL, addr, port string) (*OSOperator, error) {
	if mode != config.ModePAC && mode != config.ModeGlobal {
		return nil, fmt.Errorf("unknown mode '%s'", mode)
	}

	return &OSOperator{
		cfg: opConfig{
			mode:      mode,
			pacURL:    pacURL,
			localAddr: addr,
			localPort: port,
		},

		isStartup: false,
	}, nil
}

func (op *OSOperator) Startup() error {
	if op.isStartup {
		return nil
	}

	if err := resetProxy(op.cfg); err != nil {
		return err
	}

	op.isStartup = true
	return nil
}

func (op *OSOperator) Shutdown() error {
	if !op.isStartup {
		return nil
	}

	if err := clearProxyByMode(op.cfg.mode); err != nil {
		return err
	}

	op.isStartup = false
	return nil
}

func (op *OSOperator) ChangeMode(newMode string) (result error) {
	if newMode == op.cfg.mode {
		return nil
	}

	if !op.isStartup {
		op.cfg.mode = newMode
		return nil
	}

	newCfg := op.cfg
	newCfg.mode = newMode
	if err := op.reStartup(newCfg); err != nil {
		return err
	}

	op.cfg.mode = newMode
	return nil
}

func (op *OSOperator) ChangeLocalPort(newPort string) error {
	if newPort == op.cfg.localPort {
		return nil
	}

	if !op.isStartup || op.cfg.mode == config.ModePAC {
		op.cfg.localPort = newPort
		return nil
	}

	newCfg := op.cfg
	newCfg.localPort = newPort
	if err := op.reStartup(newCfg); err != nil {
		return err
	}

	op.cfg.localPort = newPort
	return nil
}

func (op *OSOperator) ChangePACURL(newURL string) error {
	if newURL == op.cfg.pacURL {
		return nil
	}

	if !op.isStartup || op.cfg.mode == config.ModeGlobal {
		op.cfg.pacURL = newURL
		return nil
	}

	newCfg := op.cfg
	newCfg.pacURL = newURL
	if err := op.reStartup(newCfg); err != nil {
		return err
	}

	op.cfg.pacURL = newURL
	return nil
}

func (op *OSOperator) reStartup(newCfg opConfig) (result error) {
	oldMode := op.cfg.mode

	if newCfg.mode == oldMode {
		return resetProxy(newCfg)
	}

	// Clear old mode first. Because if we reset newCfg first and
	// clear oldMode failed, there two modes will exist at the same time.
	// Clear old mode first will cause the two modes all be cleared when
	// clearProxyByMode success but resetProxy failed, but I think it's better
	// all be cleared than exist at the same time.
	if err := clearProxyByMode(op.cfg.mode); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			resetProxy(op.cfg)
		}
	}()

	return resetProxy(newCfg)
}

func resetProxy(cfg opConfig) error {
	switch cfg.mode {
	case config.ModePAC:
		return setAutoProxy(cfg.pacURL)
	case config.ModeGlobal:
		return setGlobalProxy(cfg.localAddr, cfg.localPort)
	default:
		return fmt.Errorf("unknown mode name '%s'", cfg.mode)
	}
}

func clearProxyByMode(mode string) error {
	switch mode {
	case config.ModePAC:
		return clearAutoProxy()
	case config.ModeGlobal:
		return clearGlobalProxy()
	default:
		return fmt.Errorf("unknown mode name '%s'", mode)
	}
}

func setAutoProxy(url string) error {
	return setProxy(func(nw string) error {
		return api.setAutoProxyFor(nw, url)
	})
}

func setGlobalProxy(addr, port string) error {
	return setProxy(func(nw string) error {
		return api.setGlobalProxyFor(nw, addr, port)
	})
}

func clearAllProxy() error {
	return clearProxy(func(network string) error {
		err1 := api.clearAutoProxyFor(network)
		err2 := api.clearGlobalProxyFor(network)

		if err1 != nil {
			return err1
		} else if err2 != nil {
			return err2
		} else {
			return nil
		}
	})
}

func clearAutoProxy() error {
	return clearProxy(func(network string) error {
		return api.clearAutoProxyFor(network)
	})
}

func clearGlobalProxy() error {
	return clearProxy(func(network string) error {
		return api.clearGlobalProxyFor(network)
	})
}

func setProxy(setFunc func(string) error) error {
	clearAllProxy()

	networks, err := api.listAllNetwork()
	if err != nil {
		return err
	}

	for _, nw := range networks {
		if err := setFunc(nw); err != nil {
			clearAllProxy()
			return err
		}
	}

	return nil
}

func clearProxy(clearFun func(string) error) error {
	networks, err := api.listAllNetwork()
	if err != nil {
		return err
	}

	var resultErr error
	for _, nw := range networks {
		if err := clearFun(nw); err != nil {
			resultErr = errors.New(fmt.Sprintf("clear %s error: %v; ", nw, err))
		}
	}

	return resultErr
}
