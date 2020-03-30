package core

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatcat22/ssctrl/common"
)

const (
	pacURLFile = "proxy.pac"
)

var (
	DefaultPACLocalPath = filepath.Join(common.HomeDir(), ".ssctrl", "gfwlist.js")
)

type PACServer struct {
	server *http.Server

	pacPort string
	pacData []byte

	localAddr string
	localPort string

	isStartup  bool
	shutdownCh chan struct{}
}

func NewPACServer(port, localPACFile string, localAddr, localPort string) (*PACServer, error) {
	if localPACFile == "" {
		localPACFile = DefaultPACLocalPath
	}

	pacData, err := ioutil.ReadFile(localPACFile)
	if err != nil {
		return nil, err
	}

	return &PACServer{
		pacPort: port,
		pacData: pacData,

		localAddr: localAddr,
		localPort: localPort,

		isStartup: false,
	}, nil
}

func (ps *PACServer) Startup() error {
	if ps.isStartup {
		return nil
	}

	ps.renewServer()
	ps.shutdownCh = make(chan struct{})

	var resultErr error
	go func() {
		defer func() {
			ps.isStartup = false
			close(ps.shutdownCh)
		}()

		resultErr = ps.server.ListenAndServe()
		if resultErr != http.ErrServerClosed {
			log.Printf("pac server start error: %v\n", resultErr)
		}
	}()

	t := time.NewTicker(time.Second)
	defer t.Stop()
	select {
	case <-t.C:
		ps.isStartup = true
		return nil
	case <-ps.shutdownCh:
		return resultErr
	}
}

func (ps *PACServer) Shutdown() error {
	if !ps.isStartup {
		return nil
	}

	if err := ps.server.Shutdown(context.Background()); err != nil {
		return err
	}

	<-ps.shutdownCh
	ps.server = nil
	ps.isStartup = false
	return nil
}

func (ps *PACServer) GetPACURL() string {
	url := url.URL{
		Scheme: "http",
		Host:   ps.serverAddr(),
		Path:   pacURLFile,
	}
	return url.String()
}

func (ps *PACServer) ReloadPAC(pacFile string) error {
	pacData, err := ioutil.ReadFile(pacFile)
	if err != nil {
		return err
	}

	ps.pacData = pacData
	return nil
}

func (ps *PACServer) ChangeLocalPort(newPort string) (result error) {
	if newPort == ps.localPort {
		return nil
	}
	defer func() {
		if result == nil {
			ps.localPort = newPort
		}
	}()

	//TODO
	return nil
}

func (ps *PACServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqPath := strings.TrimLeft(req.URL.Path, "/")

	if reqPath != pacURLFile {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Write(ps.pacData)
}

func (ps *PACServer) renewServer() {
	if ps.server != nil {
		panic("pac server is not nil")
	}

	ps.server = &http.Server{
		Addr:    ps.serverAddr(),
		Handler: ps,
	}
}

func (ps *PACServer) serverAddr() string {
	return "127.0.0.1:" + ps.pacPort
}
