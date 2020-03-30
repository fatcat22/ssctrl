package core

import (
	"fmt"
	"runtime"
)

type notImpAPI struct{}

func (api *notImpAPI) listAllNetwork() ([]string, error) {
	return nil, fmt.Errorf("osAPI is not implemented on %s", runtime.GOOS)
}

func (api *notImpAPI) setAutoProxyFor(network, url string) error {
	return fmt.Errorf("osAPI is not implemented on %s", runtime.GOOS)
}

func (api *notImpAPI) setGlobalProxyFor(network, addr, port string) error {
	return fmt.Errorf("osAPI is not implemented on %s", runtime.GOOS)
}

func (api *notImpAPI) setProxyBypassDomains(network, bypass string) error {
	return fmt.Errorf("osAPI is not implemented on %s", runtime.GOOS)
}

func (api *notImpAPI) clearAutoProxyFor(network string) error {
	return fmt.Errorf("osAPI is not implemented on %s", runtime.GOOS)
}

func (api *notImpAPI) clearGlobalProxyFor(network string) error {
	return fmt.Errorf("osAPI is not implemented on %s", runtime.GOOS)
}
