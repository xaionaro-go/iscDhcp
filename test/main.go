package main

import (
	"encoding/json"
	"fmt"
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

	fmt.Printf("\n\n--- Regenerating the configuration:\n\n")

	cfg.ConfigWrite(os.Stdout)
}
