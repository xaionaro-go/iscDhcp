= ISC DHCP Server API

This package controls the process (starts, stops) and the configuration file of `isc-dhcp-server`.

You can use [github.com/xaionaro-go/iscDhcp/cfg](https://github.com/xaionaro-go/iscDhcp/tree/master/cfg) as just a config parser.

```
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

result:

{
  "Options": {
    "DefaultLeaseTime": 600,
    "MaxLeaseTime": 7200,
    "Authoritative": true,
    "LogFacility": "",
    "DomainNameServers": [
      "ns1.example.org",
      "ns2.example.org"
    ],
    "DomainName": "example.org",
    "Range": {
      "Start": "",
      "End": ""
    },
    "Routers": null,
    "BroadcastAddress": "",
    "Other": null
  },
  "Subnets": {
    "10.254.239.32": {
      "Network": "10.254.239.32",
      "Mask": "AAAAAAAAAAAAAP//////4A==",
      "Options": {
        "DefaultLeaseTime": 0,
        "MaxLeaseTime": 0,
        "Authoritative": false,
        "LogFacility": "",
        "DomainNameServers": null,
        "DomainName": "",
        "Range": {
          "Start": "10.254.239.40",
          "End": "10.254.239.60"
        },
        "Routers": [
          "rtr-239-32-1.example.org"
        ],
        "BroadcastAddress": "10.254.239.31",
        "Other": null
      }
    },
    "10.5.5.0": {
      "Network": "10.5.5.0",
      "Mask": "AAAAAAAAAAAAAP//////4A==",
      "Options": {
        "DefaultLeaseTime": 600,
        "MaxLeaseTime": 7200,
        "Authoritative": false,
        "LogFacility": "",
        "DomainNameServers": [
          "ns1.internal.example.org"
        ],
        "DomainName": "internal.example.org",
        "Range": {
          "Start": "10.5.5.26",
          "End": "10.5.5.30"
        },
        "Routers": [
          "10.5.5.1"
        ],
        "BroadcastAddress": "10.5.5.31",
        "Other": null
      }
    }
  }
}
```

