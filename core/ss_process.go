package core

import (
	"os"

	"github.com/fatcat22/ssctrl/config"
)

type ssProcess interface {
	Kill() error
}

type ssProcessImpl struct {
	proc *os.Process
}

func (ssp *ssProcessImpl) Kill() error {
	if ssp.proc == nil {
		return nil
	}

	if err := ssp.proc.Kill(); err != nil {
		return err
	}
	ssp.proc.Wait() // don't care aboult return value of wait

	ssp.proc = nil
	return nil
}

type ssProcessMock struct {
	localAddr string
	localPort string
	srvCfg    config.ServerConfig

	killed bool
}

func newSSProcessMock(localAddr, localPort string, srvCfg config.ServerConfig) (ssProcess, error) {
	return &ssProcessMock{
		localAddr: localAddr,
		localPort: localPort,
		srvCfg:    srvCfg,

		killed: false,
	}, nil
}

func (ssm *ssProcessMock) Kill() error {
	ssm.killed = true
	return nil
}
