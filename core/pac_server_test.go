package core

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/fatcat22/ssctrl/common"
)

func TestGetPACURL(t *testing.T) {
	const port = "2022"

	srv, _ := createPACServer(t, port)
	srv.Startup()
	defer srv.Shutdown()

	if srv.GetPACURL() != getPACURL(port) {
		t.Fatalf("GetPACURL error: expect '%s' but got '%s'", getPACURL(port), srv.GetPACURL())
	}
}

func TestPACServer(t *testing.T) {
	const port = "1033"

	srv, expectData := createPACServer(t, port)
	srv.Startup()
	defer srv.Shutdown()

	checkGetPAC(t, expectData, port)
}

func TestReloadPAC(t *testing.T) {
	const port = "1034"

	srv, _ := createPACServer(t, port)
	srv.Startup()
	defer srv.Shutdown()

	const newPACData = "!@#$%^&*(newpac"
	newPACFile, err := common.TempFile()
	if err != nil {
		t.Fatalf("create new template pac file error: %v", err)
	}
	defer os.Remove(newPACFile)
	if err := ioutil.WriteFile(newPACFile, []byte(newPACData), os.ModePerm); err != nil {
		t.Fatalf("write new pac file error: %v", err)
	}

	if err := srv.ReloadPAC(newPACFile); err != nil {
		t.Fatalf("reload pac file error: %v", err)
	}

	checkGetPAC(t, newPACData, port)
}

func TestPACServerStartupShutdownRepeatedly(t *testing.T) {
	srv, _ := createPACServer(t, "1034")

	if err := srv.Startup(); err != nil {
		t.Fatalf("server startup error: %v", err)
	}
	if err := srv.Shutdown(); err != nil {
		t.Fatalf("server shutdown error: %v", err)
	}

	if err := srv.Startup(); err != nil {
		t.Fatalf("2sd of server startup error: %v", err)
	}
	if err := srv.Shutdown(); err != nil {
		t.Fatalf("2sd of server shutdown error: %v", err)
	}
}

func createPACServer(t *testing.T, pacPort string) (*PACServer, string) {
	tmpPAC, err := common.TempFile()
	if err != nil {
		t.Fatalf("create tempalte pac error: %v", err)
	}
	defer os.Remove(tmpPAC)

	const mockPAC = "hello, I am pac string"
	if err := ioutil.WriteFile(tmpPAC, []byte(mockPAC), os.ModePerm); err != nil {
		t.Fatalf("write mock pac file error: %v", err)
	}

	srv, err := NewPACServer(pacPort, tmpPAC, "11.22.33.44", "4321")
	if err != nil {
		t.Fatalf("create pac server error: %v", err)
	}

	return srv, mockPAC
}

func checkGetPAC(t *testing.T, expectData string, pacPort string) {
	resp, err := http.Get(getPACURL(pacPort))
	if err != nil {
		t.Fatalf("get proxy.pac error: %v", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read proxy.pac body error: %v", err)
	}

	if string(data) != expectData {
		t.Fatalf("http get pac data failed: expect '%s' but got '%s'", expectData, data)
	}
}

func getPACURL(port string) string {
	return "http://127.0.0.1:" + port + "/" + pacURLFile
}
