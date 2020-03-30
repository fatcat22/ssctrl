package core

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/fatcat22/ssctrl/common"
	"github.com/fatcat22/ssctrl/config"
)

func init() {
	isOnTest = true
	api = newOSAPI()
}

func TestProxyCoreStartup(t *testing.T) {
	const expectPACPort = "1234"
	const expectLocalPort = "2234"
	const expectMode = config.ModePAC
	const apiExpectMode = modePAC
	expectSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "3234",
		Crypt:    config.Crypt_AEAD_AES_128_GCM,
		Password: "yourpwd",
	}
	const expectPACData = "hello, testing ProxyCore"
	tmpPACFile := createMockPACFile(expectPACData, t)
	defer os.Remove(tmpPACFile)

	core, err := NewProxyCore(expectPACPort, tmpPACFile, expectLocalPort, expectMode, expectSrvCfg, "")
	if err != nil {
		t.Fatalf("create ProxyCore error: %v", err)
	}
	defer core.Shutdown()
	if err := core.Startup(); err != nil {
		t.Fatalf("ProxyCore.Startup error: %v", err)
	}

	if core.pacPort != expectPACPort {
		t.Errorf("ProxyCore: expect pac port %s but got %s", expectPACPort, core.pacPort)
	}
	if core.localPort != expectLocalPort {
		t.Errorf("ProxyCore: expect local port %s but got %s", expectLocalPort, core.localPort)
	}
	if core.mode != expectMode {
		t.Errorf("ProxyCore: expect mode %s but got %s", expectMode, core.mode)
	}
	if core.srvCfg != expectSrvCfg {
		t.Errorf("ProxyCore: expect server config %v but got %v", expectSrvCfg, core.srvCfg)
	}

	// check pac server
	resp, err := http.Get(core.pacSrv.GetPACURL())
	if err != nil {
		t.Errorf("get pac data error: %v", err)
	} else {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("read pac data error: %v", err)
		} else {
			if string(data) != expectPACData {
				t.Errorf("http get pac data failed: expect '%s' but got '%s'", expectPACData, data)
			}
		}
	}

	// check os settings
	mapi := api.(*mockAPI)
	for name, s := range mapi.settings {
		if s.mode != apiExpectMode {
			t.Errorf("network %s expect proxy mode %d but got %d", name, apiExpectMode, s.mode)
		}
		if s.pacURL != core.pacSrv.GetPACURL() {
			t.Errorf("network %s expect pac url %s but got %s", name, core.pacSrv.GetPACURL(), s.pacURL)
		}
	}

	// check ss
	ssm := core.ss.proc.(*ssProcessMock)
	if ssm.killed == true {
		t.Errorf("shadowsocks2 process should be started")
	}
	if ssm.localPort != expectLocalPort {
		t.Errorf("shadowsocks2 expect local port %s but got %s", expectLocalPort, ssm.localPort)
	}
	if ssm.srvCfg != expectSrvCfg {
		t.Errorf("shadowsocks2 expect server config %v but got %v", expectSrvCfg, ssm.srvCfg)
	}
}

func TestProxyCoreShutdown(t *testing.T) {
	const expectMode = config.ModePAC
	const apiExpectMode = modePAC
	expectSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "3234",
		Crypt:    config.Crypt_AEAD_AES_128_GCM,
		Password: "yourpwd",
	}
	const expectPACData = "hello, testing ProxyCore"
	tmpPACFile := createMockPACFile(expectPACData, t)
	defer os.Remove(tmpPACFile)

	core, err := NewProxyCore("1234", tmpPACFile, "2234", expectMode, expectSrvCfg, "")
	if err != nil {
		t.Fatalf("NewProxyCore error: %v", err)
	}
	if err := core.Startup(); err != nil {
		t.Fatalf("ProxyCore.Startup error: %v", err)
	}
	defer core.Shutdown()
	//make sure sub-modules are started
	if !core.pacSrv.isStartup {
		t.Error("pac server is not startup")
	}
	if !core.ss.isStartup {
		t.Error("ss is not startup")
	}
	if !core.op.isStartup {
		t.Error("os operator is not startup")
	}

	core.Shutdown()

	// check sub-modules are all shuted down
	if core.pacSrv.isStartup {
		t.Error("ProxyCore is shubdown but pac server is not")
	}
	if core.ss.isStartup {
		t.Error("ProxyCorre is shutdown but ss is not")
	}
	if core.op.isStartup {
		t.Error("ProxyCore is shutdown but os operator is not")
	}
}

func TestProxyCoreChangeModeToGlobal(t *testing.T) {
	testChangeMode(config.ModePAC, config.ModeGlobal, modeGlobal, t)
}

func TestProxyCoreChangeModeToPAC(t *testing.T) {
	testChangeMode(config.ModeGlobal, config.ModePAC, modePAC, t)
}

func TestProxyCoreChangeLocalPortOnPAC(t *testing.T) {
	testProxyCoreChangeLocalPort(config.ModePAC, t)
}

func TestProxyCoreChangeLocalPortOnGlobal(t *testing.T) {
	testProxyCoreChangeLocalPort(config.ModeGlobal, t)
}

func TestProxyCoreChangePACPortOnPAC(t *testing.T) {
	testProxyCoreChangePACPort(config.ModePAC, t)
}

func TestProxyCoreChangePACPortOnGlobal(t *testing.T) {
	testProxyCoreChangePACPort(config.ModeGlobal, t)
}

func TestProxyCoreChangeServerConfig(t *testing.T) {
	oldSrvCfg := config.ServerConfig{
		Address:  "99.88.77.66",
		Port:     "9876",
		Crypt:    config.Crypt_AEAD_CHACHA20_POLY1305,
		Password: "yourpwdxxx",
	}
	expectSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "3234",
		Crypt:    config.Crypt_AEAD_AES_128_GCM,
		Password: "yourpwd",
	}
	const expectPACData = "hello, testing ProxyCore"
	tmpPACFile := createMockPACFile(expectPACData, t)
	defer os.Remove(tmpPACFile)

	core, err := NewProxyCore("1234", tmpPACFile, "9100", config.ModeGlobal, oldSrvCfg, "")
	if err != nil {
		t.Fatalf("create ProxyCore error: %v", err)
	}
	defer core.Shutdown()
	if err := core.Startup(); err != nil {
		t.Fatalf("ProxyCore.Startup error: %v", err)
	}
	if err := core.ChangeServerConfig(expectSrvCfg); err != nil {
		t.Fatalf("ProxyCore.ChangePACPort error: %v", err)
	}

	ssm := core.ss.proc.(*ssProcessMock)
	if ssm.srvCfg != expectSrvCfg {
		t.Errorf("expect ss server %v but got %v", expectSrvCfg, ssm.srvCfg)
	}
}

func createMockPACFile(data string, t *testing.T) string {
	tmpFile, err := common.TempFile()
	if err != nil {
		t.Fatalf("create tempalte pac error: %v", err)
	}

	if err := ioutil.WriteFile(tmpFile, []byte(data), os.ModePerm); err != nil {
		t.Fatalf("write mock pac file error: %v", err)
	}

	return tmpFile
}

func testChangeMode(oldMode, toMode string, expectMode int, t *testing.T) {
	const expectPACPort = "1234"
	const expectLocalPort = "2234"
	expectSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "3234",
		Crypt:    config.Crypt_AEAD_AES_128_GCM,
		Password: "yourpwd",
	}
	const expectPACData = "hello, testing ProxyCore"
	tmpPACFile := createMockPACFile(expectPACData, t)
	defer os.Remove(tmpPACFile)

	core, err := NewProxyCore(expectPACPort, tmpPACFile, expectLocalPort, oldMode, expectSrvCfg, "")
	if err != nil {
		t.Fatalf("create ProxyCore error: %v", err)
	}
	defer core.Shutdown()
	if err := core.Startup(); err != nil {
		t.Fatalf("ProxyCore.Startup error: %v", err)
	}
	if err := core.ChangeMode(toMode); err != nil {
		t.Fatalf("change mode to %s error: %v", toMode, err)
	}

	mapi := api.(*mockAPI)
	for name, s := range mapi.settings {
		if s.mode != expectMode {
			t.Errorf("network %s: expect mode %d but got %d", name, expectMode, s.mode)
		}

		switch toMode {
		case config.ModeGlobal:
			if s.globalPort != expectLocalPort {
				t.Errorf("expect local port %s but got %s", expectLocalPort, s.globalPort)
			}
		case config.ModePAC:
			if s.pacURL != core.pacSrv.GetPACURL() {
				t.Errorf("expect pac url %s but got %s", core.pacSrv.GetPACURL(), s.pacURL)
			}
		default:
			t.Fatalf("unknown mode %s", toMode)
		}
	}
}

func testProxyCoreChangeLocalPort(mode string, t *testing.T) {
	const oldLocalPort = "2345"
	const expectLocalPort = "6789"
	expectSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "3234",
		Crypt:    config.Crypt_AEAD_AES_128_GCM,
		Password: "yourpwd",
	}
	const expectPACData = "hello, testing ProxyCore"
	tmpPACFile := createMockPACFile(expectPACData, t)
	defer os.Remove(tmpPACFile)

	core, err := NewProxyCore("1234", tmpPACFile, oldLocalPort, mode, expectSrvCfg, "")
	if err != nil {
		t.Fatalf("create ProxyCore error: %v", err)
	}
	defer core.Shutdown()
	if err := core.Startup(); err != nil {
		t.Fatalf("ProxyCore.Startup error: %v", err)
	}
	if err := core.ChangeLocalPort(expectLocalPort); err != nil {
		t.Fatalf("ProxyCore.ChangeLocalPort error: %v", err)
	}

	// check os operator
	mapi := api.(*mockAPI)
	for name, s := range mapi.settings {
		switch mode {
		case config.ModePAC:
			if s.mode != modePAC {
				t.Errorf("network %s: expect mode %d but got %d", name, modePAC, s.mode)
			}
			if s.globalPort != "" {
				t.Errorf("network %s: proxy mode is pac, expect local port empty but got %s", name, s.globalPort)
			}
		case config.ModeGlobal:
			if s.mode != modeGlobal {
				t.Errorf("network %s: expect mode %d but got %d", name, modeGlobal, s.mode)
			}
			if s.globalPort != expectLocalPort {
				t.Errorf("network %s: expect local port %s but got %s", name, expectLocalPort, s.globalPort)
			}
			if s.pacURL != "" {
				t.Errorf("network %s: proxy mode is global, expect pac url is empty but got %s", name, s.pacURL)
			}
		default:
			t.Fatalf("unknown mode %s", mode)
		}
	}

	// check ss
	ssm := core.ss.proc.(*ssProcessMock)
	if ssm.localPort != expectLocalPort {
		t.Errorf("expect ss local port %s but got %s", expectLocalPort, ssm.localPort)
	}

	// check pac server
	//TODO: should using http.GET to get pac data
	//and analyze port from the response data.
	if core.pacSrv.localPort != expectLocalPort {
		t.Errorf("expect pac server local port %s but got %s", expectLocalPort, ssm.localPort)
	}
}

func testProxyCoreChangePACPort(mode string, t *testing.T) {
	const oldPACPort = "2345"
	const expectPACPort = "6789"
	expectSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "3234",
		Crypt:    config.Crypt_AEAD_AES_128_GCM,
		Password: "yourpwd",
	}
	const expectPACData = "hello, testing ProxyCore"
	tmpPACFile := createMockPACFile(expectPACData, t)
	defer os.Remove(tmpPACFile)

	core, err := NewProxyCore(oldPACPort, tmpPACFile, "9100", mode, expectSrvCfg, "")
	if err != nil {
		t.Fatalf("create ProxyCore error: %v", err)
	}
	defer core.Shutdown()
	if err := core.Startup(); err != nil {
		t.Fatalf("ProxyCore.Startup error: %v", err)
	}
	if err := core.ChangePACPort(expectPACPort); err != nil {
		t.Fatalf("ProxyCore.ChangePACPort error: %v", err)
	}

	addrs := strings.Split(core.pacSrv.server.Addr, ":")
	if len(addrs) != 2 || addrs[1] != expectPACPort {
		t.Errorf("expect pac server port %s but got %s", expectPACPort, core.pacSrv.server.Addr)
	}
	mapi := api.(*mockAPI)
	for name, s := range mapi.settings {
		switch mode {
		case config.ModePAC:
			if s.mode != modePAC {
				t.Errorf("network %s: expect mode %d but got %d", name, modePAC, s.mode)
			}
			if s.pacURL != core.pacSrv.GetPACURL() {
				t.Errorf("network %s: expect pac url %s but got %s", name, core.pacSrv.GetPACURL(), s.pacURL)
			}
		case config.ModeGlobal:
			if s.mode != modeGlobal {
				t.Errorf("network %s: expect mode %d but got %d", name, modeGlobal, s.mode)
			}
			if s.pacURL != "" {
				t.Errorf("network %s: proxy mode is global, expect pac empty but got %s", name, s.pacURL)
			}
		default:
			t.Fatalf("unknown mode %s", mode)
		}
	}
}
