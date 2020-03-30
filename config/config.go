package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/fatcat22/ssctrl/common"
	toml "github.com/pelletier/go-toml"
)

type ServerConfig struct {
	Address  string `toml:"address" json:"address"`
	Port     string `toml:"port" json:"port"`
	Crypt    string `toml:"crypto,omitempty" json:"crypto"`
	Password string `toml:"password" json:"password"`
}

type appConfig struct {
	Enabled     bool   `toml:"enabled,omitempty" json:"enabled"`
	Autorun     bool   `toml:"autorun,omitempty" json:"autorun"`
	Mode        string `toml:"mode,omitempty" json:"mode"`
	LocalPort   string `toml:"localPort,omitempty" json:"localPort"`
	PACPort     string `toml:"pacPort,omitempty" json:"pacPort"`
	APIPort     string `toml:"apiPort,omitempty" json:"apiPort"`
	UsingServer string `toml:"usingServer,omitempty" json:"usingServer"`

	Servers map[string]*ServerConfig `toml:"servers" json:"servers"`
}

const (
	ModePAC                      = "pac"
	ModeGlobal                   = "global"
	Crypt_AEAD_AES_128_GCM       = "AEAD_AES_128_GCM"
	Crypt_AEAD_AES_256_GCM       = "AEAD_AES_256_GCM"
	Crypt_AEAD_CHACHA20_POLY1305 = "AEAD_CHACHA20_POLY1305"

	DefaultCrypt     = Crypt_AEAD_CHACHA20_POLY1305
	defaultEnabled   = true
	defaultMode      = ModePAC
	defaultLocalPort = "1080"
	defaultPACPort   = "1082"
	defaultAPIPort   = "1083"
)

var (
	modeValues map[string]struct{} = map[string]struct{}{
		ModePAC:    struct{}{},
		ModeGlobal: struct{}{},
	}
	cryptoValues map[string]struct{} = map[string]struct{}{
		Crypt_AEAD_AES_128_GCM:       struct{}{},
		Crypt_AEAD_AES_256_GCM:       struct{}{},
		Crypt_AEAD_CHACHA20_POLY1305: struct{}{},
	}
)

var defaultCfg = appConfig{
	Enabled:   false,
	Mode:      defaultMode,
	LocalPort: defaultLocalPort,
	PACPort:   defaultPACPort,
	APIPort:   defaultAPIPort,

	Servers: make(map[string]*ServerConfig),
}

// AppConfig contains config information for the application.
// It don't export appConfig for external packge to
// change config filed directly.
type AppConfig struct {
	c appConfig

	currentServer *ServerConfig
}

func IsValidMode(m string) bool {
	_, ok := modeValues[m]
	return ok
}

func IsValidCryptoMethod(cm string) bool {
	_, ok := cryptoValues[cm]
	return ok
}

func SetServerConfigDefault(srvCfg *ServerConfig) {
	if len(srvCfg.Crypt) == 0 {
		srvCfg.Crypt = DefaultCrypt
	}
}

func NewConfig() *AppConfig {
	return &AppConfig{
		c: appConfig{
			Servers: make(map[string]*ServerConfig),
		},
	}
}

func LoadConfig(cfgFile string) (*AppConfig, error) {
	cfgBytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	cfg := AppConfig{
		c: defaultCfg,
	}

	if err = toml.Unmarshal(cfgBytes, &cfg.c); err != nil {
		return nil, err
	}

	if err := cfg.check(); err != nil {
		return nil, err
	}

	cfg.setServerDefault()
	return &cfg, nil
}

func RestoreConfig(appCfg *AppConfig, cfgFile string) error {
	data, err := toml.Marshal(appCfg.c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(cfgFile, data, os.ModePerm)
}

func (ac *AppConfig) GetServerConfig(srvName string) (ServerConfig, error) {
	srv, ok := ac.c.Servers[srvName]
	if !ok {
		return ServerConfig{}, fmt.Errorf("unknown server name '%s'", srvName)
	}

	return *srv, nil
}

func (ac *AppConfig) GetCurrentServerConfig() (string, ServerConfig) {
	return ac.c.UsingServer, *ac.currentServer
}

func (ac *AppConfig) SetCurrentServer(name string) error {
	srv, ok := ac.c.Servers[name]
	if !ok {
		return fmt.Errorf("not exist server name '%s'", name)
	}

	ac.c.UsingServer = name
	ac.currentServer = srv
	return nil
}

func (ac *AppConfig) SetCurrentServerMust(name string) {
	if err := ac.SetCurrentServer(name); err != nil {
		panic(fmt.Sprintf("SetCurrentServer error: %v", err))
	}
}

func (ac *AppConfig) CheckServerConfig(srv ServerConfig) error {
	if !common.IsValidPort(srv.Port) {
		return fmt.Errorf("invalid server port '%s'", srv.Port)
	}
	if srv.Crypt != "" {
		if !IsValidCryptoMethod(srv.Crypt) {
			return fmt.Errorf("invalid crypto method '%s'", srv.Crypt)
		}
	}

	return nil
}

func (ac *AppConfig) UpdateServer(name string, srv ServerConfig) error {
	if name == "" {
		return errors.New("server name is empty")
	}

	if err := ac.CheckServerConfig(srv); err != nil {
		return err
	}

	if srv.Crypt == "" {
		srv.Crypt = DefaultCrypt
	}
	ac.c.Servers[name] = &srv

	if ac.c.UsingServer == name {
		ac.currentServer = &srv
	}
	return nil
}

func (ac *AppConfig) UpdateServerMust(name string, srv ServerConfig) {
	if err := ac.UpdateServer(name, srv); err != nil {
		panic(fmt.Sprintf("UpdateServer for ''%s':'%v' error: %v", name, srv, err))
	}
}

func (ac *AppConfig) CheckServerBeRemoved(name string) error {
	_, ok := ac.c.Servers[name]
	if !ok {
		return fmt.Errorf("server name '%s' not exist", name)
	}
	if name == ac.c.UsingServer {
		return fmt.Errorf("server '%s' is being used and can not be removed", name)
	}

	return nil
}

func (ac *AppConfig) RemoveServerMust(name string) {
	if err := ac.CheckServerBeRemoved(name); err != nil {
		panic(fmt.Sprintf("check server name error: %v", err))
	}

	delete(ac.c.Servers, name)
}

func (ac *AppConfig) IsEnabled() bool {
	return ac.c.Enabled
}

func (ac *AppConfig) SetEnabled(enable bool) {
	ac.c.Enabled = enable
}

func (ac *AppConfig) SetAutorun(enable bool) {
	ac.c.Autorun = enable
}

func (ac *AppConfig) GetMode() string {
	return ac.c.Mode
}

func (ac *AppConfig) SetMode(newMode string) error {
	if !IsValidMode(newMode) {
		return fmt.Errorf("invalid mode name '%s'", newMode)
	}

	ac.c.Mode = newMode
	return nil
}

func (ac *AppConfig) SetModeMust(newMode string) {
	if !IsValidMode(newMode) {
		panic(fmt.Sprintf("invalid mode name '%s'", newMode))
	}

	ac.c.Mode = newMode
}

func (ac *AppConfig) GetAPIPort() string {
	return ac.c.APIPort
}

func (ac *AppConfig) CheckAPIPort(port string) error {
	return ac.checkPort(port, &ac.c.APIPort)
}

func (ac *AppConfig) SetAPIPort(newPort string) error {
	if err := ac.CheckAPIPort(newPort); err != nil {
		return err
	}

	ac.c.APIPort = newPort
	return nil
}

func (ac *AppConfig) SetAPIPortMust(newPort string) {
	if err := ac.SetAPIPort(newPort); err != nil {
		panic(fmt.Sprintf("SetAPIPort error: %v", err))
	}
}

func (ac *AppConfig) GetLocalPort() string {
	return ac.c.LocalPort
}

func (ac *AppConfig) CheckLocalPort(port string) error {
	return ac.checkPort(port, &ac.c.LocalPort)
}

func (ac *AppConfig) SetLocalPort(newPort string) error {
	if err := ac.CheckLocalPort(newPort); err != nil {
		return err
	}

	ac.c.LocalPort = newPort
	return nil
}

func (ac *AppConfig) SetLocalPortMust(newPort string) {
	if err := ac.SetLocalPort(newPort); err != nil {
		panic(fmt.Sprintf("SetLocalPort error: %v", err))
	}
}

func (ac *AppConfig) GetPACPort() string {
	return ac.c.PACPort
}

func (ac *AppConfig) CheckPACPort(port string) error {
	return ac.checkPort(port, &ac.c.PACPort)
}

func (ac *AppConfig) SetPACPort(newPort string) error {
	if err := ac.CheckPACPort(newPort); err != nil {
		return err
	}

	ac.c.PACPort = newPort
	return nil
}

func (ac *AppConfig) SetPACPortMust(newPort string) {
	if err := ac.SetPACPort(newPort); err != nil {
		panic(fmt.Sprintf("SetPACPort error: %v", err))
	}
}

func (ac *AppConfig) Marshal(marshal func(v interface{}) ([]byte, error)) ([]byte, error) {
	return marshal(ac.c)
}

func (ac *AppConfig) check() error {
	if err := ac.checkServers(); err != nil {
		return err
	}

	if !IsValidMode(ac.c.Mode) {
		return fmt.Errorf("invalid mode name '%s'", ac.c.Mode)
	}
	if !common.IsValidPort(ac.c.LocalPort) {
		return fmt.Errorf("invalid local port '%s'", ac.c.LocalPort)
	}
	if !common.IsValidPort(ac.c.APIPort) {
		return fmt.Errorf("invalid control port '%s'", ac.c.APIPort)
	}
	if !common.IsValidPort(ac.c.PACPort) {
		return fmt.Errorf("invalid pac port '%s'", ac.c.PACPort)
	}

	if err := ac.checkRepeatPorts("", nil); err != nil {
		return err
	}

	return nil
}

func (ac *AppConfig) checkServers() error {
	if len(ac.c.Servers) <= 0 {
		return errors.New("can not find server information in config file")
	}

	if ac.c.UsingServer == "" {
		if len(ac.c.Servers) != 1 {
			var srvNames []string
			for name := range ac.c.Servers {
				srvNames = append(srvNames, name)
			}
			return fmt.Errorf("have no idea about which server should be used: [%s]", strings.Join(srvNames, ","))
		}
	} else if _, ok := ac.c.Servers[ac.c.UsingServer]; !ok {
		return fmt.Errorf("can not find server name '%s' in server list", ac.c.UsingServer)
	}

	for _, srv := range ac.c.Servers {
		if err := ac.CheckServerConfig(*srv); err != nil {
			return err
		}
	}

	return nil
}

func (ac *AppConfig) checkPort(port string, except *string) error {
	if !common.IsValidPort(port) {
		return fmt.Errorf("invalid port value '%s'", port)
	}

	return ac.checkRepeatPorts(port, except)
}

func (ac *AppConfig) checkRepeatPorts(port string, except *string) error {
	portMap := make(map[string]string, 3)

	addPortMap := func(p *string, name string) error {
		if reflect.DeepEqual(p, except) {
			return nil
		}
		if repName, ok := portMap[*p]; ok {
			return fmt.Errorf("%s '%s' is same as %s", name, *p, repName)
		}
		portMap[*p] = name
		return nil
	}

	if err := addPortMap(&ac.c.APIPort, "control port"); err != nil {
		return err
	}
	if err := addPortMap(&ac.c.LocalPort, "local port"); err != nil {
		return err
	}
	if err := addPortMap(&ac.c.PACPort, "pac port"); err != nil {
		return err
	}

	if repName, ok := portMap[port]; ok {
		return fmt.Errorf("port '%s' repeat with %s", port, repName)
	}
	return nil
}

func (ac *AppConfig) setServerDefault() {
	if ac.c.UsingServer == "" {
		for name := range ac.c.Servers {
			ac.c.UsingServer = name
			break
		}
	}

	ac.currentServer = ac.c.Servers[ac.c.UsingServer]

	for _, srv := range ac.c.Servers {
		SetServerConfigDefault(srv)
	}
}
