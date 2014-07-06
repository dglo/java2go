package parser

import (
	"io/ioutil"
	"strings"
	"testing"
)

func fillFile(t *testing.T) string {
	f, err := ioutil.TempFile("", "tstcfg")
	if err != nil {
		t.Fatal("Cannot create temporary file")
	}

	f.WriteString("PACKAGE a.b -> ab\n")
	f.WriteString("PACKAGE a.b.c -> abc\n")
	f.WriteString("INTERFACE a.b.DEF\n")
	f.WriteString("INTERFACE a.b.GHI\n")
	f.WriteString("RECEIVER a.b.XXX -> xxx\n")
	f.WriteString("RECEIVER a.b.ZZZ -> zzz\n")
	f.Close()
	return f.Name()
}

func Test_Config_Empty(t *testing.T) {
	cfg := &Config{}
	is_iface := cfg.IsInterface("foo")
	assertFalse(t, is_iface, "IsInterface() returned true")
	pkg := cfg.Package("foo")
	assertEmpty(t, pkg, "Package() returned", pkg)
	rcvr := cfg.Receiver("foo")
	assertEmpty(t, rcvr, "Receiver() returned", rcvr)
	str := cfg.String()
	if !strings.HasPrefix(str, "Config[") || !strings.HasSuffix(str, "]") {
		t.Fatal("String() returned", str)
	}
}

func Test_Config_Full(t *testing.T) {
	name := fillFile(t)

	cfg := ReadConfig(name)

	var pkg string
	pkg = cfg.Package("a.b")
	assertEqual(t, pkg, "ab")
	pkg = cfg.Package("a.b.c")
	assertEqual(t, pkg, "abc")

	var is_iface bool
	is_iface = cfg.IsInterface("a.b.DEF")
	assertTrue(t, is_iface, "a.b.DEF is not an interface")
	is_iface = cfg.IsInterface("a.b.GHI")
	assertTrue(t, is_iface, "a.b.GHI is not an interface")

	var rcvr string
	rcvr = cfg.Receiver("a.b.XXX")
	assertEqual(t, rcvr, "xxx")
	rcvr = cfg.Receiver("a.b.ZZZ")
	assertEqual(t, rcvr, "zzz")
}
