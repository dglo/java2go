package parser

import (
	"go/ast"
	"go/token"
	"fmt"
	"log"
	"strings"

	"github.com/dglo/java2go/grammar"
)

// list of Java classes which inherit from AbstractList
var javaListType = []string { "List", "ArrayList", "LinkedList", "Stack",
	"Vector" }

type TransformFunc func(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool)

type GoObject interface {
	RunTransform(xform TransformFunc, gp *GoProgram, cls GoClass,
		parent GoObject) (GoObject, bool)
}

type GoTypedObject interface {
	GoObject
	VarType() *TypeData
}

func convertToBlock(obj GoObject) (*GoBlock, error) {
	if blk, ok := obj.(*GoBlock); ok {
		return blk, nil
	}

	return nil, fmt.Errorf("%v<%T> is not a *GoBlock", obj, obj)
}

func convertToExpr(obj GoObject) (GoExpr, error) {
	if expr, ok := obj.(GoExpr); ok {
		return expr, nil
	}

	return nil, fmt.Errorf("%v<%T> is not a GoExpr", obj, obj)
}

func convertToMethod(obj GoObject) (GoMethod, error) {
	if mthd, ok := obj.(GoMethod); ok {
		return mthd, nil
	}

	return nil, fmt.Errorf("%v<%T> is not a GoMethod", obj, obj)
}

func convertToMethodArgs(obj GoObject) (*GoMethodArguments, error) {
	if mthd, ok := obj.(*GoMethodArguments); ok {
		return mthd, nil
	}

	return nil, fmt.Errorf("%v<%T> is not a *GoMethodArguments", obj, obj)
}

func convertToStmt(obj GoObject) (GoStatement, error) {
	if stmt, ok := obj.(GoStatement); ok {
		return stmt, nil
	}

	return nil, fmt.Errorf("%v<%T> is not a GoStatement", obj, obj)
}

func convertToVar(obj GoObject) (GoVar, error) {
	if govar, ok := obj.(GoVar); ok {
		return govar, nil
	}

	return nil, fmt.Errorf("%v<%T> is not a GoVar", obj, obj)
}

func convertToVarInit(obj GoObject) (*GoVarInit, error) {
	if vi, ok := obj.(*GoVarInit); ok {
		return vi, nil
	}

	return nil, fmt.Errorf("%v<%T> is not a *GoVarInit", obj, obj)
}

func getFmtClass(prog *GoProgram) GoClass {
	fmtcls := prog.findClass("fmt")
	if fmtcls == nil {
		fmtcls = NewGoFakeClass("fmt")
		prog.addClass(fmtcls)

		// make sure "fmt" is imported
		prog.addImport("fmt", "")
	}

	return fmtcls
}

type GoPkgName struct {
	pkg string
	name string
}

func (gpn *GoPkgName) Expr() ast.Expr {
	return &ast.SelectorExpr{X: ast.NewIdent(gpn.pkg),
		Sel: ast.NewIdent(gpn.name)}
}

func (gpn *GoPkgName) hasVariable(govar GoVar) bool {
	return false
}

func (gpn *GoPkgName) Init() ast.Stmt {
	return nil
}

func (gpn *GoPkgName) RunTransform(xform TransformFunc, prog *GoProgram,
	cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, gpn)
}

func (gpn *GoPkgName) String() string {
	return fmt.Sprintf("GoPkgName[%s|%s]", gpn.pkg, gpn.name)
}

func (gpn *GoPkgName) VarType() *TypeData {
	panic("GoPkgName.VarType() unimplemented")
}

// transform "array.length" to "len(array)"
func TransformArrayLen(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {
	var ref *GoClassAttribute
	var ok bool
	if ref, ok = object.(*GoClassAttribute); !ok {
		return nil, true
	}

	if ref.Suffix() != "length" || ref.VarType().vtype != VT_ARRAY {
		return nil, true
	}

	fm := NewGoFakeMethod(nil, "len", intType)
	args := &GoMethodArguments{args: []GoExpr{ ref.govar, }}

	return &GoMethodAccess{method: fm, args: args}, false
}

// transform "System.out.print*(...)" to "fmt.Print*(...)" and
// "System.err.print*(...)" to "fmt.Fprintf(os.Stderr, ...)"
func TransformSysfile(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {
	var ma *GoMethodAccess
	var ok bool
	if ma, ok = object.(*GoMethodAccess); !ok {
		return nil, true
	}

	if ma.method == nil {
		return nil, true
	}

	mcls := ma.method.Class()
	if mcls == nil || mcls.IsNil() {
		return nil, true
	}

	clsname := mcls.Name()
	name := ma.method.Name()

	if !strings.HasPrefix(clsname, "System.") ||
		!strings.HasPrefix(name, "print") {
		return nil, true
	}

	fmtcls := getFmtClass(prog)

	var fixed string
	var args *GoMethodArguments
	if strings.HasSuffix(clsname, ".out") {
		fixed = strings.ToUpper(name[:1]) + name[1:]
		args = ma.args
	} else {
		fixed = "F" + name
		arglist := make([]GoExpr, 1)
		arglist[0] = &GoPkgName{pkg: "os", name: "Stderr"}
		arglist = append(arglist, ma.args.args...)
		args = &GoMethodArguments{args: arglist}

		// make sure "os" is imported
		prog.addImport("os", "")
	}

	fm := NewGoFakeMethod(fmtcls, fixed, nil)

	return &GoMethodAccess{method: fm, args: args}, false
}

// transform "func main(String[] xxx)" to "func main()" and all "xxx" references
// to "os.Args"
func TransformMainArgs(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {
	var cm *GoClassMethod
	var ok bool
	if cm, ok = object.(*GoClassMethod); !ok {
		return nil, true
	}

	if cm.method_type != mt_main {
		return nil, true
	}

	if cm.params == nil || len(cm.params) != 1 {
		return nil, true
	}

	arg := cm.params[0]

	if cm.body.hasVariable(arg) {
		arg.SetGoName("os.Args")

		// make sure "os" is imported
		prog.addImport("os", "")
	}

	cm.params = nil

	return nil, true
}

// transform "foo(this)" to "foo(rcvr)"
func TransformThisArg(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {
	var mref *GoMethodReference
	var ok bool
	if mref, ok = object.(*GoMethodReference); !ok {
		return nil, true
	}

	if mref.args == nil {
		return nil, true
	}

	for i, arg := range mref.args.args {
		var kwd *GoKeyword
		if kwd, ok = arg.(*GoKeyword); ok {
			if kwd.token == grammar.THIS {
				mref.args.args[i] =
					NewFakeVar(prog.Receiver(cls.Name()), nil, 0)
			}
		}
	}

	return nil, true
}

// transform method calls for List variants to appropriate array operations
func TransformListMethods(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {
	var mref *GoMethodAccessVar
	var ok bool
	if mref, ok = object.(*GoMethodAccessVar); !ok {
		return nil, true
	}

	var is_list bool
	for _, n := range javaListType {
		if mref.govar.VarType().IsClass(n) {
			is_list = true
			break
		}
	}
	if !is_list {
		return nil, true
	}

	switch mref.method.Name() {
	case "add":
		if len(mref.args.args) != 1 {
			log.Printf("//ERR// Cannot convert %v add() with %d args\n",
				mref.govar.VarType(), len(mref.args.args))
			return nil, true
		}

		apnd := NewGoFakeMethod(nil, "append", mref.govar.VarType())
		args := &GoMethodArguments{args: []GoExpr{ mref.govar,
			mref.args.args[0]}}

		rhs := []GoExpr{ &GoMethodAccess{method: apnd, args: args}, }

		return &GoAssign{govar: mref.govar, tok: token.ASSIGN, rhs: rhs}, false
	case "isEmpty":
		if len(mref.args.args) != 0 {
			log.Printf("//ERR// Cannot convert %v isEmpty() with %d args\n",
				mref.govar.VarType(), len(mref.args.args))
			return nil, true
		}

		args := &GoMethodArguments{args: []GoExpr{ mref.govar, }}
		x := &GoMethodAccess{method: NewGoFakeMethod(nil, "len", intType),
			args: args}
		return &GoBinaryExpr{x: x, op: token.EQL, y: &GoLiteral{text: "0"}},
		false
	case "get":
		if len(mref.args.args) != 1 {
			log.Printf("//ERR// Cannot convert %v get() with %d args\n",
				mref.govar.VarType(), len(mref.args.args))
			return nil, true
		}

		return &GoArrayReference{govar: mref.govar, index: mref.args.args[0]},
		false
	case "size":
		fm := NewGoFakeMethod(nil, "len", intType)
		args := &GoMethodArguments{args: []GoExpr{ mref.govar, }}

		return &GoMethodAccess{method: fm, args: args}, false
	default:
		log.Printf("//ERR// Not converting %v method %v\n",
			mref.govar.VarType(), mref.method.Name())
		break
	}

	return nil, true
}

// transform various toString() calls into fmt.Sprintf
func TransformToString(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {

	var macc *GoMethodAccess
	var ok bool
	if macc, ok = object.(*GoMethodAccess); !ok {
		return nil, true
	}

	if macc.method.Name() != "toString" {
		return nil, true
	}

	if len(macc.args.args) == 1 {
		alist := make([]GoExpr, 2)
		alist[0] = NewGoLiteral("\"%v\"")
		alist[1] = macc.args.args[0]

		args := &GoMethodArguments{args: alist}

		fmtcls := getFmtClass(prog)

		mthd := fmtcls.FindMethod("Sprintf", args)
		if mthd == nil {
			mthd = NewGoFakeMethod(fmtcls, "Sprintf", stringType)
			fmtcls.AddMethod(mthd)
		}

		return &GoMethodAccess{method: mthd, args: args}, false
	}

	log.Printf("//ERR// Cannot convert %v toString() with %d args\n",
		macc.method, len(macc.args.args))
	return nil, true
}

// transform String.format method call into fmt.Sprintf
func TransformStringFormat(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {

	var mref *GoMethodReference
	var ok bool
	if mref, ok = object.(*GoMethodReference); !ok {
		return nil, true
	}

	if mref.name == "format" && !mref.class.IsNil() &&
		mref.class.Name() == "String" {
		fmtcls := getFmtClass(prog)

		return NewGoMethodReference(fmtcls, "Sprintf", mref.args, false), false
	}

	return nil, true
}

// transform string addition into fmt.Sprintf
func TransformStringAddition(parent GoObject, prog *GoProgram, cls GoClass,
	object GoObject) (GoObject, bool) {

	var bex *GoBinaryExpr
	var ok bool
	if bex, ok = object.(*GoBinaryExpr); !ok {
		return nil, true
	}

	var x GoTypedObject
	if x, ok = bex.x.(GoTypedObject); !ok {
		return nil, true
	}

	if x.VarType() != stringType {
		return nil, true
	}

	return transformBinaryStringExpression(prog, bex)
}

func joinStrings(s1 string, s2 string) string {
	if s1[0] != '"' || s1[len(s1)-1] != '"' {
		panic(fmt.Sprintf("String %v is not quoted", s1))
	}
	if s2[0] != '"' || s2[len(s2)-1] != '"' {
		panic(fmt.Sprintf("String %v is not quoted", s2))
	}
	s := fmt.Sprintf("\"%s%s\"", s1[1:len(s1)-1], s2[1:len(s2)-1])
	return s
}

func transformBinaryStringExpression(prog *GoProgram, bex *GoBinaryExpr) (GoObject, bool) {
	if bex.op != token.ADD {
		return nil, true
	}

	switch x := bex.x.(type) {
	case *GoLiteral:
		if !x.isString() {
			// ignore non-string addition
			return nil, true
		}

		if y, ok := bex.y.(*GoLiteral); ok {
			if y.isString() {
				return NewGoLiteral(joinStrings(x.text, y.text)), false
			}
		}
	case *GoMethodAccess:
		if x.method.VarType() != stringType {
			return nil, true
		}

		if x.method.Name() == "Sprintf" {
			if fmtstr, ok := x.args.args[0].(*GoLiteral); !ok {
				log.Printf("//ERR// First Sprintf argument should be " +
					" *GoLiteral not %T\n", x.args.args[0])
				return nil, true
			} else {
				x.args.args[0] = NewGoLiteral(joinStrings("\"%v\"",
					fmtstr.text))
			}
			x.args.args = append(x.args.args, bex.y)
			return x, false
		}
	}

	fmtcls := getFmtClass(prog)

	fmtstr := NewGoLiteral("\"%v%v\"")
	args := &GoMethodArguments{args: []GoExpr{fmtstr, bex.x, bex.y}}

	mthd := fmtcls.FindMethod("Sprintf", args)
	if mthd == nil {
		mthd = NewGoFakeMethod(fmtcls, "Sprintf", stringType)
		fmtcls.AddMethod(mthd)
	}

	return &GoMethodAccess{method: mthd, args: args}, false
}

// list of standard transformation rules
var StandardRules = []TransformFunc {
	TransformArrayLen,
	TransformSysfile,
	TransformMainArgs,
	TransformThisArg,
	TransformListMethods,
	TransformToString,
	TransformStringAddition,
	TransformStringFormat,
}
