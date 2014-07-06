package parser

import (
	"testing"
)

func assertEmpty(t *testing.T, val string, args ...interface{}) {
	if val != "" {
		t.Fatal(args)
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{},
	args ...interface{}) {
	if a != b {
		t.Fatal(args)
	}
}

func assertFalse(t *testing.T, val bool, args ...interface{}) {
	if val {
		t.Fatal(args)
	}
}

func assertNil(t *testing.T, val interface{}, args ...interface{}) {
	if val != nil {
		t.Fatal(args)
	}
}

func assertNotNil(t *testing.T, val interface{}, args ...interface{}) {
	if val == nil {
		t.Fatal(args)
	}
}

func assertTrue(t *testing.T, val bool, args ...interface{}) {
	if !val {
		t.Fatal(args)
	}
}
