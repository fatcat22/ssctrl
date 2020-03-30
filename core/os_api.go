package core

import (
	"runtime"
)

type osAPI interface {
	listAllNetwork() ([]string, error)
	setAutoProxyFor(network, url string) error
	setGlobalProxyFor(network, addr, port string) error
	setProxyBypassDomains(network, bypass string) error
	clearAutoProxyFor(network string) error
	clearGlobalProxyFor(network string) error
}

var api osAPI

func init() {
	api = newOSAPI()
}

func newOSAPI() osAPI {
	if isOnTest {
		return newMockAPI()
	}

	switch runtime.GOOS {
	case "darwin":
		return new(drawinAPI)
	case "windows":
		return new(notImpAPI)
	case "linux":
		return new(notImpAPI)
	default:
		return new(notImpAPI)
	}
}
