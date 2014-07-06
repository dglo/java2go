package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/dglo/java2go/parser"
)

const pgm = "package main;\npublic class foo\n{\n\tpublic static void main(String[] args) {\n\t\tSystem.out.println(\"hello\");\n\t}\n}\n"

func TestLexer(t *testing.T) {
	//parser.JulyDebug = 9

	l := parser.NewLexer(bufio.NewReader(strings.NewReader(pgm)), false)

	lval := &parser.JulySymType{}

	var num int
	for {
		rtn := l.Lex(lval)
		if rtn == 0 {
			break
		}

		num++

		var tname string
		if rtn >= 57346 {
			tname = parser.JulyToknames[rtn - 57346]
		} else {
			tname = fmt.Sprintf("tok#%d (ch '%c')", rtn, byte(rtn))
		}

		fmt.Printf("#%d: rtn %s\n\t%v\n", num, tname, lval)
	}
}

func TestBoth(t *testing.T) {
	//parser.JulyDebug = 9

	l := parser.NewLexer(bufio.NewReader(strings.NewReader(pgm)), false)

	parser.JulyParse(l)
	//if err != nil {
	//	fmt.Printf("Error in \"%v\": %v\n", input, err)
	//} else {
	//	fmt.Printf("\"%v\" -> %v<%T>\n", input, st.(java.Val), st)
	//}
	//fmt.Printf("input \"%v\" -> %d\n", pgm, sym)
}
