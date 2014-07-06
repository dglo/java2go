package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type Config struct {
	interfaceMap map[string]string
	interfaceKeys []string
	packageMap map[string]string
	packageKeys []string
	receiverMap map[string]string
	receiverKeys []string
}

const TYPE_INTERFACE = "INTERFACE"
const TYPE_PACKAGE = "PACKAGE"
const TYPE_RECEIVER = "RECEIVER"

func addEntry(entryMap map[string]string, typeName string, key string,
	val string) map[string]string {
	if entryMap == nil {
		entryMap = make(map[string]string)
	}

	if _, ok := entryMap[key]; ok {
		log.Printf("Overwriting %s entry %s value %s with %s\n", typeName,
			key, entryMap[key], val)
	}

	entryMap[key] = val

	return entryMap
}

func (cfg *Config) AddInterface(name string) {
	cfg.interfaceMap = addEntry(cfg.interfaceMap, TYPE_INTERFACE, name, name)
	cfg.interfaceKeys = nil
}

func (cfg *Config) AddPackage(name string, value string) {
	cfg.packageMap = addEntry(cfg.packageMap, TYPE_PACKAGE, name, value)
	cfg.packageKeys = nil
}

func (cfg *Config) AddReceiver(name string, value string) {
	cfg.receiverMap = addEntry(cfg.receiverMap, TYPE_RECEIVER, name, value)
	cfg.receiverKeys = nil
}

func getValue(entryMap map[string]string, key string) string {
	if entryMap != nil {
		if val, ok := entryMap[key]; ok {
			return val
		}
	}

	return ""
}

func ReadConfig(name string) *Config {
	fd, err := os.Open(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open config file %s\n", name)
		return nil
	}

	cfg := &Config{}

	scan := bufio.NewScanner(bufio.NewReader(fd))
	for scan.Scan() {
		flds := strings.Fields(scan.Text())
		if len(flds) == 0 || strings.Index(flds[0], "#") == 0 {
			// ignore empty lines and comments
			continue
		}

		switch strings.ToUpper(flds[0]) {
		case TYPE_INTERFACE:
			if len(flds) != 2 {
				log.Printf("Bad config line: %s\n", scan.Text())
			} else {
				cfg.AddInterface(flds[1])
			}
		case TYPE_PACKAGE:
			if len(flds) != 4 {
				log.Printf("Bad config line: %s\n", scan.Text())
			} else {
				cfg.AddPackage(flds[1], flds[3])
			}
		case TYPE_RECEIVER:
			if len(flds) != 4 {
				log.Printf("Bad config line: %s\n", scan.Text())
			} else {
				cfg.AddReceiver(flds[1], flds[3])
			}
		}
	}

	return cfg
}

func (cfg *Config) Dump(out io.Writer) {
	need_nl := false

	if len(cfg.packageMap) > 0 {
		if need_nl { fmt.Fprintln(out) }
		fmt.Fprintln(out, "# map Java packages to Go packages")
		for _, k := range cfg.PackageKeys() {
			fmt.Fprintf(out, "%v %v -> %v\n", TYPE_PACKAGE, k, cfg.Package(k))
		}
		need_nl = true
	}

	if len(cfg.interfaceMap) > 0 {
		if need_nl { fmt.Fprintln(out) }
		fmt.Fprintln(out, "# names which should be treated as interfaces" +
			" rather than structs")
		for _, k := range cfg.Interfaces() {
			fmt.Fprintf(out, "%v %v\n", TYPE_INTERFACE, k)
		}
		need_nl = true
	}

	if len(cfg.receiverMap) > 0 {
		if need_nl { fmt.Fprintln(out) }
		fmt.Fprintln(out, "# receiver name to use (other than 'rcvr')")
		for _, k := range cfg.ReceiverKeys() {
			fmt.Fprintf(out, "%v %v -> %v\n", TYPE_RECEIVER, k, cfg.Receiver(k))
		}
		need_nl = true
	}
}

func (cfg *Config) FindPackage(str string) string {
	var pstr string
	var pval string
	var pextra string
	for k, v := range cfg.packageMap {
		idx := strings.Index(str, k)
		if idx == 0 {
			if len(k) > len(pstr) {
				pstr = k
				pval = v
				if len(str) == len(pstr) {
					pextra = ""
					break
				}
				pextra = str[len(pstr):]
			}
		}
	}

	return pval + pextra
}

func (cfg *Config) Interfaces() []string {
	if cfg.interfaceKeys == nil {
		cfg.interfaceKeys = make([]string, len(cfg.interfaceMap))
		var i int
		for k := range cfg.interfaceMap {
			cfg.interfaceKeys[i] = k
			i += 1
		}

		sort.Sort(sort.StringSlice(cfg.interfaceKeys))
	}

	return cfg.interfaceKeys
}

func (cfg *Config) IsInterface(name string) bool {
	_, ok := cfg.interfaceMap[name]
	return ok
}

func (cfg *Config) Package(key string) string {
	return getValue(cfg.packageMap, key)
}

func (cfg *Config) PackageKeys() []string {
	if cfg.packageKeys == nil {
		cfg.packageKeys = make([]string, len(cfg.packageMap))
		var i int
		for k := range cfg.packageMap {
			cfg.packageKeys[i] = k
			i += 1
		}

		sort.Sort(sort.StringSlice(cfg.packageKeys))
	}

	return cfg.packageKeys
}

func (cfg *Config) Receiver(key string) string {
	return getValue(cfg.receiverMap, key)
}

func (cfg *Config) ReceiverKeys() []string {
	if cfg.receiverKeys == nil {
		cfg.receiverKeys = make([]string, len(cfg.receiverMap))

		var i int
		for k := range cfg.receiverMap {
			cfg.receiverKeys[i] = k
			i += 1
		}

		sort.Sort(sort.StringSlice(cfg.receiverKeys))
	}

	return cfg.receiverKeys
}

func (cfg *Config) String() string {
	return fmt.Sprintf("Config[%d pkgs, %d rcvrs]", len(cfg.packageMap),
		len(cfg.receiverMap))
}
