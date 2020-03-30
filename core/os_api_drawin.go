// +xxxbuild drawin

package core

import (
	"os/exec"
	"strings"
)

type drawinAPI struct {
}

func (api *drawinAPI) listAllNetwork() ([]string, error) {
	listAllOutput, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return nil, err
	}

	var results []string
	outputSlice := strings.Split(string(listAllOutput), "\n")
	for _, s := range outputSlice[1:] {
		if len(s) <= 0 {
			continue
		}

		results = append(results, s)
	}

	return results, nil
}

func (api *drawinAPI) setAutoProxyFor(network, url string) error {
	cmd := exec.Command("networksetup",
		"-setautoproxyurl", network, url)
	return cmd.Run()
}

func (api *drawinAPI) setGlobalProxyFor(network, addr, port string) error {
	cmd := exec.Command("networksetup",
		"-setsocksfirewallproxy", network, addr, port)

	return cmd.Run()
}

func (api *drawinAPI) setProxyBypassDomains(network, bypass string) error {
	cmd := exec.Command("networksetup",
		"-setproxybypassdomains", network, bypass)

	return cmd.Run()
}

func (api *drawinAPI) clearAutoProxyFor(network string) error {
	cmd1 := exec.Command("networksetup",
		"-setautoproxyurl", network, " ")
	err1 := cmd1.Run()

	cmd2 := exec.Command("networksetup",
		"-setautoproxystate", network, "off")
	err2 := cmd2.Run()

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	} else {
		return nil
	}
}

func (api *drawinAPI) clearGlobalProxyFor(network string) error {
	cmd1 := exec.Command("networksetup",
		"-setsocksfirewallproxy", network, "Empty")
	err1 := cmd1.Run()

	cmd2 := exec.Command("networksetup",
		"-setsocksfirewallproxystate", network, "off")
	err2 := cmd2.Run()

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	} else {
		return nil
	}
}
