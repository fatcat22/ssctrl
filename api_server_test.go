package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/fatcat22/ssctrl/config"
)

type handlerMock struct {
	enabled bool
	mode    string
	exit    bool

	localPort      string
	pacPort        string
	apiPort        string
	currentSrvName string
	servers        map[string]config.ServerConfig
	autorun        string

	enableProxy         func() error
	disableProxy        func() error
	changeMode          func(string) error
	changeLocalPort     func(string) error
	changePACPort       func(string) error
	changeAPIPort       func(string) error
	changeCurrentServer func(string) error
	updateServers       func(map[string]config.ServerConfig) error
	removeServers       func([]string) error
	autorunFunc         func(bool) error
	exitFunc            func()
	marshalConfig       func(func(v interface{}) ([]byte, error)) ([]byte, error)
}

func (h *handlerMock) EnableProxy() error {
	if h.enableProxy != nil {
		return h.enableProxy()
	}

	h.enabled = true
	return nil
}

func (h *handlerMock) DisableProxy() error {
	if h.disableProxy != nil {
		return h.disableProxy()
	}

	h.enabled = false
	return nil
}

func (h *handlerMock) ChangeMode(newMode string) error {
	if h.changeMode != nil {
		return h.changeMode(newMode)
	}

	h.mode = newMode
	return nil
}

func (h *handlerMock) ChangeLocalPort(port string) error {
	if h.changeLocalPort != nil {
		return h.changeLocalPort(port)
	}

	h.localPort = port
	return nil
}

func (h *handlerMock) ChangePACPort(port string) error {
	if h.changePACPort != nil {
		return h.changePACPort(port)
	}

	h.pacPort = port
	return nil
}

func (h *handlerMock) ChangeAPIPort(port string) error {
	if h.changeAPIPort != nil {
		return h.changeAPIPort(port)
	}

	h.apiPort = port
	return nil
}

func (h *handlerMock) ChangeCurrentServer(newSrvName string) error {
	if h.changeCurrentServer != nil {
		return h.changeCurrentServer(newSrvName)
	}

	h.currentSrvName = newSrvName
	return nil
}

func (h *handlerMock) UpdateServers(srvs map[string]config.ServerConfig) error {
	if h.updateServers != nil {
		return h.updateServers(srvs)
	}

	h.servers = make(map[string]config.ServerConfig)
	for name, srv := range srvs {
		h.servers[name] = srv
	}
	return nil
}

func (h *handlerMock) RemoveServers(names []string) error {
	if h.removeServers != nil {
		return h.removeServers(names)
	}

	if h.servers == nil {
		return nil
	}
	for _, name := range names {
		delete(h.servers, name)
	}
	return nil
}

func (h *handlerMock) Autorun(enable bool) error {
	if h.autorunFunc != nil {
		return h.autorunFunc(enable)
	}

	if enable {
		h.autorun = "yes"
	} else {
		h.autorun = "no"
	}
	return nil
}

func (h *handlerMock) Exit() {
	if h.exitFunc != nil {
		h.exitFunc()
		return
	}

	h.exit = true
}

func (h *handlerMock) MarshalConfig(marshalFunc func(v interface{}) ([]byte, error)) ([]byte, error) {
	if h.marshalConfig != nil {
		return h.marshalConfig(marshalFunc)
	}
	return nil, nil
}

func TestGetConfigSuccess(t *testing.T) {
	const testCfg = "abc123hello"
	h := &handlerMock{
		enabled: true,
		mode:    "mockmode",

		marshalConfig: func(func(interface{}) ([]byte, error)) ([]byte, error) {
			return []byte(testCfg), nil
		},
	}
	const port = "2022"

	srv, err := NewAPIServer(port, h)
	if err != nil {
		t.Fatalf("NewAPIServer error: %v", err)
	}
	srv.Startup()
	defer srv.Shutdown()

	resp, err := http.Get(getCtrlURL(port, "config"))
	if err != nil {
		t.Fatalf("http get config error: %v", err)
	}

	msg, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get config failed. status code: %d. error message: %s", resp.StatusCode, string(msg))
	}

	if string(msg) != testCfg {
		t.Errorf("unexpect config: expect '%s' but got '%s'", testCfg, string(msg))
	}
}

func TestGetConfigError(t *testing.T) {
	h := &handlerMock{
		enabled: true,
		mode:    "mockmode",

		marshalConfig: func(func(interface{}) ([]byte, error)) ([]byte, error) {
			return nil, errors.New("test get config error")
		},
	}
	const port = "2022"

	srv, err := NewAPIServer(port, h)
	if err != nil {
		t.Fatalf("NewAPIServer error: %v", err)
	}
	srv.Startup()
	defer srv.Shutdown()

	resp, err := http.Get(getCtrlURL(port, "config"))
	if err != nil {
		t.Fatalf("http get config error: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("get config success. but we expect error")
	}
}

func TestEanbleProxySuccess(t *testing.T) {
	testPostSuccess(
		"enable",
		"",
		func(h *handlerMock) {
			if h.enabled != true {
				t.Errorf("post 'enable' command success but it does not take effect")
			}
		},
		t,
	)
}

func TestEanbleProxyFailed(t *testing.T) {
	errVal := errors.New("failed test for 'enable'")
	testPostFailed(
		"enable",
		"",
		func(h *handlerMock) error {
			h.enabled = false
			h.enableProxy = func() error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.enabled == true {
				t.Errorf("expect proxy disabled but got enabled")
			}
		},
		t,
	)
}

func TestDisableProxySuccess(t *testing.T) {
	testPostSuccess(
		"disable",
		"",
		func(h *handlerMock) {
			if h.enabled != false {
				t.Errorf("'disable' command does not take effect")
			}
		},
		t,
	)
}

func TestDisableProxyFailed(t *testing.T) {
	errVal := errors.New("failed test for 'disable'")
	testPostFailed(
		"disable",
		"",
		func(h *handlerMock) error {
			h.enabled = true
			h.disableProxy = func() error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.enabled == false {
				t.Errorf("expect proxy enabled but got disabled")
			}
		},
		t,
	)
}

func TestChangeModeSuccess(t *testing.T) {
	testPostSuccess(
		"mode",
		"pac",
		func(h *handlerMock) {
			if h.mode != "pac" {
				t.Errorf("expect mode %s but got %s", "pac", h.mode)
			}
		},
		t,
	)

	testPostSuccess(
		"mode",
		"global",
		func(h *handlerMock) {
			if h.mode != "global" {
				t.Errorf("expect mode %s but got %s", "global", h.mode)
			}
		},
		t,
	)
}

func TestChangeModeFailed(t *testing.T) {
	const expectMode = "abc"
	errVal := errors.New("failed test for 'mode'")

	testPostFailed(
		"mode",
		"xxxmode",
		func(h *handlerMock) error {
			h.mode = expectMode
			return nil
		},
		func(h *handlerMock) {
			if h.mode != expectMode {
				t.Errorf("expect mode %s but got %s", expectMode, h.mode)
			}
		},
		t,
	)

	testPostFailed(
		"mode",
		"pac",
		func(h *handlerMock) error {
			h.mode = expectMode
			h.changeMode = func(string) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.mode != expectMode {
				t.Errorf("expect mode %s but got %s", expectMode, h.mode)
			}
		},
		t,
	)
}

func TestChangeLocalPortSuccess(t *testing.T) {
	const expectPort = "8012"
	testPostSuccess(
		"localPort",
		expectPort,
		func(h *handlerMock) {
			if h.localPort != expectPort {
				t.Errorf("set local port failed: expect '%s' but got '%s'", expectPort, h.localPort)
			}
		},
		t,
	)
}

func TestChangeLocalPortFailed(t *testing.T) {
	const expectLocalPort = "2233"
	errVal := errors.New("failed test for 'localPort'")
	testPostFailed(
		"localPort",
		"9876",
		func(h *handlerMock) error {
			h.localPort = expectLocalPort
			h.changeLocalPort = func(string) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.localPort != expectLocalPort {
				t.Errorf("expect local port %s but got %s", expectLocalPort, h.localPort)
			}
		},
		t,
	)
}

func TestChangePACPortSuccess(t *testing.T) {
	const expectPort = "8012"
	testPostSuccess(
		"pacPort",
		expectPort,
		func(h *handlerMock) {
			if h.pacPort != expectPort {
				t.Errorf("set pac port failed: expect '%s' but got '%s'", expectPort, h.pacPort)
			}
		},
		t,
	)
}

func TestChangePACPortFailed(t *testing.T) {
	const expectPACPort = "2233"
	errVal := errors.New("failed test for 'pacPort'")
	testPostFailed(
		"pacPort",
		"9876",
		func(h *handlerMock) error {
			h.pacPort = expectPACPort
			h.changePACPort = func(string) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.pacPort != expectPACPort {
				t.Errorf("expect pac port %s but got %s", expectPACPort, h.pacPort)
			}
		},
		t,
	)
}

func TestChangeAPIPortSuccess(t *testing.T) {
	const expectPort = "8012"
	testPostSuccess(
		"apiPort",
		expectPort,
		func(h *handlerMock) {
			if h.apiPort != expectPort {
				t.Errorf("set api port failed: expect '%s' but got '%s'", expectPort, h.apiPort)
			}
		},
		t,
	)
}

func TestChangeAPIPortFailed(t *testing.T) {
	const expectAPIPort = "2233"
	errVal := errors.New("failed test for 'apiPort'")
	testPostFailed(
		"apiPort",
		"9876",
		func(h *handlerMock) error {
			h.apiPort = expectAPIPort
			h.changeAPIPort = func(string) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.apiPort != expectAPIPort {
				t.Errorf("expect api port %s but got %s", expectAPIPort, h.apiPort)
			}
		},
		t,
	)
}

func TestChangeCurrentServerSuccess(t *testing.T) {
	const expectName = "testserver"
	testPostSuccess(
		"currentServer",
		expectName,
		func(h *handlerMock) {
			if h.currentSrvName != expectName {
				t.Errorf("set current server failed: expect '%s' but got '%s'", expectName, h.currentSrvName)
			}
		},
		t,
	)
}

func TestChangeCurrentServerFailed(t *testing.T) {
	const expectName = "testserver"
	errVal := errors.New("failed test for 'currentServer'")
	testPostFailed(
		"currentServer",
		"unexpectServerName",
		func(h *handlerMock) error {
			h.currentSrvName = expectName
			h.changeCurrentServer = func(string) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.currentSrvName != expectName {
				t.Errorf("expect current server '%s' but got '%s'", expectName, h.currentSrvName)
			}
		},
		t,
	)
}

func TestUpdateServersSuccess(t *testing.T) {
	expectServers := map[string]config.ServerConfig{
		"srvName1": config.ServerConfig{
			Address:  "11.22.33.44",
			Port:     "1414",
			Crypt:    config.Crypt_AEAD_AES_256_GCM,
			Password: "srv1pwdxm!",
		},
		"srvName2": config.ServerConfig{
			Address:  "81.82.83.84",
			Port:     "8484",
			Password: "srv2pwdxm@",
		},
	}

	data, err := json.Marshal(expectServers)
	if err != nil {
		t.Fatalf("marshal servers error: %v", err)
	}

	testPostSuccess(
		"updateServers",
		string(data),
		func(h *handlerMock) {
			if !reflect.DeepEqual(h.servers, expectServers) {
				t.Errorf("expect servers '%v' but got '%v'", expectServers, h.servers)
			}
		},
		t,
	)
}

func TestUpdateServersFailed(t *testing.T) {
	expectSrv := map[string]config.ServerConfig{
		"srvadf1": config.ServerConfig{
			Address:  "11.22.33.44",
			Port:     "1414",
			Crypt:    config.Crypt_AEAD_AES_256_GCM,
			Password: "srv1pwdxm!",
		},
	}
	data, err := json.Marshal(expectSrv)
	if err != nil {
		t.Fatalf("marshal server data error: %v", err)
	}

	errVal := errors.New("failed test for 'updateServers'")
	testPostFailed(
		"updateServers",
		string(data),
		func(h *handlerMock) error {
			h.servers = make(map[string]config.ServerConfig)
			// do deep copy
			for name, srv := range expectSrv {
				h.servers[name] = srv
			}
			h.updateServers = func(map[string]config.ServerConfig) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if !reflect.DeepEqual(h.servers, expectSrv) {
				t.Errorf("expect server '%v' but got '%v'", expectSrv, h.servers)
				return
			}
		},
		t,
	)
}

func TestRemoveServersSuccess(t *testing.T) {
	const srvName1 = "srvName1"
	const srvName2 = "srvName2"
	const srvName3 = "srvName3"
	rmServersName := []string{srvName1, srvName2}
	const remainName = srvName3
	servers := map[string]config.ServerConfig{
		srvName1: config.ServerConfig{
			Address:  "11.22.33.44",
			Port:     "1414",
			Crypt:    config.Crypt_AEAD_AES_256_GCM,
			Password: "srv1pwdxm!",
		},
		srvName2: config.ServerConfig{
			Address:  "81.82.83.84",
			Port:     "8484",
			Password: "srv2pwdxm@",
		},
		srvName3: config.ServerConfig{
			Address:  "61.62.63.64",
			Port:     "6464",
			Crypt:    config.Crypt_AEAD_AES_128_GCM,
			Password: "srv3pwdxm@",
		},
	}

	postData, err := json.Marshal(rmServersName)
	if err != nil {
		t.Fatalf("marshal %v error: %v", rmServersName, err)
	}

	testPostSuccessWithSetFunc(
		"removeServers",
		string(postData),
		func(h *handlerMock) {
			h.servers = make(map[string]config.ServerConfig)
			for n, s := range servers {
				h.servers[n] = s
			}
		},
		func(h *handlerMock) {
			if len(h.servers) != 1 {
				t.Errorf("expect only %d server config but got %d", 1, len(h.servers))
			}
			srv, ok := h.servers[remainName]
			if !ok {
				t.Errorf("expect server '%s' exist but not find", remainName)
			}
			if srv != servers[remainName] {
				t.Errorf("expect server config '%v' but got '%v'", servers[remainName], srv)
			}
		},
		t,
	)
}

func TestRemoveServersFailed(t *testing.T) {
	expectServers := map[string]config.ServerConfig{
		"srvName1": config.ServerConfig{
			Address:  "11.22.33.44",
			Port:     "1414",
			Crypt:    config.Crypt_AEAD_AES_256_GCM,
			Password: "srv1pwdxm!",
		},
		"srvName2": config.ServerConfig{
			Address:  "81.82.83.84",
			Port:     "8484",
			Password: "srv2pwdxm@",
		},
		"srvName3": config.ServerConfig{
			Address:  "61.62.63.64",
			Port:     "6464",
			Crypt:    config.Crypt_AEAD_AES_128_GCM,
			Password: "srv3pwdxm@",
		},
	}

	errVal := errors.New("failed test for 'removeServers'")
	testPostFailed(
		"removeServers",
		`["srvName1", "srvName2"]`,
		func(h *handlerMock) error {
			h.servers = make(map[string]config.ServerConfig)
			// do deep copy
			for name, srv := range expectServers {
				h.servers[name] = srv
			}
			h.removeServers = func([]string) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if !reflect.DeepEqual(h.servers, expectServers) {
				t.Errorf("expect server '%v' but got '%v'", expectServers, h.servers)
				return
			}
		},
		t,
	)
}

func TestAutorunSuccess(t *testing.T) {
	testPostSuccess(
		"autorun",
		"enable",
		func(h *handlerMock) {
			if h.autorun != "yes" {
				t.Errorf("expect autorun '%s' but get '%s'", "yes", h.autorun)
			}
		},
		t,
	)

	testPostSuccess(
		"autorun",
		"disable",
		func(h *handlerMock) {
			if h.autorun != "no" {
				t.Errorf("expect autorun '%s' but get '%s'", "no", h.autorun)
			}
		},
		t,
	)
}

func TestAutorunFailed(t *testing.T) {
	const expectAutorun = "xxx"
	errVal := errors.New("failed test for 'currentServer'")
	testPostFailed(
		"autorun",
		"enable",
		func(h *handlerMock) error {
			h.autorun = expectAutorun
			h.autorunFunc = func(bool) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.autorun != expectAutorun {
				t.Errorf("expect autorun is '%s' but got '%s'", expectAutorun, h.autorun)
			}
		},
		t,
	)

	testPostFailed(
		"autorun",
		"disable",
		func(h *handlerMock) error {
			h.autorun = expectAutorun
			h.autorunFunc = func(bool) error {
				return errVal
			}
			return errVal
		},
		func(h *handlerMock) {
			if h.autorun != expectAutorun {
				t.Errorf("expect autorun is '%s' but got '%s'", expectAutorun, h.autorun)
			}
		},
		t,
	)

	const invalidArg = "disabcde"
	testPostFailed(
		"autorun",
		invalidArg,
		nil,
		func(h *handlerMock) {
			if len(h.autorun) != 0 {
				t.Errorf("expect autorun is empty but got '%s'", h.autorun)
			}
		},
		t,
	)
}

func getCtrlURL(port, cmd string) string {
	return "http://127.0.0.1:" + port + "/" + cmd
}

func testPostSuccess(cmd, data string, checkFunc func(*handlerMock), t *testing.T) {
	testPostSuccessWithSetFunc(cmd, data, nil, checkFunc, t)
}

func testPostSuccessWithSetFunc(cmd, data string, setFunc func(*handlerMock), checkFunc func(*handlerMock), t *testing.T) {
	const apiPort = "2022"

	h := &handlerMock{
		enabled: true,
		mode:    "mockmode",
	}
	if setFunc != nil {
		setFunc(h)
	}

	srv, err := NewAPIServer(apiPort, h)
	if err != nil {
		t.Fatalf("NewAPIServer error: %v", err)
	}
	srv.Startup()
	defer srv.Shutdown()

	resp, err := http.Post(getCtrlURL(apiPort, cmd), "application/json", strings.NewReader(data))
	if err != nil {
		t.Fatalf("http post '%s' error: %v", cmd, err)
	}

	msg, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("post '%s' failed. status code: %d. error message: %s", cmd, resp.StatusCode, string(msg))
	}

	checkFunc(h)
}

func testPostFailed(cmd, data string, setFunc func(*handlerMock) error, checkFunc func(*handlerMock), t *testing.T) {
	const apiPort = "2022"

	h := &handlerMock{
		enabled: false,
		mode:    "mockmode",
	}
	var errMsg error
	if setFunc != nil {
		errMsg = setFunc(h)
	}

	srv, err := NewAPIServer(apiPort, h)
	if err != nil {
		t.Fatalf("NewAPIServer error: %v", err)
	}
	srv.Startup()
	defer srv.Shutdown()

	resp, err := http.Post(getCtrlURL(apiPort, cmd), "application/json", strings.NewReader(data))
	if err != nil {
		t.Fatalf("http post '%s' error: %v", cmd, err)
	}

	msg, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("post '%s' success, but we expect failed", cmd)
	}
	if errMsg != nil && string(msg) != errMsg.Error() {
		t.Errorf("expect message '%s' but got '%s'", errMsg, string(msg))
	}
	checkFunc(h)
}
