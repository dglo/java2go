package main

import (
	"flag"
	"fmt"
	"go/format"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/dglo/java2go/dumper"
	"github.com/dglo/java2go/grammar"
	"github.com/dglo/java2go/parser"
)

const sep = "------------"

func analyze(jp *grammar.JProgramFile, path string, config *parser.Config,
	rules []parser.TransformFunc, print_report, verbose bool) *parser.GoProgram {
	if print_report {
		fmt.Println(sep + " CONVERT " + sep)
		log.Printf("/** Convert %s **/\n", path)
	}

	gp := parser.NewGoProgram(convertPathToGo(path), config, verbose)
	gp.Analyze(jp)

	for _, rule := range rules {
		gp.RunTransform(rule, gp, nil, nil)
	}

	return gp
}

func convertPathToGo(path string) string {
    i := strings.LastIndex(path, "/")

    var name string
    if i < 0 {
        name = path
    } else {
        name = path[i+1:]
    }

    if strings.HasSuffix(strings.ToLower(name), ".java") {
        name = name[:len(name)-5]
    }

    return name + ".go"
}

func dump(gp *parser.GoProgram) {
}

func parseDirectory(dir string, cfg *parser.Config, outPath string,
	debugLex bool, printReport bool, verbose bool) {
	if flist, err := ioutil.ReadDir(dir); err != nil {
		log.Printf("Cannot read \"%v\": %v\n", dir, err)
	} else {
		for _, finfo := range flist {
			fullpath := path.Join(dir, finfo.Name())
			if finfo.IsDir() {
				parseDirectory(fullpath, cfg, outPath, debugLex, printReport,
					verbose)
			} else {
				parseFile(fullpath, finfo, cfg, outPath, debugLex, printReport,
					verbose, false)
			}
		}
	}
}

func parseFile(path string, finfo os.FileInfo, cfg *parser.Config,
	outPath string, debugLex bool, printReport bool, verbose bool,
	logUnknownFile bool) {
	if strings.HasSuffix(path, ".java") {
		parseJava(path, cfg, outPath, debugLex, printReport, verbose)
	} else if strings.HasSuffix(path, ".go") {
		parseGo(path)
	} else if logUnknownFile {
		log.Printf("Ignoring unknown file \"%v\"", path)
	}
}

func parseGo(path string) {
	fset := token.NewFileSet() // positions are relative to fset

	fmt.Printf("Parsing %s\n", path)

	// Parse the file
	f, err := goparser.ParseFile(fset, path, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	dumper.Dump("Parser", f)

	format.Node(os.Stdout, fset, f)
}

func parseJava(path string, cfg *parser.Config, dirPath string,
	debugLex, print_report, verbose bool) {
	if verbose { fmt.Printf("// %s\n", path) }
	l := grammar.NewFileLexer(path, debugLex)
	if l != nil {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					fmt.Fprintf(os.Stderr, "panic in %v: ??%v??<%T>\n",
						path, r, r)
				} else {
					fmt.Fprintf(os.Stderr, "panic in %v: %v\n", path, err)
				}
				debug.PrintStack()
			}
		}()

		prtn := grammar.JulyParse(l)
		if prtn != 0 {
			fmt.Fprintf(os.Stderr, "parser returned %d\n", prtn)
		}

		if l.JavaProgram() == nil {
			fmt.Println("No code found")
		} else {
			rules := parser.StandardRules

			gp := analyze(l.JavaProgram(), path, cfg, rules, print_report,
				verbose)
			if print_report {
				fmt.Println(sep + " GODUMP " + sep)
				gp.WriteString(os.Stdout)
				fmt.Println()
				fmt.Println(sep + " PARSE TREE " + sep)
				gp.DumpTree()
				fmt.Println(sep + " GO " + sep)
				gp.Dump(os.Stdout)
			}
			if dirPath != "" {
				if err := gp.Write(dirPath); err != nil {
					fmt.Fprintf(os.Stderr, "Cannot write %v: %v\n", gp.Name(),
						err)
				}
			} else if !print_report {
				gp.Dump(os.Stdout)
			}
		}

		l.Close()
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Config file")

	var debugFlag bool
	flag.BoolVar(&debugFlag, "debug", false, "Enable parser debugging")

	var debugLexFlag bool
	flag.BoolVar(&debugLexFlag, "debugLex", false, "Enable lexer debugging")

	var dirPath string
	flag.StringVar(&dirPath, "dir", "", "Directory where Go code is written")

	var reportFlag bool
	flag.BoolVar(&reportFlag, "report", false,
		"Do not print dumps/translations to stdout")

	var verboseFlag bool
	flag.BoolVar(&verboseFlag, "verbose", false, "Print more stuff")

	flag.Parse()

	if debugFlag {
		grammar.JulyDebug = 9
	}

	log.SetFlags(0)

	var cfg *parser.Config
	for _, f := range flag.Args() {
		finfo, err := os.Stat(f)
		if err != nil {
			log.Printf("Bad file %v: %v\n", f, err)
			continue
		}

		if cfg == nil {
			if configPath != "" {
				cfg = parser.ReadConfig(configPath)
			}
			if cfg == nil {
				cfg = &parser.Config{}
			}
		}

		if finfo.IsDir() {
			parseDirectory(f, cfg, dirPath, debugLexFlag, reportFlag,
				verboseFlag)
		} else {
			parseFile(f, finfo, cfg, dirPath, debugLexFlag, reportFlag,
				verboseFlag, true)
		}
	}
}
