package config

import (
	"html/template"
	"io/ioutil"
	"os"
	"testing"
)

type configData struct {
	Enabled bool
	Mode    string

	LocalPort    string
	LocalAddress string

	PACPort string
	APIPort string

	UsingServer string

	Server1Name     string
	Server1Addr     string
	Server1Port     string
	Server1Crypto   string
	Server1Password string

	Server2Name     string
	Server2Addr     string
	Server2Port     string
	Server2Crypto   string
	Server2Password string
}

var cfgData = configData{
	Enabled: true,
	Mode:    "pac",

	LocalPort:    "1081",
	LocalAddress: "127.0.0.1",

	PACPort: "1082",
	APIPort: "1083",

	UsingServer: "myserver2",

	Server1Name:     "myserver1",
	Server1Addr:     "11.22.33.44",
	Server1Port:     "8088",
	Server1Crypto:   "AEAD_CHACHA20_POLY1305",
	Server1Password: "1234abcd",

	Server2Name:     "myserver2",
	Server2Addr:     "www.example.com",
	Server2Port:     "9099",
	Server2Crypto:   "AEAD_AES_256_GCM",
	Server2Password: "examplepwd",
}

func TestLoadConfig(t *testing.T) {
	const cfgTemplate = `
    # ssctrl config file

    enabled = {{.Enabled}}
    mode = "{{.Mode}}"

    localPort = "{{.LocalPort}}"
    localAddress = "{{.LocalAddress}}"

    pacPort = "{{.PACPort}}"
    apiPort = "{{.APIPort}}"

    usingServer = "{{.UsingServer}}"


    [servers]
        [servers.{{.Server2Name}}]
            address = "{{.Server2Addr}}"
            port = "{{.Server2Port}}"
            crypto = "{{.Server2Crypto}}"
            password = "{{.Server2Password}}"
        [servers.{{.Server1Name}}]
        address = "{{.Server1Addr}}"
        port = "{{.Server1Port}}"
        crypto = "{{.Server1Crypto}}"
        password = "{{.Server1Password}}"
    `

	cfgFile, err := ioutil.TempFile("", "ssctrl")
	if err != nil {
		t.Fatalf("create template file failed: %v", err)
	}
	cfgPath := cfgFile.Name()
	defer func() {
		cfgFile.Close()
		os.Remove(cfgPath)
	}()

	tobj := template.Must(template.New("config").Parse(cfgTemplate))
	if err := tobj.Execute(cfgFile, cfgData); err != nil {
		t.Fatalf("execute template error: %v", err)
	}

	cfgFile.Close()

	appCfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("load config error: %v", err)
	}

	if appCfg.IsEnabled() != cfgData.Enabled {
		t.Errorf("unexpect enabled value: expect %t but got %t", cfgData.Enabled, appCfg.IsEnabled())
	}
	if appCfg.GetMode() != cfgData.Mode {
		t.Errorf("unexpect mode: expect %s but got %s", cfgData.Mode, appCfg.GetMode())
	}
	if appCfg.GetLocalPort() != cfgData.LocalPort {
		t.Errorf("unexpect local port: expect %s but got %s", cfgData.LocalPort, appCfg.GetLocalPort())
	}
	if appCfg.GetPACPort() != cfgData.PACPort {
		t.Errorf("unexpect pac port: expect %s but got %s", cfgData.PACPort, appCfg.GetPACPort())
	}
	if appCfg.GetAPIPort() != cfgData.APIPort {
		t.Errorf("unexpect control port: expect %s but got %s", cfgData.APIPort, appCfg.GetAPIPort())
	}
	if name, _ := appCfg.GetCurrentServerConfig(); name != cfgData.UsingServer {
		t.Errorf("unexpect current server name: expect %s but got %s", cfgData.UsingServer, name)
	}

	srv1, err := appCfg.GetServerConfig(cfgData.Server1Name)
	if err != nil {
		t.Fatalf("get config for server '%s' error: %v", cfgData.Server1Name, err)
	}
	if srv1.Address != cfgData.Server1Addr {
		t.Errorf("address of server '%s' error: expect %s but got %s", cfgData.Server1Name, cfgData.Server1Addr, srv1.Address)
	}
	if srv1.Port != cfgData.Server1Port {
		t.Errorf("port of server '%s' error: expect %s but got %s", cfgData.Server1Name, cfgData.Server1Port, srv1.Port)
	}
	if srv1.Crypt != cfgData.Server1Crypto {
		t.Errorf("crypto method of server '%s' error: expect %s but got %s", cfgData.Server1Name, cfgData.Server1Crypto, srv1.Crypt)
	}
	if srv1.Password != cfgData.Server1Password {
		t.Errorf("crtypt password of server '%s' error: expect %s but got %s", cfgData.Server1Name, cfgData.Server1Password, srv1.Password)
	}

	srv2, err := appCfg.GetServerConfig(cfgData.Server2Name)
	if err != nil {
		t.Fatalf("get config for server '%s' error: %v", cfgData.Server2Name, err)
	}
	if srv2.Address != cfgData.Server2Addr {
		t.Errorf("address of server '%s' error: expect %s but got %s", cfgData.Server2Name, cfgData.Server2Addr, srv2.Address)
	}
	if srv2.Port != cfgData.Server2Port {
		t.Errorf("port of server '%s' error: expect %s but got %s", cfgData.Server2Name, cfgData.Server2Port, srv2.Port)
	}
	if srv2.Crypt != cfgData.Server2Crypto {
		t.Errorf("crypto method of server '%s' error: expect %s but got %s", cfgData.Server2Name, cfgData.Server2Crypto, srv2.Crypt)
	}
	if srv2.Password != cfgData.Server2Password {
		t.Errorf("crtypt password of server '%s' error: expect %s but got %s", cfgData.Server2Name, cfgData.Server2Password, srv2.Password)
	}
}

func TestDefaultConfig(t *testing.T) {
	const cfgTemplate = `
    [servers]
    [servers.{{.Server2Name}}]
        address = "{{.Server2Addr}}"
        port = "{{.Server2Port}}"
        password = "{{.Server2Password}}"
`
	cfgFile, err := ioutil.TempFile("", "ssctrl")
	if err != nil {
		t.Fatalf("create template file failed: %v", err)
	}
	cfgPath := cfgFile.Name()
	defer func() {
		cfgFile.Close()
		os.Remove(cfgPath)
	}()

	tobj := template.Must(template.New("config").Parse(cfgTemplate))
	if err := tobj.Execute(cfgFile, cfgData); err != nil {
		t.Fatalf("execute template error: %v", err)
	}

	cfgFile.Close()

	appCfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("load config error: %v", err)
	}

	if appCfg.IsEnabled() != defaultEnabled {
		t.Errorf("unexpect default enabled value: expect %t but got %t", defaultEnabled, appCfg.IsEnabled())
	}
	if appCfg.GetMode() != defaultMode {
		t.Errorf("unexpect default mode: expect %s but got %s", defaultMode, appCfg.GetMode())
	}
	if appCfg.GetLocalPort() != defaultLocalPort {
		t.Errorf("unexpect default local port: expect %s but got %s", defaultLocalPort, appCfg.GetLocalPort())
	}
	if appCfg.GetPACPort() != defaultPACPort {
		t.Errorf("unexpect default pac port: expect %s but got %s", defaultPACPort, appCfg.GetAPIPort())
	}
	if appCfg.GetAPIPort() != defaultAPIPort {
		t.Errorf("unexpect default control port: expect %s but got %s", defaultAPIPort, appCfg.GetAPIPort())
	}
	srvName, srv := appCfg.GetCurrentServerConfig()
	if srvName != cfgData.Server2Name {
		t.Errorf("unexpect default current server name: expect %s but got %s", srvName, srvName)
	}
	if srv.Address != cfgData.Server2Addr {
		t.Errorf("address of default server '%s' error: expect %s but got %s", srvName, cfgData.Server2Addr, srv.Address)
	}
	if srv.Port != cfgData.Server2Port {
		t.Errorf("port of default server '%s' error: expect %s but got %s", srvName, cfgData.Server2Port, srv.Port)
	}
	if srv.Crypt != DefaultCrypt {
		t.Errorf("crypto method of default server '%s' error: expect %s but got %s", srvName, DefaultCrypt, srv.Crypt)
	}
	if srv.Password != cfgData.Server2Password {
		t.Errorf("crtypt password of server '%s' error: expect %s but got %s", srvName, cfgData.Server2Password, srv.Password)
	}
}
