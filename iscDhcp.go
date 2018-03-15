package iscDhcp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/go-ps"
	"github.com/xaionaro-go/iscDhcp/cfg"
	"os"
	"os/exec"
)

const (
	CFG_PATH = "/etc/dhcp/dhcpd-dynamic.conf"
)

const (
	RUNNING = Status(1)
	STOPPED = Status(2)
)

var (
	ErrCannotRun      = errors.New("cannot run")
	ErrAlreadyRunning = errors.New("dhcpd is already started")
)

type Status int

type DHCP struct {
	Config *cfg.Config
}

func NewDHCP() *DHCP {
	return &DHCP{Config: cfg.NewConfig()}
}

func (dhcp *DHCP) ReloadConfig() error {
	return dhcp.Config.LoadFrom(CFG_PATH)
}
func (dhcp DHCP) SaveConfig() error {
	return dhcp.Config.ConfigWriteTo(CFG_PATH)
}
func (dhcp DHCP) findProcess() *os.Process {
	processes, err := ps.Processes()
	if err != nil {
		panic(err)
	}
	for _, process := range processes {
		if process.Executable() == "dhcpd" {
			osProcess, err := os.FindProcess(process.Pid())
			if err != nil {
				panic(err)
			}
			return osProcess
		}
	}
	return nil
}
func (dhcp DHCP) Status() Status {
	process := dhcp.findProcess()
	if process != nil {
		return RUNNING
	}
	return STOPPED
}
func (dhcp DHCP) StartProcess() (err error) {
	/*if dhcp.Status() != STOPPED {
		return ErrAlreadyRunning
	}*/
	cmd := exec.Command("service", "isc-dhcp-server", "start") // Works only with Debian/Ubuntu
	var outputBuf bytes.Buffer
	outputBufWriter := bufio.NewWriter(&outputBuf)
	cmd.Stdout = outputBufWriter
	cmd.Stderr = outputBufWriter
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v\noutput: %v", err.Error(), outputBuf.String())
		return err
	}
	if dhcp.Status() != RUNNING {
		return ErrCannotRun
	}
	return nil
}
func (dhcp DHCP) Start() (err error) {
	err = dhcp.SaveConfig()
	if err != nil {
		return err
	}
	return dhcp.StartProcess()
}
func (dhcp DHCP) StopProcess() error {
	process := dhcp.findProcess()
	if process == nil {
		return nil
	}
	return process.Kill()
}
func (dhcp DHCP) Stop() error {
	return dhcp.StopProcess()
}
func (dhcp DHCP) Restart() error {
	dhcp.Stop()
	return dhcp.Start()
}

