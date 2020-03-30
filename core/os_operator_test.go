package core

import (
	"testing"

	"github.com/fatcat22/ssctrl/config"
)

func init() {
	isOnTest = true
	api = newOSAPI()
}

func TestStartupPAC(t *testing.T) {
	const expectMode = config.ModePAC
	const expectURL = "http://127.0.0.1:2222/proxy.pac"
	mapi := resetMockAPI()
	expectNet, _ := mapi.listAllNetwork()

	op, err := NewOSOperator(expectMode, expectURL, "", "")
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}

	if op.isStartup != true {
		t.Errorf("OSOperator startup success but OSOperator.isStartup == %v", op.isStartup)
	}

	if len(expectNet) != len(mapi.settings) {
		nets := make([]string, 0)
		for n := range mapi.settings {
			nets = append(nets, n)
		}
		t.Errorf("OSOperator startup success but not all network is set. expect networks '%v' but got '%v'", expectNet, nets)
	}

	for _, network := range expectNet {
		nc, ok := mapi.settings[network]
		if !ok {
			t.Errorf("proxy of network '%s' is net set", network)
			continue
		}

		if nc.mode != modePAC {
			t.Errorf("expect pac mode but got mode value '0x%x'", nc.mode)
		}
		if nc.pacURL != expectURL {
			t.Errorf("expect pac url '%s' but got '%s'", expectURL, nc.pacURL)
		}
	}
}

func TestStartupGlobal(t *testing.T) {
	const expectMode = config.ModeGlobal
	const expectAddr = "11.22.99.88"
	const expectPort = "1234"
	mapi := resetMockAPI()
	expectNet, _ := mapi.listAllNetwork()

	op, err := NewOSOperator(expectMode, "abcd", expectAddr, expectPort)
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}

	if op.isStartup != true {
		t.Errorf("OSOperator startup success but OSOperator.isStartup == %v", op.isStartup)
	}

	if len(expectNet) != len(mapi.settings) {
		nets := make([]string, 0)
		for n := range mapi.settings {
			nets = append(nets, n)
		}
		t.Errorf("OSOperator startup success but not all network is set. expect networks '%v' but got '%v'", expectNet, nets)
	}

	for _, network := range expectNet {
		nc, ok := mapi.settings[network]
		if !ok {
			t.Errorf("proxy of network '%s' is net set", network)
			continue
		}

		if nc.mode != modeGlobal {
			t.Errorf("expect pac mode but got mode value '0x%x'", nc.mode)
		}
		if nc.globalAddr != expectAddr {
			t.Errorf("expect global address '%s' but got '%s'", expectAddr, nc.globalAddr)
		}
		if nc.globalPort != expectPort {
			t.Errorf("expect global port '%s' but got '%s'", expectPort, nc.globalPort)
		}
	}
}

func TestShutdownPAC(t *testing.T) {
	mapi := resetMockAPI()

	op, err := NewOSOperator(config.ModePAC, "abcd", "11.22.33.44", "1234")
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}

	if err := op.Shutdown(); err != nil {
		t.Fatalf("shutdown OSOperator error: %v", err)
	}

	if op.isStartup {
		t.Errorf("OSOperator shutdown success but OSOperator.isStartup == %v", op.isStartup)
	}
	for _, nc := range mapi.settings {
		if nc.mode != 0 {
			t.Errorf("OSOperator is shutdown but mode value is '%d'", nc.mode)
		}
		if nc.pacURL != "" {
			t.Errorf("OSOperator is shutdown but pac url is '%s'", nc.pacURL)
		}
	}
}

func TestShutdownGlobal(t *testing.T) {
	mapi := resetMockAPI()

	op, err := NewOSOperator(config.ModeGlobal, "abcd", "11.22.33.44", "1234")
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}

	if err := op.Shutdown(); err != nil {
		t.Fatalf("shutdown OSOperator error: %v", err)
	}

	if op.isStartup {
		t.Errorf("OSOperator shutdown success but OSOperator.isStartup == %v", op.isStartup)
	}
	for _, nc := range mapi.settings {
		if nc.mode != 0 {
			t.Errorf("OSOperator is shutdown but mode value is '%d'", nc.mode)
		}
		if nc.globalAddr != "" {
			t.Errorf("OSOperator is shutdown but global address is '%s'", nc.globalAddr)
		}
		if nc.globalPort != "" {
			t.Errorf("OSOperator is shutdown but global port is '%s'", nc.globalPort)
		}
	}
}

func TestChangeMode(t *testing.T) {
	const expectMode = modeGlobal
	const expectAddr = "11.22.33.44"
	const expectPort = "1234"
	mapi := resetMockAPI()

	op, err := NewOSOperator(config.ModePAC, "abcd", expectAddr, expectPort)
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}

	// test ChangeMode befer Startup
	if err := op.ChangeMode(config.ModeGlobal); err != nil {
		t.Fatalf("OSOperator.ChangeMode error: %v", err)
	}
	for _, s := range mapi.settings {
		if s.mode != 0 {
			t.Errorf("OSOperator isn't startup but change mode take effect")
		}
	}

	op, err = NewOSOperator(config.ModePAC, "abcd", expectAddr, expectPort)
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}

	for name, s := range mapi.settings {
		if s.mode != modePAC {
			t.Errorf("network %s: expect mode pac but got %d", name, s.mode)
		}
	}

	if err := op.ChangeMode(config.ModeGlobal); err != nil {
		t.Fatalf("OSOperator.ChangeMode error: %v", err)
	}

	for name, s := range mapi.settings {
		if s.mode != modeGlobal {
			t.Errorf("network %s: expect mode global but got %d", name, s.mode)
		}
		if s.globalAddr != expectAddr {
			t.Errorf("network %s: expect address %s but got %s", name, expectAddr, s.globalAddr)
		}
		if s.globalPort != expectPort {
			t.Errorf("network %s: expect port %s but got %s", name, expectPort, s.globalPort)
		}
	}
}

func TestChangeLocalPort(t *testing.T) {
	const expectAddr = "11.22.33.44"
	const oldPort = "5432"
	const expectPort = "1234"

	// test ChangeLocalPort befer Startup
	mapi := resetMockAPI()
	op, err := NewOSOperator(config.ModeGlobal, "abcd", expectAddr, oldPort)
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.ChangeLocalPort(expectPort); err != nil {
		t.Fatalf("OSOperator.ChangeLocalPort error: %v", err)
	}

	for _, s := range mapi.settings {
		if s.globalPort != "" {
			t.Errorf("OSOperator isn't startup but change local port take effect")
		}
	}

	// test ChangeLocalPort after startup but with pac mode
	mapi = resetMockAPI()
	op, err = NewOSOperator(config.ModePAC, "abcd", expectAddr, oldPort)
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}
	if err := op.ChangeLocalPort(expectPort); err != nil {
		t.Fatalf("OSOperator.ChangeLocalPort error: %v", err)
	}

	for _, s := range mapi.settings {
		if s.globalPort != "" {
			t.Errorf("current mode is pac but ChangeLocalPort take effect, expect empty but got %s", s.globalPort)
		}
	}

	// test ChangeLocalPort after startup and with global mode
	mapi = resetMockAPI()
	op, err = NewOSOperator(config.ModeGlobal, "abcd", expectAddr, oldPort)
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}
	if err := op.ChangeLocalPort(expectPort); err != nil {
		t.Fatalf("OSOperator.ChangeLocalPort error: %v", err)
	}

	for _, s := range mapi.settings {
		if s.globalPort != expectPort {
			t.Errorf("invalid local port after ChangeLocalPort: expect %s but got %s", expectPort, s.globalPort)
		}
	}
}

func TestChangePACURL(t *testing.T) {
	const expectURL = "http://12344/pac.pac"

	// test ChangePACURL befer Startup
	mapi := resetMockAPI()
	op, err := NewOSOperator(config.ModePAC, "abcd", "11.22.33.44", "1234")
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.ChangePACURL(expectURL); err != nil {
		t.Fatalf("OSOperator.ChangePACURL error: %v", err)
	}

	for _, s := range mapi.settings {
		if s.pacURL != "" {
			t.Errorf("OSOperator isn't startup but change pac url take effect")
		}
	}

	// test ChangePACURL after startup but with global mode
	mapi = resetMockAPI()
	op, err = NewOSOperator(config.ModeGlobal, "abcd", "11.22.33.44", "1234")
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}
	if err := op.ChangePACURL(expectURL); err != nil {
		t.Fatalf("OSOperator.ChangePACURL error: %v", err)
	}

	for _, s := range mapi.settings {
		if s.pacURL != "" {
			t.Errorf("current mode is global but ChangePACURL take effect, expect empty but got %s", s.pacURL)
		}
	}

	// test ChangePACURL after startup and with pac mode
	mapi = resetMockAPI()
	op, err = NewOSOperator(config.ModePAC, "abcd", "11.22.33.44", "1234")
	if err != nil {
		t.Fatalf("new OSOperator error: %v", err)
	}
	if err := op.Startup(); err != nil {
		t.Fatalf("start OSOperator error: %v", err)
	}
	if err := op.ChangePACURL(expectURL); err != nil {
		t.Fatalf("OSOperator.ChangePACURL error: %v", err)
	}

	for _, s := range mapi.settings {
		if s.pacURL != expectURL {
			t.Errorf("invalid pac url after ChangePACURL: expect %s but got %s", expectURL, s.pacURL)
		}
	}
}
