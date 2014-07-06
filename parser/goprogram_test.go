package parser

import (
	"bytes"
	"fmt"
	"testing"
)

func Test_FixName(t *testing.T) {
	raw := "NaMe"

	modmap := make(map[string]string)
	modmap["public"] = "NaMe"
	modmap["protected"] = "NaMe"
	modmap["private"] = "naMe"

	for modtype, fixed := range modmap {
		jmod := NewJModifiers(modtype, nil)
		nm := fixName(raw, jmod)
		assertEqual(t, nm, fixed, "For modtype", modtype, "expected", fixed,
			"not", nm)
	}
}

func Test_GoProgram_Basic(t *testing.T) {
	name := "foo"
	gp := NewGoProgram(name, nil, false)

	gp.Analyze(nil)

	rcvr := gp.Receiver("foo")
	assertEqual(t, rcvr, "rcvr", "Expected rcvr, not", rcvr)

	imap := gp.Imports()
	assertNotNil(t, imap, "ImportMap() should not return nil")
	assertEqual(t, len(imap), 0, "Did not expect ImportMap to contain", imap)

	decls := gp.Decls()
	assertNotNil(t, decls, "Decls() should not return nil")
	assertEqual(t, len(decls), 0, "Did not expect Decls to contain", decls)

	dout := &bytes.Buffer{}
	gp.Dump(dout)
	dstr := dout.String()
	expDstr := "package main\n"
	assertEqual(t, dstr, expDstr, "Expected '", expDstr, "' not '", dstr, "'")

	file := gp.File()
	assertNotNil(t, file, "File() should not return nil")
	assertNotNil(t, file.Name, "File.Name should not be nil")
	assertNotNil(t, file.Name.Name, "main",
		"File.Name.Name should be", "main", file.Name.Name)
	assertNotNil(t, file.Decls, "File.Decls() should not return nil")
	assertEqual(t, len(file.Decls), 0,
		"Did not expect File.Decls to contain", file.Decls)
	assertNotNil(t, file.Imports, "File.Imports() should not return nil")
	assertEqual(t, len(file.Imports), 0,
		"Did not expect File.Imports to contain", file.Imports)
	assertEqual(t, len(file.Unresolved), 0,
		"File.Unresolved should be empty but contains", len(file.Unresolved),
		"entries")
	assertEqual(t, len(file.Comments), 0,
		"File.Comments should be empty but contains", len(file.Comments),
		"entries")

	fs := gp.FileSet()
	fmt.Printf("FileSet -> %v\n", fs)

	sout := &bytes.Buffer{}
	gp.WriteString(sout)
	fmt.Printf("File -> %v\n", sout.String())

}
