package updater

import (
	"io/ioutil"
	"os"
)

const (
	defaultInterpreter    = "/bin/bash"
	defaultInterpreterArg = "-c"
	defaultScriptSuffix   = ".sh"
	defaultWorkDir        = "/tmp"
)

func getVmuuid() string {
	b, err := ioutil.ReadFile("/opt/cloud/common/vmuuid")
	if err != nil {
		return ""
	}
	return string(b)
}

func getHostName() string {
	b, err := ioutil.ReadFile("/opt/cloud/common/hostname")
	if err != nil {
		if hostname, err := os.Hostname(); err == nil {
			return hostname
		}
		return ""
	}
	return string(b)
}
