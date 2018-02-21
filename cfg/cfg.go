package cfg

import (
	"bufio"
	"fmt"
	"github.com/timtadh/lexmachine"
	"github.com/xaionaro-go/isccfg"
	"net"
	"os"
	"strconv"
	"strings"
)

type Range struct {
	Start net.IP `json:",omitempty"`
	End   net.IP `json:",omitempty"`
}

type CustomOptions map[int][]byte

type Options struct {
	DefaultLeaseTime  int           `json:",omitempty"`
	MaxLeaseTime      int           `json:",omitempty"`
	Authoritative     bool          `json:",omitempty"`
	LogFacility       string        `json:",omitempty"`
	DomainNameServers []string      `json:",omitempty"`
	DomainName        string        `json:",omitempty"`
	Range             Range         `json:",omitempty"`
	Routers           []string      `json:",omitempty"`
	BroadcastAddress  string        `json:",omitempty"`
	Filename          string        `json:",omitempty"`
	NextServer        string        `json:",omitempty"`
	RootPath          string        `json:",omitempty"`
	MTU               int           `json:",omitempty"`
	Custom            CustomOptions `json:",omitempty"`
}

type Subnet struct {
	Network net.IP
	Mask    net.IPMask
	Options Options
}
type Subnets map[string]Subnet

const (
	BYTEARRAY = ValueType(1)
)

type ValueType int
type UserDefinedOptionField struct {
	Code      int
	ValueType ValueType
}

type UserDefinedOptionFields map[string]*UserDefinedOptionField

type Root struct {
	UserDefinedOptionFields UserDefinedOptionFields
	Options                 Options
	Subnets                 Subnets
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

func (subnet *Subnet) parse(root *Root, netStr string, cfgRaw *isccfg.Config) (err error) {
	var maskStr string
	cfgRaw, _ = cfgRaw.Unwrap()
	cfgRaw, maskStr = cfgRaw.Unwrap()

	subnet.Network = net.ParseIP(netStr)
	subnet.Mask = net.IPMask(net.ParseIP(maskStr))

	for k, v := range *cfgRaw {
		err = subnet.Options.parse(root, k, v)
		if err != nil {
			return
		}
	}

	return nil
}
func (root *Root) addUserDefinedOptionField(k string, c *isccfg.Config) (err error) {
	words := c.Unroll()
	name := k
	if len(words) < 3 {
		return fmt.Errorf(`too short: %v`, words)
	}
	if words[1] != "=" {
		return fmt.Errorf(`"=" is expected, got: %v`, words[1])
	}
	code, err := strconv.Atoi(words[0])
	if err != nil {
		return err
	}

	var valueType ValueType
	valueTypeStr := strings.Join(words[2:], " ")
	switch valueTypeStr {
	case "array of integer 8":
		valueType = BYTEARRAY
	default:
		return fmt.Errorf(`this case is not implemented, yet: %v`, valueTypeStr)
	}

	field := UserDefinedOptionField{
		Code:      code,
		ValueType: valueType,
	}

	//fmt.Println("field", name, field)
	root.UserDefinedOptionFields[name] = &field

	return nil
}

func (options *Options) parse(root *Root, k string, value isccfg.Value) (err error) {
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
	case "filename":
		options.Filename = cfgRaw.Values()[0]
	case "next-server":
		options.NextServer = cfgRaw.Values()[0]

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
			case "root-path":
				options.RootPath = c.Values()[0]
			case "interface-mtu":
				options.MTU, err = strconv.Atoi(c.Values()[0])
			case "static-routes":

			default:
				field := root.UserDefinedOptionFields[k]
				if field == nil {
					c, codeWord := c.Unwrap()
					if codeWord != "code" {
						fmt.Fprintf(os.Stderr, "Not recognized option: %v\n", k)
						break
					}
					err := root.addUserDefinedOptionField(k, c)
					if err != nil {
						return err
					}
					break
				}

				switch field.ValueType {
				case BYTEARRAY:
					var result []byte
					bytesStr := c.Values()
					for _, str := range bytesStr {
						oneByte, err := strconv.Atoi(str)
						if err != nil {
							return err
						}
						result = append(result, byte(oneByte))
					}
					options.Custom[field.Code] = result
				default:
					panic("This shouldn't happened")
				}

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
		Subnets:                 Subnets{},
		UserDefinedOptionFields: UserDefinedOptionFields{},
		Options: Options{
			Custom: CustomOptions{},
		},
	}
	for k, v := range cfgRaw {
		if k == "subnet" {
			continue
		}
		err := cfg.Root.Options.parse(&cfg.Root, k, v)
		if err != nil {
			return err
		}
	}
	for k, v := range cfgRaw {
		if k != "subnet" {
			continue
		}
		for net, netDetails := range *(v.(*isccfg.Config)) {
			newSubnet := Subnet{
				Options: Options{
					Custom: CustomOptions{},
				},
			}
			err := newSubnet.parse(&cfg.Root, net, netDetails.(*isccfg.Config))
			if err != nil {
				return err
			}
			cfg.Root.Subnets[net] = newSubnet
		}
	}

	return nil
}
