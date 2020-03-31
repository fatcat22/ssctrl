package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatcat22/ssctrl/common"
	"github.com/fatcat22/ssctrl/config"
)

var ssProcessIsMock = false

type ShadowSocks struct {
	localAddr string
	localPort string
	srvCfg    config.ServerConfig

	ssPath string
	proc   ssProcess

	isStartup bool
}

func NewShadowSocks(ssPath, localAddr, localPort string, srvCfg config.ServerConfig) (*ShadowSocks, error) {
	return &ShadowSocks{
		localAddr: localAddr,
		localPort: localPort,
		srvCfg:    srvCfg,

		ssPath: ssPath,
		proc:   nil,

		isStartup: false,
	}, nil
}

func (ss *ShadowSocks) Startup() error {
	if ss.isStartup {
		return nil
	}

	proc, err := startSSProcess(ss.ssPath, ss.localAddr, ss.localPort, ss.srvCfg)
	if err != nil {
		return err
	}

	ss.proc = proc
	ss.isStartup = true
	return nil
}

func (ss *ShadowSocks) Shutdown() error {
	if !ss.isStartup || ss.proc == nil {
		return nil
	}

	if err := ss.proc.Kill(); err != nil {
		return err
	}

	ss.isStartup = false
	return nil
}

func (ss *ShadowSocks) ChangeLocalPort(newPort string) (result error) {
	if newPort == ss.localPort {
		return nil
	}
	defer func() {
		if result == nil {
			ss.localPort = newPort
		}
	}()

	if !ss.isStartup {
		return nil
	}

	newProc, err := startSSProcess(ss.ssPath, ss.localAddr, newPort, ss.srvCfg)
	if err != nil {
		return err
	}

	ss.proc.Kill()
	ss.proc = newProc
	return nil
}

func (ss *ShadowSocks) ChangeServerConfig(newSrvCfg config.ServerConfig) (result error) {
	if newSrvCfg == ss.srvCfg {
		return nil
	}
	defer func() {
		if result == nil {
			ss.srvCfg = newSrvCfg
		}
	}()

	if !ss.isStartup {
		return nil
	}

	// Because local port is not changed,
	// we must kill old ss process first, or the new
	// ss process will listen local port failed.
	if err := ss.proc.Kill(); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			ss.proc, _ = startSSProcess(ss.ssPath, ss.localAddr, ss.localPort, ss.srvCfg)
		}
	}()

	newProc, err := startSSProcess(ss.ssPath, ss.localAddr, ss.localPort, newSrvCfg)
	if err != nil {
		return err
	}
	ss.proc = newProc
	return nil
}

func startSSProcess(ssPath, localAddr, localPort string, srvCfg config.ServerConfig) (ssProcess, error) {
	if isOnTest {
		return newSSProcessMock(localAddr, localPort, srvCfg)
	}

	if len(ssPath) == 0 {
		var err error
		ssPath, err = getSSPath()
		if err != nil {
			return nil, err
		}
	}

	argv := []string{
		ssPath,
		"-c",
		fmt.Sprintf("ss://%s:%s@%s:%s", srvCfg.Crypt, srvCfg.Password, srvCfg.Address, srvCfg.Port),
		"-socks",
		localAddr + ":" + localPort,
		"-u",
		"-udptun",
		":8053=8.8.8.8:53,:8054=8.8.4.4:53",
		"-tcptun",
		":8053=8.8.8.8:53,:84=8.8.4.4:53",
	}

	file, err := os.Create(filepath.Join(common.HomeDir(), ".ssctrl", "ss2.log"))
	if err != nil {
		return nil, err
	}
	attr := &os.ProcAttr{
		Files: []*os.File{
			os.Stdin,
			file,
			file,
		},
	}

	proc, err := os.StartProcess(ssPath, argv, attr)
	if err != nil {
		return nil, err
	}

	return &ssProcessImpl{
		proc: proc,
	}, nil
}

func getSSPath() (string, error) {
	ssName := "go-shadowsocks2"
	if runtime.GOOS == "windows" {
		ssName = "go-shadowsocks2.exe"
	}

	exePath, err := common.ExeFile()
	if err != nil {
		return "", err
	}
	ssPath := filepath.Join(filepath.Dir(exePath), ssName)
	if _, err := os.Stat(ssPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("can not find shadowsocks program at '%s'", ssPath)
		}
	}

	return ssPath, nil
}
