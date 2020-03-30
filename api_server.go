package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/fatcat22/ssctrl/config"
)

type Handler interface {
	EnableProxy() error
	DisableProxy() error
	ChangeMode(string) error
	ChangeLocalPort(port string) error
	ChangePACPort(port string) error
	ChangeAPIPort(port string) error
	ChangeCurrentServer(newSrvName string) error
	UpdateServers(map[string]config.ServerConfig) error
	RemoveServers(names []string) error
	Autorun(bool) error

	Exit()
	MarshalConfig(func(v interface{}) ([]byte, error)) ([]byte, error)
}

type handleFunc func(w http.ResponseWriter, req *http.Request)

type apiServer struct {
	server  *http.Server
	apiPort string

	ctrlHandler Handler

	getRoute  map[string]handleFunc
	postRoute map[string]handleFunc

	isStartup  bool
	shutdownCh chan struct{}
}

func NewAPIServer(port string, h Handler) (*apiServer, error) {
	ctrlSrv := &apiServer{
		apiPort: port,

		ctrlHandler: h,

		isStartup: false,
	}

	ctrlSrv.setRoute()
	return ctrlSrv, nil
}

func (as *apiServer) Startup() error {
	if as.isStartup {
		return nil
	}

	as.renewServer()
	as.shutdownCh = make(chan struct{})

	var resultErr error
	go func() {
		defer func() {
			as.isStartup = false
			close(as.shutdownCh)
		}()
		resultErr = as.server.ListenAndServe()
		if resultErr != http.ErrServerClosed {
			log.Printf("api server start error: %v\n", resultErr)
		}
	}()

	t := time.NewTicker(time.Second)
	defer t.Stop()
	select {
	case <-t.C:
		as.isStartup = true
		return nil
	case <-as.shutdownCh:
		return resultErr
	}
}

func (as *apiServer) Shutdown() error {
	if !as.isStartup {
		return nil
	}

	as.server.Shutdown(context.Background())

	<-as.shutdownCh
	as.server = nil
	as.isStartup = false
	return nil
}

func (as *apiServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var routeTable map[string]handleFunc

	switch req.Method {
	case "GET":
		routeTable = as.getRoute
	case "POST":
		routeTable = as.postRoute
	default:
		w.Write([]byte(fmt.Sprintf("unsupport request '%s'", req.Method)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, ok := routeTable[req.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f(w, req)
}

func (as *apiServer) renewServer() {
	if as.server != nil {
		panic("api server is not nil")
	}

	as.server = &http.Server{
		Addr:    "127.0.0.1:" + as.apiPort,
		Handler: as,
	}
}

func (as *apiServer) setRoute() {
	as.getRoute = map[string]handleFunc{
		"/config": as.handleGetConfig,
	}

	as.postRoute = map[string]handleFunc{
		"/enable":        as.handleEnableProxy,
		"/disable":       as.handleDisableProxy,
		"/exit":          as.handleExit,
		"/mode":          as.handleChangeMode,
		"/localPort":     as.handleChangeLocalPort,
		"/pacPort":       as.handleChangePACPort,
		"/apiPort":       as.handleChangeAPIPort,
		"/currentServer": as.handleChangeCurrentServer,
		"/updateServers": as.handleUpdateServers,
		"/removeServers": as.handleRemoveServers,
		"/autorun":       as.handleAutorun,
	}
}

func (as *apiServer) handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	data, err := as.ctrlHandler.MarshalConfig(json.Marshal)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("marshal config error"))
		//TODO: log
		return
	}

	w.Write(data)
}

func (as *apiServer) handleEnableProxy(w http.ResponseWriter, _ *http.Request) {
	as.handleReq(
		w,
		nil,
		false,
		nil,
		func(string) error { return as.ctrlHandler.EnableProxy() },
	)
}

func (as *apiServer) handleDisableProxy(w http.ResponseWriter, _ *http.Request) {
	as.handleReq(
		w,
		nil,
		false,
		nil,
		func(string) error { return as.ctrlHandler.DisableProxy() },
	)
}

func (as *apiServer) handleExit(w http.ResponseWriter, _ *http.Request) {
	as.ctrlHandler.Exit()
}

func (as *apiServer) handleChangeMode(w http.ResponseWriter, req *http.Request) {
	as.handleReq(
		w,
		req,
		true,
		func(mode string) error {
			if !config.IsValidMode(mode) {
				return fmt.Errorf("unknown mode name '%s'", mode)
			}
			return nil
		},
		as.ctrlHandler.ChangeMode,
	)
}

func (as *apiServer) handleChangeLocalPort(w http.ResponseWriter, req *http.Request) {
	as.handleReq(
		w,
		req,
		true,
		nil,
		as.ctrlHandler.ChangeLocalPort,
	)
}

func (as *apiServer) handleChangePACPort(w http.ResponseWriter, req *http.Request) {
	as.handleReq(
		w,
		req,
		true,
		nil,
		as.ctrlHandler.ChangePACPort,
	)
}

func (as *apiServer) handleChangeAPIPort(w http.ResponseWriter, req *http.Request) {
	as.handleReq(
		w,
		req,
		true,
		nil,
		as.ctrlHandler.ChangeAPIPort,
	)
}

func (as *apiServer) handleChangeCurrentServer(w http.ResponseWriter, req *http.Request) {
	as.handleReq(
		w,
		req,
		true,
		nil,
		as.ctrlHandler.ChangeCurrentServer,
	)
}

func (as *apiServer) handleUpdateServers(w http.ResponseWriter, req *http.Request) {
	updateServers := func(data string) error {
		servers := make(map[string]config.ServerConfig)

		if err := json.Unmarshal([]byte(data), &servers); err != nil {
			return err
		}
		return as.ctrlHandler.UpdateServers(servers)
	}

	as.handleReq(
		w,
		req,
		true,
		nil,
		updateServers,
	)
}

func (as *apiServer) handleRemoveServers(w http.ResponseWriter, req *http.Request) {
	removeServers := func(data string) error {
		var serversName []string

		if err := json.Unmarshal([]byte(data), &serversName); err != nil {
			return err
		}
		return as.ctrlHandler.RemoveServers(serversName)
	}

	as.handleReq(
		w,
		req,
		true,
		nil,
		removeServers,
	)
}

func (as *apiServer) handleAutorun(w http.ResponseWriter, req *http.Request) {
	const enableArg = "enable"
	const disableArg = "disable"

	as.handleReq(
		w,
		req,
		true,
		func(arg string) error {
			if arg != enableArg && arg != disableArg {
				return fmt.Errorf("unknown argument '%s'", arg)
			}
			return nil
		},
		func(arg string) error {
			var enable bool
			switch arg {
			case enableArg:
				enable = true
			case disableArg:
				enable = false
			default:
				return fmt.Errorf("unknown argument '%s'", arg)
			}
			return as.ctrlHandler.Autorun(enable)
		},
	)
}

func (as *apiServer) handleReq(
	w http.ResponseWriter,
	req *http.Request,
	hasArg bool,
	checkArg func(string) error,
	handle func(string) error) {

	var arg string
	if hasArg {
		argBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		arg = string(argBytes)
		if checkArg != nil {
			if err := checkArg(arg); err != nil {
				w.WriteHeader(http.StatusNotAcceptable)
				w.Write([]byte(fmt.Sprintf("%v.(%v)", err, argBytes)))
				return
			}
		}
	}

	if err := handle(arg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
