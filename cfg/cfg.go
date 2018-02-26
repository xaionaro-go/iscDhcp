package cfg

import (
	"bufio"
	"fmt"
	"github.com/timtadh/lexmachine"
	"github.com/xaionaro-go/isccfg"
	"io"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Range struct {
	Start net.IP `json:",omitempty"`
	End   net.IP `json:",omitempty"`
}

type CustomOptions map[int][]byte
type NS net.NS
type NSs []NS

type Options struct {
	DefaultLeaseTime  int           `json:",omitempty"`
	MaxLeaseTime      int           `json:",omitempty"`
	Authoritative     bool          `json:",omitempty"`
	LogFacility       string        `json:",omitempty"`
	DomainName        string        `json:",omitempty"`
	DomainNameServers NSs           `json:",omitempty"`
	Range             Range         `json:",omitempty"`
	Routers           []string      `json:",omitempty"`
	BroadcastAddress  string        `json:",omitempty"`
	NextServer        string        `json:",omitempty"`
	Filename          string        `json:",omitempty"`
	RootPath          string        `json:",omitempty"`
	MTU               int           `json:",omitempty"`
	Custom            CustomOptions `json:",omitempty"`
}


func (nss NSs) ToIPs() (result []net.IP) {
	for _, ns := range nss {
		result = append(result, net.ParseIP(ns.Host))
	}
	return
}
func (nss NSs) ToStrings() (result []string) {
	for _, ns := range nss {
		result = append(result, ns.Host)
	}
	return
}
func (nss NSs) ToNetNSs() (result []net.NS) {
	for _, ns := range nss {
		result = append(result, net.NS(ns))
	}
	return
}
func (nss *NSs) Set(newNssRaw []string) {
	*nss = NSs{}
	for _, newNsRaw := range newNssRaw {
		*nss = append(*nss, NS{Host: newNsRaw})
	}
	return
}

func (options Options) configWrite(out io.Writer, root Root, indent string) (err error) {
	if options.DefaultLeaseTime != 0 {
		_, err = fmt.Fprintf(out, "%vdefault-lease-time %v;\n", indent, options.DefaultLeaseTime)
		if err != nil {
			return err
		}
	}
	if options.MaxLeaseTime != 0 {
		fmt.Fprintf(out, "%vmax-lease-time %v;\n", indent, options.MaxLeaseTime)
	}
	if options.Authoritative {
		fmt.Fprintf(out, "%vauthoritative;\n", indent)
	}
	if options.LogFacility != "" {
		fmt.Fprintf(out, "%vlog-facility %v;\n", indent, options.LogFacility)
	}
	if options.DomainName != "" {
		fmt.Fprintf(out, "%voption domain-name \"%v\";\n", indent, options.DomainName)
	}
	if len(options.DomainNameServers) > 0 {
		fmt.Fprintf(out, "%voption domain-name-servers %v;\n", indent, strings.Join(options.DomainNameServers.ToStrings(), ", "))
	}
	if options.Range.Start != nil {
		fmt.Fprintf(out, "%vrange %v %v;\n", indent, options.Range.Start, options.Range.End)
	}
	if len(options.Routers) > 0 {
		fmt.Fprintf(out, "%voption routers %v;\n", indent, strings.Join(options.Routers, ", "))
	}
	if options.BroadcastAddress != "" {
		fmt.Fprintf(out, "%voption broadcast-address %v;\n", indent, options.BroadcastAddress)
	}
	if options.NextServer != "" {
		fmt.Fprintf(out, "%vnext-server %v;\n", indent, options.NextServer)
	}
	if options.Filename != "" {
		fmt.Fprintf(out, "%vfilename \"%v\";\n", indent, options.Filename)
	}
	if options.RootPath != "" {
		fmt.Fprintf(out, "%voption root-path \"%v\";\n", indent, options.RootPath)
	}
	if options.MTU != 0 {
		fmt.Fprintf(out, "%voption interface-mtu %v;\n", indent, options.MTU)
	}

	var keys []int
	for k := range options.Custom {
		keys = append(keys, k)
	}
	sort.IntSlice(keys).Sort()

	customOptionNameMap := map[int]string{}
	for k, f := range root.UserDefinedOptionFields {
		customOptionNameMap[f.Code] = k
	}
	customOptionValueTypeMap := map[int]ValueType{}
	for _, f := range root.UserDefinedOptionFields {
		customOptionValueTypeMap[f.Code] = f.ValueType
	}

	for _, k := range keys {
		option := options.Custom[k]

		var valueString string
		switch customOptionValueTypeMap[k] {
		case BYTEARRAY:
			var result []string
			for _, byteValue := range option {
				result = append(result, strconv.Itoa(int(byteValue)))
			}
			valueString = strings.Join(result, ", ")
		default:
			panic("This shouldn't happened")
		}

		fmt.Fprintf(out, "%voption %v %v;\n", indent, customOptionNameMap[k], valueString)
	}

	return nil
}

func (options Options) ConfigWrite(out io.Writer, root Root) error {
	return options.configWrite(out, root, "")
}

type Subnet struct {
	Network net.IPNet
	Options Options
}
type Subnets map[string]Subnet

func (subnet Subnet) ConfigWrite(out io.Writer, root Root) (err error) {
	_, err = fmt.Fprintf(out, "\nsubnet %v netmask %v {\n", subnet.Network.IP, net.IP(subnet.Network.Mask).String())
	if err != nil {
		return err
	}
	err = subnet.Options.configWrite(out, root, "\t")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "}\n")
	if err != nil {
		return err
	}
	return nil
}

func (subnets Subnets) ConfigWrite(out io.Writer, root Root) error {
	var keys []string
	for k := range subnets {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()

	for _, k := range keys {
		subnet := subnets[k]
		err := subnet.ConfigWrite(out, root)
		if err != nil {
			return err
		}
	}

	return nil
}

type ToSubnetI interface {
	ToSubnet() Subnet
}

func (subnet Subnet) ToSubnet() Subnet {
	return subnet
}

func (subnets Subnets) ISet(subnetI ToSubnetI) error {
	subnet := subnetI.ToSubnet()
	if subnet.Network.IP.String() == (net.IP{}).String() {
		panic("subnet.Network.IP is empty")
	}
	subnets[subnet.Network.IP.String()] = subnet
	return nil
}

const (
	BYTEARRAY = ValueType(1)
)

type ValueType int

func (vt ValueType) ConfigString() string {
	switch vt {
	case BYTEARRAY:
		return "array of integer 8"
	}
	panic("This shouldn't happened")
	return ""
}

type UserDefinedOptionField struct {
	Code      int
	ValueType ValueType
}

type UserDefinedOptionFields map[string]*UserDefinedOptionField

func (fields UserDefinedOptionFields) ConfigWrite(out io.Writer) (err error) {
	var keys []string
	for k := range fields {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()

	for _, k := range keys {
		field := fields[k]
		_, err = fmt.Fprintf(out, "option %v code %v = %v;\n", k, field.Code, field.ValueType.ConfigString())
		if err != nil {
			return err
		}
	}

	return err
}

type Root struct {
	UserDefinedOptionFields UserDefinedOptionFields
	Options                 Options
	Subnets                 Subnets
}

func NewRoot() *Root {
	return &Root{
		Subnets:                 Subnets{},
		UserDefinedOptionFields: UserDefinedOptionFields{},
		Options: Options{
			Custom: CustomOptions{},
		},
	}
}

func NewSubnet() *Subnet {
	return &Subnet {
		Options: Options{
			Custom: CustomOptions{},
		},
	}
}

func (root Root) ConfigWrite(out io.Writer) (err error) {
	err = root.UserDefinedOptionFields.ConfigWrite(out)
	if err != nil {
		return err
	}
	err = root.Options.ConfigWrite(out, root)
	if err != nil {
		return err
	}
	err = root.Subnets.ConfigWrite(out, root)
	if err != nil {
		return err
	}

	return nil
}

type Config struct {
	Root
	lexer *lexmachine.Lexer
}

func NewConfig() *Config {
	return &Config{
		Root: *NewRoot(),
		lexer: isccfg.NewLexer(),
	}
}

func (subnet *Subnet) parse(root *Root, netStr string, cfgRaw *isccfg.Config) (err error) {
	var maskStr string
	cfgRaw, _ = cfgRaw.Unwrap()
	cfgRaw, maskStr = cfgRaw.Unwrap()

	subnet.Network.IP   = net.ParseIP(netStr)
	subnet.Network.Mask = net.IPMask(net.ParseIP(maskStr))

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
				options.DomainNameServers.Set(c.Values())
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

func (cfg Config) ConfigWriteTo(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	cfgWriter := bufio.NewWriter(file)
	defer cfgWriter.Flush()
	return cfg.ConfigWrite(cfgWriter)
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

	cfg.Root = *NewRoot()
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
			newSubnet := *NewSubnet()
			err := newSubnet.parse(&cfg.Root, net, netDetails.(*isccfg.Config))
			if err != nil {
				return err
			}
			cfg.Root.Subnets[net] = newSubnet
		}
	}

	return nil
}

func (cfg Config) ConfigWrite(out io.Writer) error {
	return cfg.Root.ConfigWrite(out)
}
