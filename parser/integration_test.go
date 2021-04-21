package parser

import (
	"fmt"
	"os"
	"testing"

	"java2go/grammar"
	"java2go/testutil"
)

func Test_Main(t *testing.T) {
	src := "public class foo\n" +
		"{\n" +
		" private int val;\n" +
		" public int getVal() { return val; }\n" +
		" public foo(int val) { this.val = val; }\n" +
		" public static final void main(String[] args) {\n" +
		"  for (int i = 0; i < args.length; i++) {\n" +
		"   System.out.println(\"Arg#\" + i + \"=\" + args[i]);\n" +
		"   System.err.printf(\"Arg#%d=%s\\n\", i, args[i]);\n" +
		"  }\n" +
		" }\n" +
		"}\n"

	rdr := grammar.NewStringReader(src)

	lx := grammar.NewLexer(rdr, false)

	rtn := grammar.JulyParse(lx)
	testutil.AssertEqual(t, rtn, 0, "Expected", 0, "not", rtn)
	testutil.AssertNotNil(t, lx.JavaProgram(), "Parser did not return Java parse tree")

	pgm := NewGoProgram("", nil, false)
	pgm.Analyze(lx.JavaProgram())

	fmt.Println("========= PARSE TREE =============")
	pgm.WriteString(os.Stdout)
	fmt.Println()

	for _, rule := range StandardRules {
		pgm.RunTransform(rule, pgm, nil, nil)
	}

	fmt.Println("========= FINAL PROGRAM =============")
	pgm.Dump(os.Stdout)
}
