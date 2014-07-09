package parser

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dglo/java2go/grammar"
	"github.com/dglo/java2go/testutil"
)

func Test_FixName(t *testing.T) {
	raw := "NaMe"

	modmap := make(map[string]string)
	modmap["public"] = "NaMe"
	modmap["protected"] = "NaMe"
	modmap["private"] = "naMe"

	for modtype, fixed := range modmap {
		jmod := grammar.NewJModifiers(modtype, nil)
		nm := fixName(raw, jmod)
		testutil.AssertEqual(t, nm, fixed, "For modtype", modtype, "expected", fixed,
			"not", nm)
	}
}

func Test_GoProgram_Basic(t *testing.T) {
	name := "foo"
	gp := NewGoProgram(name, nil, false)

	gp.Analyze(nil)

	rcvr := gp.Receiver("foo")
	testutil.AssertEqual(t, rcvr, "rcvr", "Expected rcvr, not", rcvr)

	imap := gp.Imports()
	testutil.AssertNotNil(t, imap, "ImportMap() should not return nil")
	testutil.AssertEqual(t, len(imap), 0, "Did not expect ImportMap to contain", imap)

	decls := gp.Decls()
	testutil.AssertNotNil(t, decls, "Decls() should not return nil")
	testutil.AssertEqual(t, len(decls), 0, "Did not expect Decls to contain", decls)

	dout := &bytes.Buffer{}
	gp.Dump(dout)
	dstr := dout.String()
	expDstr := "package main\n"
	testutil.AssertEqual(t, dstr, expDstr, "Expected '", expDstr, "' not '", dstr, "'")

	file := gp.File()
	testutil.AssertNotNil(t, file, "File() should not return nil")
	testutil.AssertNotNil(t, file.Name, "File.Name should not be nil")
	testutil.AssertNotNil(t, file.Name.Name, "main",
		"File.Name.Name should be", "main", file.Name.Name)
	testutil.AssertNotNil(t, file.Decls, "File.Decls() should not return nil")
	testutil.AssertEqual(t, len(file.Decls), 0,
		"Did not expect File.Decls to contain", file.Decls)
	testutil.AssertNotNil(t, file.Imports, "File.Imports() should not return nil")
	testutil.AssertEqual(t, len(file.Imports), 0,
		"Did not expect File.Imports to contain", file.Imports)
	testutil.AssertEqual(t, len(file.Unresolved), 0,
		"File.Unresolved should be empty but contains", len(file.Unresolved),
		"entries")
	testutil.AssertEqual(t, len(file.Comments), 0,
		"File.Comments should be empty but contains", len(file.Comments),
		"entries")

	fs := gp.FileSet()
	fmt.Printf("FileSet -> %v\n", fs)

	sout := &bytes.Buffer{}
	gp.WriteString(sout)
	fmt.Printf("File -> %v\n", sout.String())

}
