package main

import (
	"strings"
	"testing"

	"github.com/fatcat22/ssctrl/config"
)

type proxyCoreMock struct {
	isStartup bool

	mode      string
	localPort string
	pacPort   string
	srvCfg    config.ServerConfig
}

func (cm *proxyCoreMock) Startup() error {
	cm.isStartup = true
	return nil
}

func (cm *proxyCoreMock) Shutdown() {
	cm.isStartup = false
}

func (cm *proxyCoreMock) ChangeMode(newMode string) error {
	cm.mode = newMode
	return nil
}

func (cm *proxyCoreMock) ChangeLocalPort(newPort string) error {
	cm.localPort = newPort
	return nil
}

func (cm *proxyCoreMock) ChangePACPort(newPort string) error {
	cm.pacPort = newPort
	return nil
}

func (cm *proxyCoreMock) ChangeServerConfig(newSrvCfg config.ServerConfig) error {
	cm.srvCfg = newSrvCfg
	return nil
}

func TestControlerStartupWithEnable(t *testing.T) {
	const expectMode = config.ModePAC
	const expectPACPort = "1234"
	const expectAPIPort = "4321"
	const expectLocalPort = "5678"
	cm := &proxyCoreMock{}

	expectSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "8899",
		Crypt:    config.Crypt_AEAD_AES_256_GCM,
		Password: "yourpwd",
	}
	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(expectMode); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetPACPort(expectPACPort); err != nil {
		t.Fatalf("AppConfig.SetPACPort error: %v", err)
	}
	if err := cfg.SetAPIPort(expectAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}
	if err := cfg.SetLocalPort(expectLocalPort); err != nil {
		t.Fatalf("AppConfig.SetLocalPort error: %v", err)
	}
	if err := cfg.UpdateServer("mysrv", expectSrvCfg); err != nil {
		t.Fatalf("AppConfig.UpdateServer error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	// check api server port
	addrs := strings.Split(ctrl.apiSrv.server.Addr, ":")
	if len(addrs) != 2 || addrs[1] != expectAPIPort {
		t.Errorf("expect api port %s but got %s", expectAPIPort, addrs)
	}

	// check proxy core
	if cm.isStartup != true {
		t.Errorf("controler is startup but core is not")
	}
}

func TestControlerStartupWithDisable(t *testing.T) {
	const expectAPIPort = "4321"
	cm := &proxyCoreMock{}

	cfg := config.NewConfig()
	cfg.SetEnabled(false)
	if err := cfg.SetAPIPort(expectAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if cm.isStartup == true {
		t.Errorf("controler is startup but core is disabled and should not be startup")
	}
}

func TestControlerEnable(t *testing.T) {
	const expectMode = config.ModePAC
	const expectAPIPort = "4321"
	cm := &proxyCoreMock{
		mode: expectMode,
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(false)
	if err := cfg.SetMode(expectMode); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort(expectAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy error: %v", err)
	}

	if cfg.IsEnabled() != true {
		t.Errorf("EnableProxy success but config is disable")
	}
	if cm.isStartup != true {
		t.Errorf("EnabledProxy success but core is not started up")
	}
	if cm.mode != expectMode {
		t.Errorf("expect core mode %s but got %s", expectMode, cm.mode)
	}
}

func TestControlerDisable(t *testing.T) {
	const expectMode = config.ModePAC
	const expectAPIPort = "4321"
	cm := &proxyCoreMock{
		isStartup: true,
		mode:      expectMode,
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(expectMode); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort(expectAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy error: %v", err)
	}

	if cfg.IsEnabled() != false {
		t.Errorf("DisableProxy success but config is enabled")
	}
	if cm.isStartup == true {
		t.Errorf("DisableProxy success but core is still startup")
	}
}

func TestControlerChangeMode(t *testing.T) {
	oldMode := config.ModeGlobal
	const expectMode = config.ModePAC
	const expectAPIPort = "4321"
	cm := &proxyCoreMock{
		isStartup: true,
		mode:      oldMode,
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(oldMode); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort(expectAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.ChangeMode(expectMode); err != nil {
		t.Fatalf("ChangeMode error: %v", err)
	}

	if cfg.IsEnabled() != true {
		t.Errorf("proxy should be enabled")
	}
	if cfg.GetMode() != expectMode {
		t.Errorf("expect mode %s but got %s", expectMode, cfg.GetMode())
	}
	if cm.isStartup != true {
		t.Errorf("core should be started up")
	}
	if cm.mode != expectMode {
		t.Errorf("expect mode %s but got %s", expectMode, cm.mode)
	}
}

func TestControlerChangeLocalPort(t *testing.T) {
	const expectMode = config.ModePAC
	const expectLocalPort = "1234"
	oldLocalPort := "2234"
	const expectAPIPort = "4321"
	cm := &proxyCoreMock{
		isStartup: true,
		mode:      expectMode,
		localPort: oldLocalPort,
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(expectMode); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort(expectAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}
	if err := cfg.SetLocalPort(oldLocalPort); err != nil {
		t.Fatalf("APPconfig.SetLocalPort error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.ChangeLocalPort(expectLocalPort); err != nil {
		t.Fatalf("ChangeLocalPort error: %v", err)
	}

	if cfg.IsEnabled() != true {
		t.Errorf("proxy should be enabled")
	}
	if cfg.GetMode() != expectMode {
		t.Errorf("expect mode %s but got %s", expectMode, cfg.GetMode())
	}
	if cm.isStartup != true {
		t.Errorf("core should be started up")
	}
	if cm.localPort != expectLocalPort {
		t.Errorf("expect local port %s but got %s", expectLocalPort, cm.localPort)
	}
}

func TestControlerChangePACPort(t *testing.T) {
	const expectMode = config.ModePAC
	const expectPACPort = "1234"
	oldPACPort := "2234"
	const expectAPIPort = "4321"
	cm := &proxyCoreMock{
		isStartup: true,
		mode:      expectMode,
		pacPort:   oldPACPort,
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(expectMode); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort(expectAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}
	if err := cfg.SetPACPort(oldPACPort); err != nil {
		t.Fatalf("APPconfig.SetLocalPort error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.ChangePACPort(expectPACPort); err != nil {
		t.Fatalf("ChangePACPort error: %v", err)
	}

	if cfg.IsEnabled() != true {
		t.Errorf("proxy should be enabled")
	}
	if cfg.GetMode() != expectMode {
		t.Errorf("expect mode %s but got %s", expectMode, cfg.GetMode())
	}
	if cm.isStartup != true {
		t.Errorf("core should be started up")
	}
	if cm.pacPort != expectPACPort {
		t.Errorf("expect local port %s but got %s", expectPACPort, cm.pacPort)
	}
}

func TestControlerChangeAPIPort(t *testing.T) {
	oldAPIPort := "2234"
	const expectAPIPort = "4321"
	const pacPort = "5432"
	const localPort = "6432"
	cm := &proxyCoreMock{
		isStartup: true,
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(config.ModePAC); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort(oldAPIPort); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}
	if err := cfg.SetPACPort(pacPort); err != nil {
		t.Fatalf("AppConfig.SetPACPort error: %v", err)
	}
	if err := cfg.SetLocalPort(localPort); err != nil {
		t.Fatalf("AppConfig.SetLocalPort error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.ChangeAPIPort(expectAPIPort); err != nil {
		t.Fatalf("ChangeAPIPort error: %v", err)
	}

	addrs := strings.Split(ctrl.apiSrv.server.Addr, ":")
	if len(addrs) != 2 || addrs[1] != expectAPIPort {
		t.Errorf("expect api port %s but got %s", expectAPIPort, ctrl.apiSrv.server.Addr)
	}
}

func TestControlerChangeCurrentServer(t *testing.T) {
	cm := &proxyCoreMock{}
	const oldSrvName = "server1"
	oldSrvCfg := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "1122",
		Crypt:    config.Crypt_AEAD_AES_256_GCM,
		Password: "server1pwd",
	}
	const expectSrvName = "server1"
	expectSrvCfg := config.ServerConfig{
		Address:  "99.88.77.66",
		Port:     "9988",
		Crypt:    config.Crypt_AEAD_CHACHA20_POLY1305,
		Password: "server2pwd",
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(config.ModePAC); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort("4321"); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}
	if err := cfg.UpdateServer(oldSrvName, oldSrvCfg); err != nil {
		t.Fatalf("AppConfig.UpdateServer error: %v", err)
	}
	if err := cfg.UpdateServer(expectSrvName, expectSrvCfg); err != nil {
		t.Fatalf("AppConfig.UpdateServer error: %v", err)
	}
	if err := cfg.SetCurrentServer(oldSrvName); err != nil {
		t.Fatalf("AppConfig.SetCurrentServer error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.ChangeCurrentServer(expectSrvName); err != nil {
		t.Fatalf("ChangeCurrentServer error: %v", err)
	}

	if cm.srvCfg != expectSrvCfg {
		t.Errorf("expect server config '%v' but got '%v'", expectSrvCfg, cm.srvCfg)
	}
}

func TestControlerUpdateServers(t *testing.T) {
	cm := &proxyCoreMock{}
	const currentSrvName = "server1"
	const srvWithoutCrypt = "server2"
	oldCurrentSrv := config.ServerConfig{
		Address:  "11.22.33.44",
		Port:     "1122",
		Crypt:    config.Crypt_AEAD_AES_256_GCM,
		Password: "server1pwd",
	}
	servers := map[string]config.ServerConfig{
		currentSrvName: config.ServerConfig{
			Address:  "119.229.339.449",
			Port:     "9129",
			Crypt:    config.Crypt_AEAD_AES_128_GCM,
			Password: "server1pwdnew",
		},
		srvWithoutCrypt: config.ServerConfig{
			Address:  "99.88.77.66",
			Port:     "7766",
			Password: "server2pwd",
		},
		"server3": config.ServerConfig{
			Address:  "69.78.77.86",
			Port:     "4455",
			Crypt:    config.Crypt_AEAD_CHACHA20_POLY1305,
			Password: "server2pwd",
		},
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(config.ModePAC); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort("4321"); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}
	if err := cfg.UpdateServer(currentSrvName, oldCurrentSrv); err != nil {
		t.Fatalf("AppConfig.UpdateServer error: %v", err)
	}
	if err := cfg.SetCurrentServer(currentSrvName); err != nil {
		t.Fatalf("AppConfig.SetCurrentServer error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	if err := ctrl.UpdateServers(servers); err != nil {
		t.Fatalf("UpdateServers error: %v", err)
	}

	// check current config in config
	name, srv := cfg.GetCurrentServerConfig()
	if name != currentSrvName {
		t.Errorf("expect current server %s but got %s", currentSrvName, name)
	}
	if srv != servers[currentSrvName] {
		t.Errorf("expect current server config '%v' but got '%v'", servers[currentSrvName], srv)
	}
	// check current config in core
	if cm.srvCfg != servers[currentSrvName] {
		t.Errorf("expect current server in Core '%v' but got '%v'", servers[currentSrvName], cm.srvCfg)
	}
	// check all config in config.AppConfig
	for name, expectCfg := range servers {
		srv, err := cfg.GetServerConfig(name)
		if err != nil {
			t.Errorf("get server by name %s error: %v", name, err)
			continue
		}
		if name == srvWithoutCrypt {
			expectCfg.Crypt = config.DefaultCrypt
		}

		if srv != expectCfg {
			t.Errorf("config of '%s' error: expect '%v' but got '%v'", name, expectCfg, srv)
			continue
		}
	}
}

func TestControlerRemoveServers(t *testing.T) {
	cm := &proxyCoreMock{}
	const srvName1 = "srvName1"
	const srvName2 = "srvName2"
	const currentSrvName = "srvName3"
	servers := map[string]config.ServerConfig{
		srvName1: config.ServerConfig{
			Address:  "119.229.339.449",
			Port:     "9129",
			Crypt:    config.Crypt_AEAD_AES_128_GCM,
			Password: "server1pwdnew",
		},
		srvName2: config.ServerConfig{
			Address:  "99.88.77.66",
			Port:     "7766",
			Password: "server2pwd",
		},
		currentSrvName: config.ServerConfig{
			Address:  "69.78.77.86",
			Port:     "4455",
			Crypt:    config.Crypt_AEAD_CHACHA20_POLY1305,
			Password: "server2pwd",
		},
	}

	cfg := config.NewConfig()
	cfg.SetEnabled(true)
	if err := cfg.SetMode(config.ModePAC); err != nil {
		t.Fatalf("AppConfig.SetMode error: %v", err)
	}
	if err := cfg.SetAPIPort("4321"); err != nil {
		t.Fatalf("AppConfig.SetAPIPort error: %v", err)
	}
	for name, srv := range servers {
		if err := cfg.UpdateServer(name, srv); err != nil {
			t.Fatalf("AppConfig.UpdateServer error: %v", err)
		}
	}
	if err := cfg.SetCurrentServer(currentSrvName); err != nil {
		t.Fatalf("AppConfig.SetCurrentServer error: %v", err)
	}

	ctrl, err := NewControler(cfg, cm, nil)
	if err != nil {
		t.Fatalf("NewControler error: %v", err)
	}

	if err := ctrl.Startup(); err != nil {
		t.Fatalf("Controler.Startup error: %v", err)
	}
	defer ctrl.Shutdown()

	// check remove current server
	err = ctrl.RemoveServers([]string{srvName1, currentSrvName})
	if err == nil {
		t.Errorf("RemoveServers success but current server could not be removed")
	}
	name, srv := cfg.GetCurrentServerConfig()
	if name != currentSrvName || srv != servers[currentSrvName] {
		t.Errorf("expect current server '%s:%v' but got '%s:%v'", currentSrvName, servers[currentSrvName], name, srv)
	}
	srv, err = cfg.GetServerConfig(currentSrvName)
	if err != nil {
		t.Errorf("can not find current server by name '%s'", currentSrvName)
	}
	if srv != servers[currentSrvName] {
		t.Errorf("expect current server config '%v' but got '%v'", servers[currentSrvName], srv)
	}
	srv, err = cfg.GetServerConfig(srvName1)
	if err != nil {
		t.Errorf("can not find server config by name '%s'", srvName1)
	}
	if srv != servers[srvName1] {
		t.Errorf("expect server config '%v' but got '%v'", servers[srvName1], srv)
	}

	// check remove server success
	if err := ctrl.RemoveServers([]string{srvName1}); err != nil {
		t.Errorf("RemoveServers error: %v", err)
	}
	if _, err := cfg.GetServerConfig(srvName1); err == nil {
		t.Errorf("expect server '%s' has been removed but not", srvName1)
	}
	name, srv = cfg.GetCurrentServerConfig()
	if name != currentSrvName {
		t.Errorf("expect current server %s but got %s", currentSrvName, name)
	}
	if srv != servers[currentSrvName] {
		t.Errorf("expect current server config '%v' but got '%v'", servers[currentSrvName], srv)
	}
}
