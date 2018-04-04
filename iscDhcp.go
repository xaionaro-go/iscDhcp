package iscDhcp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/go-ps"
	"github.com/xaionaro-go/iscDhcp/cfg"
	"sync"
	"time"
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
	ErrCannotStop     = errors.New("cannot stop dhcpd")
)

type Status int

type DHCP struct {
	Config *cfg.Config
	mutex  *sync.Mutex
}

func NewDHCP() *DHCP {
	return &DHCP{Config: cfg.NewConfig(), mutex: &sync.Mutex{}}
}

func (dhcp *DHCP) ReloadConfig() error {
	dhcp.lock()
	defer dhcp.unlock()
	return dhcp.reloadConfig()
}
func (dhcp *DHCP) reloadConfig() error {
	return dhcp.Config.LoadFrom(CFG_PATH)
}
func (dhcp DHCP) SaveConfig() error {
	dhcp.lock()
	defer dhcp.unlock()
	return dhcp.saveConfig()
}
func (dhcp DHCP) saveConfig() error {
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
	dhcp.lock()
	defer dhcp.unlock()
	return dhcp.status()
}
func (dhcp DHCP) status() Status {
	process := dhcp.findProcess()
	if process != nil {
		return RUNNING
	}
	return STOPPED
}
func (dhcp DHCP) startProcess() (err error) {
	/*if dhcp.status() != STOPPED {
		return ErrAlreadyRunning
	}*/
	process := dhcp.findProcess()
	if process != nil {
		return ErrAlreadyRunning
	}
	os.Remove("/var/run/dhcpd.pid")
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
	if dhcp.status() != RUNNING {
		return ErrCannotRun
	}
	return nil
}
func (dhcp DHCP) Start() (err error) {
	dhcp.lock()
	defer dhcp.unlock()
	return dhcp.start()
}
func (dhcp DHCP) start() (err error) {
	err = dhcp.saveConfig()
	if err != nil {
		return err
	}
	return dhcp.startProcess()
}
func (dhcp DHCP) stopProcess() error {
	cmd := exec.Command("service", "isc-dhcp-server", "stop") // Works only with Debian/Ubuntu
	cmd.Run()
	process := dhcp.findProcess()
	if process == nil {
		return nil
	}
	defer func(){
		os.Remove("/var/run/dhcpd.pid")
	}()
	process.Kill()

	i := 0
	for dhcp.findProcess() != nil {	// 10 seconds timeout
		if (i > 100) {
			return ErrCannotStop
		}
		time.Sleep(time.Millisecond * 100)
		i++
	}

	return nil
}
func (dhcp DHCP) Stop() error {
	dhcp.lock()
	defer dhcp.unlock()
	return dhcp.stop()
}
func (dhcp DHCP) stop() error {
	return dhcp.stopProcess()
}
func (dhcp DHCP) Restart() error {
	dhcp.lock()
	defer dhcp.unlock()
	dhcp.stop()
	return dhcp.start()
}

func (dhcp DHCP) lock() {
	//fmt.Println("dhcp.lock()")
	dhcp.mutex.Lock()
}

func (dhcp DHCP) unlock() {
	//fmt.Println("dhcp.unlock()")
	dhcp.mutex.Unlock()
}

func (dhcp DHCP) SetConfig(cfg cfg.Root) {
	dhcp.lock()
	defer dhcp.unlock()
	dhcp.Config.Root = cfg
}
