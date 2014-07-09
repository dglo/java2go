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

// configuration file
type Config struct {
	interfaceMap map[string]string
	interfaceList []string
	packageMap map[string]string
	packageList []string
	receiverMap map[string]string
	receiverList []string
}

// keyword for defining Java interfaces
const typeInterface = "INTERFACE"
// keyword for mapping Java package names to Go names
const typePackage = "PACKAGE"
// keyword for declaring receiver names for a class
const typeReceiver = "RECEIVER"

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

func (cfg *Config) addInterface(name string) {
	cfg.interfaceMap = addEntry(cfg.interfaceMap, typeInterface, name, name)
	cfg.interfaceList = nil
}

func (cfg *Config) addPackage(name string, value string) {
	cfg.packageMap = addEntry(cfg.packageMap, typePackage, name, value)
	cfg.packageList = nil
}

func (cfg *Config) addReceiver(name string, value string) {
	cfg.receiverMap = addEntry(cfg.receiverMap, typeReceiver, name, value)
	cfg.receiverList = nil
}

func getValue(entryMap map[string]string, key string) string {
	if entryMap != nil {
		if val, ok := entryMap[key]; ok {
			return val
		}
	}

	return ""
}

// read the configuration file
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
		case typeInterface:
			if len(flds) != 2 {
				log.Printf("Bad config line: %s\n", scan.Text())
			} else {
				cfg.addInterface(flds[1])
			}
		case typePackage:
			if len(flds) != 4 {
				log.Printf("Bad config line: %s\n", scan.Text())
			} else {
				cfg.addPackage(flds[1], flds[3])
			}
		case typeReceiver:
			if len(flds) != 4 {
				log.Printf("Bad config line: %s\n", scan.Text())
			} else {
				cfg.addReceiver(flds[1], flds[3])
			}
		}
	}

	return cfg
}

// print the configuration data to the output
func (cfg *Config) Dump(out io.Writer) {
	need_nl := false

	if len(cfg.packageMap) > 0 {
		if need_nl { fmt.Fprintln(out) }
		fmt.Fprintln(out, "# map Java packages to Go packages")
		for _, k := range cfg.packageKeys() {
			fmt.Fprintf(out, "%v %v -> %v\n", typePackage, k,
				cfg.packageName(k))
		}
		need_nl = true
	}

	if len(cfg.interfaceMap) > 0 {
		if need_nl { fmt.Fprintln(out) }
		fmt.Fprintln(out, "# names which should be treated as interfaces" +
			" rather than structs")
		for _, k := range cfg.interfaces() {
			fmt.Fprintf(out, "%v %v\n", typeInterface, k)
		}
		need_nl = true
	}

	if len(cfg.receiverMap) > 0 {
		if need_nl { fmt.Fprintln(out) }
		fmt.Fprintln(out, "# receiver name to use (other than 'rcvr')")
		for _, k := range cfg.receiverKeys() {
			fmt.Fprintf(out, "%v %v -> %v\n", typeReceiver, k, cfg.receiver(k))
		}
		need_nl = true
	}
}

func (cfg *Config) findPackage(str string) string {
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

func (cfg *Config) interfaces() []string {
	if cfg.interfaceList == nil {
		cfg.interfaceList = make([]string, len(cfg.interfaceMap))
		var i int
		for k := range cfg.interfaceMap {
			cfg.interfaceList[i] = k
			i += 1
		}

		sort.Sort(sort.StringSlice(cfg.interfaceList))
	}

	return cfg.interfaceList
}

func (cfg *Config) isInterface(name string) bool {
	_, ok := cfg.interfaceMap[name]
	return ok
}

func (cfg *Config) packageName(key string) string {
	return getValue(cfg.packageMap, key)
}

func (cfg *Config) packageKeys() []string {
	if cfg.packageList == nil {
		cfg.packageList = make([]string, len(cfg.packageMap))
		var i int
		for k := range cfg.packageMap {
			cfg.packageList[i] = k
			i += 1
		}

		sort.Sort(sort.StringSlice(cfg.packageList))
	}

	return cfg.packageList
}

func (cfg *Config) receiver(key string) string {
	return getValue(cfg.receiverMap, key)
}

func (cfg *Config) receiverKeys() []string {
	if cfg.receiverList == nil {
		cfg.receiverList = make([]string, len(cfg.receiverMap))

		var i int
		for k := range cfg.receiverMap {
			cfg.receiverList[i] = k
			i += 1
		}

		sort.Sort(sort.StringSlice(cfg.receiverList))
	}

	return cfg.receiverList
}

// print a minimal description of the configuration rules
func (cfg *Config) String() string {
	return fmt.Sprintf("Config[%d ifaces, %d pkgs, %d rcvrs]",
		len(cfg.interfaceMap), len(cfg.packageMap), len(cfg.receiverMap))
}
