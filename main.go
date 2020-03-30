package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatcat22/ssctrl/common"
	"github.com/fatcat22/ssctrl/config"
	"github.com/fatcat22/ssctrl/core"
)

func main() {
	defaultCfgFile := filepath.Join(common.HomeDir(), ".ssctrl", "config.toml")
	cfg, err := config.LoadConfig(defaultCfgFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	_, srvCfg := cfg.GetCurrentServerConfig()
	core, err := core.NewProxyCore(cfg.GetPACPort(), "", cfg.GetLocalPort(), cfg.GetMode(), srvCfg, "")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	svc, err := NewSSService()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	controler, err := NewControler(cfg, core, svc)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err := controler.Startup(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer func() {
		controler.Shutdown()
		config.RestoreConfig(cfg, defaultCfgFile)
	}()

	svc.Wait()
}
