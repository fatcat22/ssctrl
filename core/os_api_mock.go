package core

import "fmt"

const (
	modePAC    = 0x01
	modeGlobal = 0x02
)

type mockSettings struct {
	mode       int
	pacURL     string
	globalAddr string
	globalPort string
	bypassDom  string
}

type mockAPI struct {
	settings map[string]*mockSettings
}

func newMockAPI() *mockAPI {
	mapi := &mockAPI{
		settings: make(map[string]*mockSettings),
	}

	nets := []string{
		"WiFi",
		"wlan",
	}
	for _, n := range nets {
		mapi.settings[n] = &mockSettings{}
	}

	return mapi
}

func (m *mockAPI) listAllNetwork() ([]string, error) {
	var nets []string
	for name := range m.settings {
		nets = append(nets, name)
	}
	return nets, nil
}

func (m *mockAPI) setAutoProxyFor(network, url string) error {
	settings, ok := m.settings[network]
	if !ok {
		return fmt.Errorf("unknown network '%s'", network)
	}

	settings.mode |= modePAC
	settings.pacURL = url

	return nil
}

func (m *mockAPI) setGlobalProxyFor(network, addr, port string) error {
	settings, ok := m.settings[network]
	if !ok {
		return fmt.Errorf("unknown network '%s'", network)
	}

	settings.mode |= modeGlobal
	settings.globalAddr = addr
	settings.globalPort = port

	return nil
}

func (m *mockAPI) setProxyBypassDomains(network, bypass string) error {
	settings, ok := m.settings[network]
	if !ok {
		return fmt.Errorf("unknown network '%s'", network)
	}

	settings.bypassDom = bypass

	return nil
}

func (m *mockAPI) clearAutoProxyFor(network string) error {
	settings, ok := m.settings[network]
	if !ok {
		return fmt.Errorf("unknown network '%s'", network)
	}

	settings.mode &= ^modePAC
	settings.pacURL = ""

	return nil
}

func (m *mockAPI) clearGlobalProxyFor(network string) error {
	settings, ok := m.settings[network]
	if !ok {
		return fmt.Errorf("unknown network '%s'", network)
	}

	settings.mode &= ^modeGlobal
	settings.globalAddr = ""
	settings.globalPort = ""

	return nil
}

func resetMockAPI() *mockAPI {
	api = newMockAPI()
	return api.(*mockAPI)
}
