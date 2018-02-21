package iscDhcp

import (
	"github.com/xaionaro-go/iscDhcp/cfg"
)

const (
	CFG_PATH = "/etc/dhcp/dhcpd-dynamic.conf"
)

struct DHCP {
	cfg cfg.Root

}

func NewDHCP() *DHCP {
	return &DHCP{}
}

func (dhcp *DHCP) ReloadConfig() {
	dhcp.cfg.LoadFrom(CFG_PATH)
}

