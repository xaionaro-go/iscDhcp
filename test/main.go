package main

import (
	"encoding/json"
	"github.com/xaionaro-go/iscDhcp/cfg"
	"os"
)

func main() {
	cfg := cfg.NewConfig()
	err := cfg.LoadFrom("/etc/dhcp/dhcpd.conf")
	if err != nil {
		panic(err)
	}

	jsonEncoder := json.NewEncoder(os.Stderr)
	jsonEncoder.SetIndent("", "  ")
	jsonEncoder.Encode(cfg)
}
