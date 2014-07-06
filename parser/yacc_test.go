package parser

import (
	"fmt"
	"os"
	"testing"
)

func Test_Simple(t *testing.T) {
	pgm := "public class foo{ private int val;" +
		" public int getVal() { return val; }" +
		" public foo(int val) { this.val = val; }" +
		"}";

	rdr := NewStringReader(pgm)

	lx := NewLexer(rdr, false)

	rtn := JulyParse(lx)
	assertEqual(t, rtn, 0, "Expected", 0, "not", rtn)
}

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
		"}\n";

	rdr := NewStringReader(src)

	lx := NewLexer(rdr, false)

	rtn := JulyParse(lx)
	assertEqual(t, rtn, 0, "Expected", 0, "not", rtn)
	assertNotNil(t, lx.JavaProgram(), "Parser did not return Java parse tree")

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
