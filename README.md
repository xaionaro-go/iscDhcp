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

--- Regenerating the configuration:

option ms-classless-static-routes code 249 = array of integer 8;
option rfc3442-classless-static-routes code 121 = array of integer 8;
default-lease-time 600;
max-lease-time 7200;
authoritative;
domain-name example.org;
domain-name-servers ns1.example.org, ns2.example.org;

subnet 10.254.239.32 netmask 255.255.255.224 {
	range 10.254.239.40 10.254.239.60;
	option routers rtr-239-32-1.example.org;
	option broadcast-address 10.254.239.31;
}

subnet 10.5.5.0 netmask 255.255.255.224 {
	default-lease-time 600;
	max-lease-time 7200;
	domain-name internal.example.org;
	domain-name-servers ns1.internal.example.org;
	range 10.5.5.26 10.5.5.30;
	option routers 10.5.5.1;
	option broadcast-address 10.5.5.31;
}

subnet 10.55.0.0 netmask 255.255.255.0 {
	domain-name campus.someuniver.edu;
	domain-name-servers 10.55.0.31;
	range 10.55.0.230 10.55.0.239;
	option routers 10.55.0.8;
	option broadcast-address 10.55.0.255;
	next-server 10.55.0.12;
	filename "pxelinux.0";
	option root-path "/srv/share/nfs/public";
	option interface-mtu 1300;
	option rfc3442-classless-static-routes 24, 192, 168, 100, 10, 55, 0, 11;
	option ms-classless-static-routes 24, 192, 168, 100, 10, 55, 0, 11;
}
```

