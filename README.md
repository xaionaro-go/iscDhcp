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
  "UserDefinedOptionFields": {
    "ms-classless-static-routes": {
      "Code": 249,
      "ValueType": 1
    },
    "rfc3442-classless-static-routes": {
      "Code": 121,
      "ValueType": 1
    }
  },
  "Options": {
    "DefaultLeaseTime": 600,
    "MaxLeaseTime": 7200,
    "Authoritative": true,
    "DomainNameServers": [
      "ns1.example.org",
      "ns2.example.org"
    ],
    "DomainName": "example.org",
    "Range": {}
  },
  "Subnets": {
    "10.254.239.32": {
      "Network": "10.254.239.32",
      "Mask": "AAAAAAAAAAAAAP//////4A==",
      "Options": {
        "Range": {
          "Start": "10.254.239.40",
          "End": "10.254.239.60"
        },
        "Routers": [
          "rtr-239-32-1.example.org"
        ],
        "BroadcastAddress": "10.254.239.31"
      }
    },
    "10.5.5.0": {
      "Network": "10.5.5.0",
      "Mask": "AAAAAAAAAAAAAP//////4A==",
      "Options": {
        "DefaultLeaseTime": 600,
        "MaxLeaseTime": 7200,
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
        "BroadcastAddress": "10.5.5.31"
      }
    },
    "10.55.0.0": {
      "Network": "10.55.0.0",
      "Mask": "AAAAAAAAAAAAAP//////AA==",
      "Options": {
        "DomainNameServers": [
          "10.55.0.31"
        ],
        "DomainName": "campus.someuniver.edu",
        "Range": {
          "Start": "10.55.0.230",
          "End": "10.55.0.239"
        },
        "Routers": [
          "10.55.0.8"
        ],
        "BroadcastAddress": "10.55.0.255",
        "Filename": "pxelinux.0",
        "NextServer": "10.55.0.12",
        "RootPath": "/srv/share/nfs/public",
        "MTU": 1300,
        "Custom": {
          "121": "GMCoZAoyAAs=",
          "249": "GMCoZAoyAAs="
        }
      }
    }
  }
}
```

