package parser

import (
	"fmt"
	"strings"
	"testing"

	"java2go/testutil"
)

func Test_TypeData_Primitive(t *testing.T) {
	allTypes := []string{
		"void", "boolean", "byte", "char", "short",
		"int", "long", "float", "double", "string",
	}

	for _, str := range allTypes {
		for dim := 0; dim < 5; dim++ {
			td := NewTypeDataPrimitive(str, dim)

			_, is_nil := td.TypeName()
			testutil.AssertEqual(t, dim == 0 && strings.EqualFold(str, "void"), is_nil,
				str, "should not be nil", is_nil)

			var expStr string
			switch str {
			case "boolean":
				expStr = "bool"
			case "short":
				expStr = "int16"
			case "long":
				expStr = "int64"
			case "float":
				expStr = "float32"
			case "double":
				expStr = "float64"
			default:
				expStr = str
			}
			for i := 1; i <= dim; i++ {
				expStr = "[]" + expStr
			}

			testutil.AssertEqual(t, td.String(), expStr, "Expected", expStr,
				"not", td.String())

			testutil.AssertFalse(t, td.isObject(), "Unexpected object", str)
		}
	}

	defer func() {
		if x := recover(); x != nil {
			testutil.AssertEqual(t, x, "Unrecognized primitive type xxx")
		}
	}()

	td := NewTypeDataPrimitive("xxx", 0)
	testutil.AssertNotNil(t, td, "Got nil TypeData for bad primitive type")
}

type FakeDictionary struct {
	dict map[string]bool
}

func (fd *FakeDictionary) ImportedType(typename string) string {
	if _, ok := fd.dict[typename]; ok {
		return typename
	}

	return ""
}

func (fd *FakeDictionary) IsInterface(typename string) bool {
	if val, ok := fd.dict[typename]; ok {
		return val
	}

	return false
}

func Test_TypeData_Object(t *testing.T) {
	fd := &FakeDictionary{dict: make(map[string]bool)}
	fd.dict["abc"] = true
	fd.dict["def"] = false

	for _, str := range []string{"abc", "def", "ghi"} {
		for dim := 0; dim < 5; dim++ {
			td := NewTypeDataObject(fd, str, dim)
			validate(t, str, dim, td, fd)
		}
	}
}

func validate(t *testing.T, str string, dim int, td *TypeData,
	fd *FakeDictionary) {
	vt, is_nil := td.TypeName()
	testutil.AssertEqual(t, dim == 0 && strings.EqualFold(str, "void"), is_nil,
		str, "should not be nil", is_nil)
	if !is_nil {
		testutil.AssertNotNil(t, vt, str, "dim", dim, "is nil ::", vt)
	}

	expName := str
	if dim == 0 {
		expName = str
	} else {
		var star string
		if fd != nil && !fd.IsInterface(str) {
			star = "*"
		}
		expName = fmt.Sprintf("array_%s%s_dim%d", star, str, dim)
	}
	testutil.AssertEqual(t, expName, td.Name(), "Expected name", expName,
		"not", td.Name())

	var expStr string
	if fd == nil || fd.IsInterface(str) {
		expStr = str
	} else {
		expStr = "*" + str
	}
	for i := 1; i <= dim; i++ {
		expStr = "[]" + expStr
	}

	testutil.AssertEqual(t, expStr, td.String(), "Expected string ", expStr,
		"not", td.String())

	if fd != nil {
		if dim == 0 {
			testutil.AssertTrue(t, td.isObject(), "Expected object", str)
		} else {
			testutil.AssertFalse(t, td.isObject(), "Unexpected object", str)
		}
	}
}
