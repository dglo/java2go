package grammar

import (
	"testing"

	"java2go/testutil"
)

func Test_Simple(t *testing.T) {
	pgm := "public class foo{ private int val;" +
		" public int getVal() { return val; }" +
		" public foo(int val) { this.val = val; }" +
		"}"

	rdr := NewStringReader(pgm)

	lx := NewLexer(rdr, false)

	rtn := JulyParse(lx)
	testutil.AssertEqual(t, rtn, 0, "Expected", 0, "not", rtn)
}
