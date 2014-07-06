package parser

import (
	"fmt"
	"go/ast"
	"strings"
)

type VarType int

// bool
// byte rune char
// uint8 uint16 uint32 uint64
// int8 int16 int32 int64
// uintptr
// float32 float64
// complex64 complex128
// string
const (
	VT_ILLEGAL VarType = iota
	VT_VOID
	VT_BOOL
	VT_BYTE
	VT_CHAR
	VT_INT16
	VT_INT
	VT_INT64
	VT_FLOAT32
	VT_FLOAT64
	VT_STRING
	VT_GENERIC_OBJECT
	VT_ARRAY
	VT_MAP
	VT_INTERFACE
	VT_CLASS
)

func (vt VarType) String() string {
	switch vt {
	case VT_VOID: return "void"
	case VT_BOOL: return "bool"
	case VT_BYTE: return "byte"
	case VT_CHAR: return "char"
	case VT_INT16: return "int16"
	case VT_INT: return "int"
	case VT_INT64: return "int64"
	case VT_FLOAT32: return "float32"
	case VT_FLOAT64: return "float64"
	case VT_STRING: return "string"
	case VT_GENERIC_OBJECT: return "{}interface"
	case VT_ARRAY: return "??array??"
	case VT_MAP: return "??map??"
	case VT_INTERFACE: return "??interface??"
	case VT_CLASS: return "??class??"
	}

	return fmt.Sprintf("??VarType#%d??", vt)
}

type TypeDictionary interface {
	ImportedType(string) string
	IsInterface(string) bool
}

type TypeData struct {
	vtype VarType
	vclass string
	array_dims int
	type1 *TypeData
	type2 *TypeData
}

var genericObject = &TypeData{vtype: VT_GENERIC_OBJECT}
var voidType = &TypeData{vtype: VT_VOID}
var boolType = &TypeData{vtype: VT_BOOL}
var byteType = &TypeData{vtype: VT_BYTE}
var charType = &TypeData{vtype: VT_CHAR}
var shortType = &TypeData{vtype: VT_INT16}
var intType = &TypeData{vtype: VT_INT}
var longType = &TypeData{vtype: VT_INT64}
var floatType = &TypeData{vtype: VT_FLOAT32}
var doubleType = &TypeData{vtype: VT_FLOAT64}
var stringType = &TypeData{vtype: VT_STRING}

func NewTypeDataPrimitive(typename string, dims int) *TypeData {
	var noDimType *TypeData
	switch strings.ToLower(typename) {
	case "void":
		noDimType = voidType
	case "boolean":
		noDimType = boolType
	case "byte":
		noDimType = byteType
	case "char":
		noDimType = charType
	case "short":
		noDimType = shortType
	case "int":
		noDimType = intType
	case "long":
		noDimType = longType
	case "float":
		noDimType = floatType
	case "double":
		noDimType = doubleType
	case "string":
		noDimType = stringType
	default:
		panic(fmt.Sprintf("Unrecognized primitive type %v", typename))
	}

	if dims == 0 {
		return noDimType
	}

	return &TypeData{vtype: VT_ARRAY, type1: noDimType, array_dims: dims}
}

func NewTypeDataObject(tdict TypeDictionary, typename string, dims int) *TypeData {
	var vtype VarType

	imptype := tdict.ImportedType(typename)
	if imptype == "" {
		imptype = typename
		vtype = VT_CLASS
	} else if tdict.IsInterface(imptype) {
		vtype = VT_INTERFACE
	} else {
		vtype = VT_CLASS
	}

	td := &TypeData{vtype: vtype, vclass: imptype}
	if dims > 0 {
		return &TypeData{vtype: VT_ARRAY, array_dims: dims, type1: td}
	}

	return td
}

var identBool = ast.NewIdent("bool")
var identByte = ast.NewIdent("byte")
var identChar = ast.NewIdent("char")
var identInt16 = ast.NewIdent("int16")
var identInt = ast.NewIdent("int")
var identInt64 = ast.NewIdent("int64")
var identFloat32 = ast.NewIdent("float32")
var identFloat64 = ast.NewIdent("float64")
var identString = ast.NewIdent("string")

func (vdata *TypeData) Decl() ast.Expr {
	return vdata.Expr()
}

func (vdata *TypeData) Equals(odata *TypeData) bool {
	if vdata == nil || odata == nil {
		return false
	}

	if vdata.vtype != odata.vtype || vdata.array_dims != odata.array_dims {
		return false
	}

	if vdata.type1 != nil || odata.type1 != nil {
		if (vdata.type1 == nil && odata.type1 != nil) ||
			(vdata.type1 != nil && odata.type1 == nil) {
			return false
		}

		if !vdata.type1.Equals(odata.type1) {
			return false
		}
	}

	if vdata.type2 != nil || odata.type2 != nil {
		if (vdata.type2 == nil && odata.type2 != nil) ||
			(vdata.type2 != nil && odata.type2 == nil) {
			return false
		}

		if !vdata.type2.Equals(odata.type2) {
			return false
		}
	}

	return true
}

func (vdata *TypeData) Expr() ast.Expr {
	if vdata.vtype != VT_ARRAY {
		typename, is_nil := vdata.TypeName()
		if is_nil {
			return nil
		}

		expr := typename
		for i := 0; i < vdata.array_dims; i++ {
			expr = &ast.ArrayType{Elt: typename}
		}

		return expr
	}

	if vdata.type1 != nil {
		return &ast.ArrayType{Elt: vdata.type1.Expr()}
	}

	return &ast.ArrayType{Elt: genericObject.Expr()}
}

func (vdata *TypeData) IsClass(name string) bool {
	return vdata.vtype == VT_CLASS && vdata.vclass == name
}

func (vdata *TypeData) isObject() bool {
	return vdata.vtype == VT_INTERFACE || vdata.vtype == VT_CLASS
}

func (vdata *TypeData) Name() string {
	switch vdata.vtype {
	case VT_ARRAY:
		tstr := "array"
		if vdata.type1 == nil {
			tstr += "_Object"
		} else {
			tstr += "_"
			tstr += vdata.type1.String()
		}

		if vdata.array_dims > 0 {
			tstr += fmt.Sprintf("_dim%d", vdata.array_dims)
		}

		return tstr
	case VT_MAP:
		kstr := "map"
		if vdata.type1 == nil {
			kstr += "_Object"
		} else {
			kstr += "_"
			kstr += vdata.type1.String()
		}

		if vdata.type2 == nil {
			kstr += "_Object"
		} else {
			kstr += "_"
			kstr += vdata.type2.String()
		}

		return kstr
	case VT_INTERFACE:
		return vdata.vclass
	case VT_CLASS:
		return vdata.vclass
	default:
		break
	}

	return vdata.vtype.String()
}

func (vdata *TypeData) String() string {
	switch vdata.vtype {
	case VT_ARRAY:
		var tstr string
		if vdata.type1 == nil {
			tstr = genericObject.String()
		} else {
			tstr = vdata.type1.String()
		}

		for i := 0; i < vdata.array_dims; i++ {
			tstr = "[]" + tstr
		}

		return tstr
	case VT_MAP:
		var kstr string
		if vdata.type1 == nil {
			kstr = genericObject.String()
		} else {
			kstr = vdata.type1.String()
		}

		var vstr string
		if vdata.type2 == nil {
			vstr = genericObject.String()
		} else {
			vstr = vdata.type2.String()
		}

		return "map[" + kstr + "]" + vstr
	case VT_INTERFACE:
		return vdata.vclass
	case VT_CLASS:
		return "*" + vdata.vclass
	default:
		break
	}

	return vdata.vtype.String()
}

func (vdata *TypeData) TypeName() (ast.Expr, bool) {
	switch vdata.vtype {
	case VT_VOID:
		return nil, true
	case VT_BOOL:
		return identBool, false
	case VT_BYTE:
		return identByte, false
	case VT_CHAR:
		return identChar, false
	case VT_INT16:
		return identInt16, false
	case VT_INT:
		return identInt, false
	case VT_INT64:
		return identInt64, false
	case VT_FLOAT32:
		return identFloat32, false
	case VT_FLOAT64:
		return identFloat64, false
	case VT_STRING:
		return identString, false
	case VT_GENERIC_OBJECT:
		return &ast.InterfaceType{Methods: &ast.FieldList{}}, false
	case VT_ARRAY:
		return ast.NewIdent("<<array>>"), false
	case VT_MAP:
		return ast.NewIdent("<<map>>"), false
	case VT_INTERFACE:
		return ast.NewIdent(vdata.vclass), false
	case VT_CLASS:
		return &ast.StarExpr{X: ast.NewIdent(vdata.vclass)}, false
	default:
		break
	}

	panic(fmt.Sprintf("Unknown VarType %v", vdata.vtype))
}
