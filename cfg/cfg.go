package cfg

import (
	"bufio"
	"fmt"
	"github.com/timtadh/lexmachine"
	"github.com/xaionaro-go/isccfg"
	"net"
	"os"
	"strconv"
)

type Range struct {
	Start net.IP
	End   net.IP
}

type Options struct {
	DefaultLeaseTime  int
	MaxLeaseTime      int
	Authoritative     bool
	LogFacility       string
	DomainNameServers []string
	DomainName        string
	Range             Range
	Routers           []string
	BroadcastAddress  string
	Other             map[int][]byte
}

type Subnet struct {
	Network net.IP
	Mask    net.IPMask
	Options Options
}
type Subnets map[string]Subnet

type Root struct {
	Options Options
	Subnets Subnets
}

type Config struct {
	Root
	lexer *lexmachine.Lexer
}

func NewConfig() *Config {
	return &Config{
		Root: Root{
			Subnets: Subnets{},
		},
		lexer: isccfg.NewLexer(),
	}
}

func (subnet *Subnet) parse(netStr string, cfgRaw *isccfg.Config) (err error) {
	var maskStr string
	cfgRaw, _ = cfgRaw.Unwrap()
	cfgRaw, maskStr = cfgRaw.Unwrap()

	subnet.Network = net.ParseIP(netStr)
	subnet.Mask = net.IPMask(net.ParseIP(maskStr))

	for k, v := range *cfgRaw {
		err = subnet.Options.parse(k, v)
		if err != nil {
			return
		}
	}

	return nil
}
func (options *Options) parse(k string, value isccfg.Value) (err error) {
	cfgRaw, _ := value.(*isccfg.Config)
	if k == "_value" {
		k = value.([]string)[0]
	}

	switch k {
	case "authoritative":
		options.Authoritative = true
	case "default-lease-time":
		options.DefaultLeaseTime, err = strconv.Atoi(cfgRaw.Values()[0])
	case "max-lease-time":
		options.MaxLeaseTime, err = strconv.Atoi(cfgRaw.Values()[0])
	case "log-facility":
		options.LogFacility = cfgRaw.Values()[0]
	case "ddns-update-style":
		// TODO: implement this

	case "range":
		var startStr, endStr string
		cfgRaw, startStr = cfgRaw.Unwrap()
		for startStr == "dynamic-bootp" {
			cfgRaw, startStr = cfgRaw.Unwrap()
		}
		endStr = cfgRaw.Values()[0]
		options.Range.Start = net.ParseIP(startStr)
		options.Range.End = net.ParseIP(endStr)

	case "option":
		for k, v := range *cfgRaw {
			c := v.(*isccfg.Config)
			switch k {
			case "domain-name":
				options.DomainName = c.Values()[0]
			case "domain-name-servers":
				options.DomainNameServers = c.Values()
			case "broadcast-address":
				options.BroadcastAddress = c.Values()[0]
			case "routers":
				options.Routers = c.Values()
			default:
				fmt.Fprintf(os.Stderr, "Not recognized option: %v\n", k)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Not recognized: %v\n", k)
	}

	return
}

func (cfg *Config) LoadFrom(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	cfgReader := bufio.NewReader(file)

	cfgRaw, err := isccfg.Parse(cfgReader)
	if err != nil {
		return err
	}

	cfg.Root = Root{
		Subnets: Subnets{},
	}
	for k, v := range cfgRaw {
		switch k {
		case "subnet":
			for net, netDetails := range *(v.(*isccfg.Config)) {
				newSubnet := Subnet{}
				err := newSubnet.parse(net, netDetails.(*isccfg.Config))
				if err != nil {
					return err
				}
				cfg.Root.Subnets[net] = newSubnet
			}
		default:
			err := cfg.Root.Options.parse(k, v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
