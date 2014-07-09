package grammar

import (
	"testing"

	"github.com/dglo/java2go/testutil"
)

func Test_FileLexer(t *testing.T) {
	pgm := "public class foo{ private int val;" +
		" public int getVal() { return val; }" +
		" public foo(int val) { this.val = val; } }"

	rdr := NewStringReader(pgm)

	lx := NewLexer(rdr, false)
	testutil.AssertNotNil(t, lx, "Lexer should not be nil")
}
