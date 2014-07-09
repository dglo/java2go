package testutil

import (
	"testing"
)

func AssertEmpty(t *testing.T, val string, args ...interface{}) {
	if val != "" {
		t.Fatal(args)
	}
}

func AssertEqual(t *testing.T, a interface{}, b interface{},
	args ...interface{}) {
	if a != b {
		t.Fatal(args)
	}
}

func AssertFalse(t *testing.T, val bool, args ...interface{}) {
	if val {
		t.Fatal(args)
	}
}

func AssertNil(t *testing.T, val interface{}, args ...interface{}) {
	if val != nil {
		t.Fatal(args)
	}
}

func AssertNotNil(t *testing.T, val interface{}, args ...interface{}) {
	if val == nil {
		t.Fatal(args)
	}
}

func AssertTrue(t *testing.T, val bool, args ...interface{}) {
	if !val {
		t.Fatal(args)
	}
}
