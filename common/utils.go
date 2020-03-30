package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

/*
tempFile touch a template file and return it's full path.
It is the caller's responsibility to remove the file when no longer needed.
*/
func TempFile() (string, error) {
	file, err := ioutil.TempFile("", "ssctrl")
	if err != nil {
		return "", err
	}
	defer file.Close()

	return file.Name(), nil
}

func defaultCfgPath() string {
	return filepath.Join(HomeDir(), ".ssctrl", "config.toml")
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func ExeFile() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	fi, err := os.Lstat(exePath)
	if err != nil {
		return "", err
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return filepath.EvalSymlinks(exePath)
	}
	return exePath, nil
}

func CheckIPV4Format(ip string) error {
	ip = strings.TrimSpace(ip)

	// 15 is length of string like XXX.XXX.XXX.XXX
	if len(ip) > 15 {
		return errors.New(fmt.Sprintf("invalid length of ip v4 address: '%s'", ip))
	}

	for _, c := range ip {
		if unicode.IsSpace(c) {
			return errors.New(fmt.Sprintf("space(s) in ip address: '%s'", ip))
		}
	}

	fields := strings.Split(ip, ".")
	if len(fields) != 4 {
		return errors.New(fmt.Sprintf("invalid format of ip v4 address: '%s'", ip))
	}

	for _, field := range fields {
		if _, err := strconv.ParseUint(field, 10, 8); err != nil {
			return errors.New(fmt.Sprintf("invalid field '%s' in ip v4 address: '%s'", field, ip))
		}
	}

	return nil
}

func IsValidPort(port string) bool {
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return false
	}
	return true
}
