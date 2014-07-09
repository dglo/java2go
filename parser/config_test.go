package parser

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/dglo/java2go/testutil"
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
	is_iface := cfg.isInterface("foo")
	testutil.AssertFalse(t, is_iface, "isInterface() returned true")
	pkg := cfg.packageName("foo")
	testutil.AssertEmpty(t, pkg, "Package() returned", pkg)
	rcvr := cfg.receiver("foo")
	testutil.AssertEmpty(t, rcvr, "Receiver() returned", rcvr)
	str := cfg.String()
	if !strings.HasPrefix(str, "Config[") || !strings.HasSuffix(str, "]") {
		t.Fatal("String() returned", str)
	}
}

func Test_Config_Full(t *testing.T) {
	name := fillFile(t)

	cfg := ReadConfig(name)

	var pkg string
	pkg = cfg.packageName("a.b")
	testutil.AssertEqual(t, pkg, "ab")
	pkg = cfg.packageName("a.b.c")
	testutil.AssertEqual(t, pkg, "abc")

	var is_iface bool
	is_iface = cfg.isInterface("a.b.DEF")
	testutil.AssertTrue(t, is_iface, "a.b.DEF is not an interface")
	is_iface = cfg.isInterface("a.b.GHI")
	testutil.AssertTrue(t, is_iface, "a.b.GHI is not an interface")

	var rcvr string
	rcvr = cfg.receiver("a.b.XXX")
	testutil.AssertEqual(t, rcvr, "xxx")
	rcvr = cfg.receiver("a.b.ZZZ")
	testutil.AssertEqual(t, rcvr, "zzz")
}
