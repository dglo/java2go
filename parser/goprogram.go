package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"log"
	"math"
	"os"
	"path"
	//"runtime/debug"
	"sort"
	"strings"

	"github.com/dglo/java2go/dumper"
	"github.com/dglo/java2go/grammar"
)

// assign new struct to receiver
func assignNewStruct(rvar GoVar, class GoMethodOwner) *GoAssign {
	rhs := make([]GoExpr, 1)
	rhs[0] = &GoReference{cls: class}

	return &GoAssign{govar: rvar, tok: token.ASSIGN, rhs: rhs}
}

func fixName(name string, modifiers *grammar.JModifiers) string {
	if name != "" && modifiers != nil {
		if modifiers.IsSet(grammar.ModPrivate) {
			return strings.ToLower(name[:1]) + name[1:]
		} else if modifiers.IsSet(grammar.ModPublic) {
			return strings.ToUpper(name[:1]) + name[1:]
		}
	}

	return name
}

func singleStatement(name string, stmts []ast.Stmt) (ast.Stmt, bool) {
	if stmts == nil || len(stmts) == 0 {
		return nil, true
	} else if len(stmts) != 1 {
		panic(fmt.Sprintf("%s can only handle 1 statement, not %d",
			name, len(stmts)))
	}

	return stmts[0], false
}

type FakeVar struct {
	name string
	dims int
}

func NewFakeVar(name string, type_args []*grammar.JTypeArgument, dims int) *FakeVar {
	if type_args != nil && len(type_args) > 0 {
		log.Printf("//ERR// Not handling type_args in %v\n", name)
	}

	return &FakeVar{name: name, dims: dims}
}

func (fv *FakeVar) Equals(govar GoVar) bool {
	return govar == fv
}

func (fv *FakeVar) Expr() ast.Expr {
	var expr ast.Expr

	expr = fv.Ident()
	for i := 0; i < fv.dims; i++ {
		expr = &ast.ArrayType{Elt: expr}
	}

	return expr
}

func (fv *FakeVar) GoName() string {
	return fv.name
}

func (fv *FakeVar) hasVariable(govar GoVar) bool {
	return fv.Equals(govar)
}

func (fv *FakeVar) Ident() *ast.Ident {
	return ast.NewIdent(fv.name)
}

func (fv *FakeVar) Init() ast.Stmt {
	return nil
}

func (fv *FakeVar) IsClassField() bool {
	return false
}

func (fv *FakeVar) IsFinal() bool {
	return false
}

func (fv *FakeVar) IsStatic() bool {
	return false
}

func (fv *FakeVar) Name() string {
	return fv.name
}

func (fv *FakeVar) Receiver() string {
	return ""
}

func (fv *FakeVar) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, fv)
}

func (fv *FakeVar) SetGoName(newname string) {
	fv.name = newname
}

func (fv *FakeVar) String() string {
	return fmt.Sprintf("%s[dims=%d]", fv.name, fv.dims)
}

func (fv *FakeVar) Type() ast.Expr {
	return nil
}

func (fv *FakeVar) VarType() *TypeData {
	return nil
}

type FileManager struct {
	set *token.FileSet
	file *token.File
	off int
}

func NewFileManager(name string) *FileManager {
	mgr := &FileManager{set: token.NewFileSet()}
	mgr.file = mgr.set.AddFile(name, -1, math.MaxInt32)
	return mgr
}

func (mgr *FileManager) FileSet() *token.FileSet {
	return mgr.set
}

func (mgr *FileManager) NextPos() token.Pos {
	off := mgr.off
	mgr.file.AddLine(off)
	mgr.off += 10
	return mgr.file.Pos(off)
}

type GoArrayExpr interface {
	GoExpr
}

type GoArrayAlloc struct {
	typedata *TypeData
	args []GoExpr
}

func (aa *GoArrayAlloc) Expr() ast.Expr {
	args := make([]ast.Expr, len(aa.args) + 1)
	args[0] = &ast.ArrayType{Elt: aa.typedata.Expr()}
	if aa.args != nil && len(aa.args) > 0 {
		for i, arg := range aa.args {
			args[i + 1] = arg.Expr()
		}
	}

	return &ast.CallExpr{Fun: ast.NewIdent("make"), Args: args}
}

func (aa *GoArrayAlloc) hasVariable(govar GoVar) bool {
	for _, a := range aa.args {
		if a.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (aa *GoArrayAlloc) Init() ast.Stmt {
	return nil
}

func (aa *GoArrayAlloc) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, a := range aa.args {
		obj, is_nil := a.RunTransform(xform, prog, cls, aa)
		if !is_nil {
			var err error
			if aa.args[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, aa)
}

func (aa *GoArrayAlloc) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoArrayAlloc[")
	b.WriteString(aa.typedata.String())
	b.WriteString("|")
	if aa.args != nil && len(aa.args) > 0 {
		for i, a := range aa.args {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(a.String())
		}
	}
	b.WriteString("]")
	return b.String()
}

func (aa *GoArrayAlloc) VarType() *TypeData {
	return aa.typedata
}

type GoArrayInit struct {
	typedata *TypeData
	elems []GoExpr
}

func (ai *GoArrayInit) Expr() ast.Expr {
	elements := make([]ast.Expr, len(ai.elems))
	if ai.elems != nil && len(ai.elems) > 0 {
		for i, elem := range ai.elems {
			elements[i] = elem.Expr()
		}
	}

	return &ast.CompositeLit{Type: ai.typedata.Expr(), Elts: elements}
}

func (ai *GoArrayInit) hasVariable(govar GoVar) bool {
	for _, e := range ai.elems {
		if e.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (ai *GoArrayInit) Init() ast.Stmt {
	return nil
}

func (ai *GoArrayInit) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, e := range ai.elems {
		obj, is_nil := e.RunTransform(xform, prog, cls, ai)
		if !is_nil {
			var err error
			if ai.elems[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, ai)
}

func (ai *GoArrayInit) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoArrayInit[")
	b.WriteString(ai.typedata.String())
	b.WriteString("|")
	if ai.elems != nil && len(ai.elems) > 0 {
		for i, e := range ai.elems {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(e.String())
		}
	}
	b.WriteString("]")
	return b.String()
}

func (ai *GoArrayInit) VarType() *TypeData {
	return ai.typedata
}

type GoArrayReference struct {
	govar GoVar
	obj GoExpr
	index GoExpr
}

func (nae *GoArrayReference) Equals(govar GoVar) bool {
	if nae == govar || (nae.govar != nil && nae.govar.Equals(govar)) {
		return true
	}

	return false
}

func (nae *GoArrayReference) Expr() ast.Expr {
	return nae.IndexExpr()
}

func (nae *GoArrayReference) hasVariable(govar GoVar) bool {
	if nae.govar != nil && nae.govar.Equals(govar) {
		return true
	}

	if nae.obj != nil && nae.obj.hasVariable(govar) {
		return true
	}

	if nae.index != nil && nae.index.hasVariable(govar) {
		return true
	}

	return false
}

func (nae *GoArrayReference) GoName() string {
	if nae.govar != nil {
		return nae.govar.GoName()
	}

	return nae.obj.String()
}

func (nae *GoArrayReference) Ident() *ast.Ident {
	if nae.govar != nil {
		return nae.govar.Ident()
	}

	return nil
}

func (nae *GoArrayReference) IndexExpr() *ast.IndexExpr {
	var name ast.Expr
	if nae.obj == nil {
		name = nae.govar.Ident()
		if name == nil {
			panic(fmt.Sprintf("GoArrayReference govar %v returns nil Ident()",
				nae.govar))
		}
	} else {
		name = nae.obj.Expr()
		if name == nil {
			panic(fmt.Sprintf("GoArrayReference obj %v returns nil Expr()",
				nae.obj))
		}
	}

	return &ast.IndexExpr{X: name, Index: nae.index.Expr()}
}

func (nae *GoArrayReference) Init() ast.Stmt {
	return nil
}

func (nae *GoArrayReference) IsClassField() bool {
	return false
}

func (nae *GoArrayReference) IsFinal() bool {
	return false
}

func (nae *GoArrayReference) IsStatic() bool {
	return false
}

func (nae *GoArrayReference) Name() string {
	if nae.obj == nil {
		return nae.govar.Name()
	}

	return nae.obj.String()
}

func (nae *GoArrayReference) Receiver() string {
	if nae.obj == nil {
		return nae.govar.Receiver()
	}

	return ""
}

func (nae *GoArrayReference) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if nae.govar != nil {
		obj, is_nil := nae.govar.RunTransform(xform, prog, cls, nae)
		if !is_nil {
			var err error
			if nae.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if nae.obj != nil {
		obj, is_nil := nae.obj.RunTransform(xform, prog, cls, nae)
		if !is_nil {
			var err error
			if nae.obj, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if nae.index != nil {
		obj, is_nil := nae.index.RunTransform(xform, prog, cls, nae)
		if !is_nil {
			var err error
			if nae.index, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, nae)
}

func (nae *GoArrayReference) SetGoName(newname string) {
	if nae.govar != nil {
		nae.govar.SetGoName(newname)
	}

	panic("Cannot change Go name of object")
}

func (nae *GoArrayReference) String() string {
	var vstr string
	if nae.govar != nil {
		vstr = nae.govar.String()
	}

	var ostr string
	if nae.obj != nil {
		ostr = nae.obj.String()
	}

	var istr string
	if nae.index != nil {
		istr = nae.index.String()
	}

	return "GoArrayReference[" + vstr + "|" + ostr + "|" + istr + "]"
}

func (nae *GoArrayReference) Type() ast.Expr {
	return nil
}

func (nae *GoArrayReference) VarType() *TypeData {
	return nil
}

type GoAssign struct {
	govar GoVar
	tok token.Token
	rhs []GoExpr
}

func (asgn *GoAssign) Expr() ast.Expr {
	if len(asgn.rhs) != 1 {
		panic("Cannot assign multiple expressions to a single variable")
	}

	return &ast.BinaryExpr{X: asgn.govar.Expr(), Op: token.ASSIGN,
		Y: asgn.rhs[0].Expr()}
}

func (asgn *GoAssign) hasVariable(govar GoVar) bool {
	if asgn.govar.Equals(govar) {
		return true
	}

	for _, e := range asgn.rhs {
		if e.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (asgn *GoAssign) Init() ast.Stmt {
	if stmt, is_nil := singleStatement("Assignment", asgn.Stmts()); !is_nil {
		return stmt
	}

	return nil
}

func (asgn *GoAssign) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	obj, is_nil := asgn.govar.RunTransform(xform, prog, cls, parent)
	if !is_nil {
		var err error
		if asgn.govar, err = convertToVar(obj); err != nil {
			panic(err)
		}
	}

	for i, r := range asgn.rhs {
		obj, is_nil := r.RunTransform(xform, prog, cls, asgn)
		if !is_nil {
			var err error
			if asgn.rhs[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, asgn)
}

func (asgn *GoAssign) Stmts() []ast.Stmt {
	lhs := make([]ast.Expr, 1)
	lhs[0] = asgn.govar.Expr()

	rhs := make([]ast.Expr, len(asgn.rhs))
	if asgn.rhs != nil {
		for i, expr := range asgn.rhs {
			rhs[i] = expr.Expr()
		}
	}

	return []ast.Stmt { &ast.AssignStmt{Lhs: lhs, Tok: asgn.tok, Rhs: rhs}, }
}

func (asgn *GoAssign) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoAssign[")
	b.WriteString(asgn.govar.String())
	b.WriteString("|")
	b.WriteString(asgn.tok.String())
	b.WriteString("|")
	if asgn.rhs != nil && len(asgn.rhs) > 0 {
		for i, r := range asgn.rhs {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(r.String())
		}
	}
	b.WriteString("]")
	return b.String()
}

func (asgn *GoAssign) VarType() *TypeData {
	return asgn.govar.VarType()
}

type GoBinaryExpr struct {
	x GoExpr
	op token.Token
	y GoExpr
}

func (bex *GoBinaryExpr) BinaryExpr() *ast.BinaryExpr {
	x := bex.x.Expr()

	if bex.op == token.SHR {
		xargs := make([]ast.Expr, 1)
		xargs[0] = x

		log.Printf("WARNING: >>> replacement casting %v to uint32\n", x)

		x = &ast.CallExpr{Fun: ast.NewIdent("uint32"), Args: xargs}
	}

	return &ast.BinaryExpr{X: x, Op: bex.op, Y: bex.y.Expr()}
}

func (bex *GoBinaryExpr) Expr() ast.Expr {
	return bex.BinaryExpr()
}

func (bex *GoBinaryExpr) hasVariable(govar GoVar) bool {
	if bex.x != nil && bex.x.hasVariable(govar) {
		return true
	}

	if bex.y != nil && bex.y.hasVariable(govar) {
		return true
	}

	return false
}

func (bex *GoBinaryExpr) Init() ast.Stmt {
	return nil
}

func (bex *GoBinaryExpr) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if bex.x != nil {
		obj, is_nil := bex.x.RunTransform(xform, prog, cls, bex)
		if !is_nil {
			var err error
			if bex.x, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if bex.y != nil {
		obj, is_nil := bex.y.RunTransform(xform, prog, cls, bex)
		if !is_nil {
			var err error
			if bex.y, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, bex)
}

func (bex *GoBinaryExpr) String() string {
	return "GoBinaryExpr[" + bex.x.String() + "|" + bex.op.String() + "|" +
		bex.y.String() + "]"
}

func (bex *GoBinaryExpr) VarType() *TypeData {
	xt := bex.x.VarType()
	yt := bex.y.VarType()

	if xt == nil {
		return yt
	}

	return xt
}

type GoBlock struct {
	stmts []GoStatement
}

func (blk *GoBlock) BlockStmt() *ast.BlockStmt {
	list := make([]ast.Stmt, 0)
	for _, stmt := range blk.stmts {
		ss := stmt.Stmts()
		if ss != nil {
			list = append(list, ss...)
		}
	}

	return &ast.BlockStmt{List: list}
}

func (blk *GoBlock) hasVariable(govar GoVar) bool {
	for _, stmt := range blk.stmts {
		if stmt.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (blk *GoBlock) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, s := range blk.stmts {
		obj, is_nil := s.RunTransform(xform, prog, cls, blk)
		if !is_nil {
			var err error
			if blk.stmts[i], err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, blk)
}

func (blk *GoBlock) Stmts() []ast.Stmt {
	return []ast.Stmt { blk.BlockStmt(), }
}

func (blk *GoBlock) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoBlock[")
	if blk.stmts != nil && len(blk.stmts) > 0 {
		for i, s := range blk.stmts {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(s.String())
		}
	}
	return b.String()
}

func (blk *GoBlock) VarType() *TypeData {
	panic("GoBlock.VarType() unimplemented")
}

type GoBranchStmt struct {
	tok token.Token
	label string
}

func (bs *GoBranchStmt) hasVariable(govar GoVar) bool {
	return false
}

func (bs *GoBranchStmt) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, bs)
}

func (bs *GoBranchStmt) Stmts() []ast.Stmt {
	var label *ast.Ident
	if bs.label != "" {
		label = ast.NewIdent(bs.label)
	}

	return []ast.Stmt { &ast.BranchStmt{Tok: bs.tok, Label: label}, }
}

func (bs *GoBranchStmt) String() string {
	return "GoBranchStmt[" + bs.tok.String() + "|" + bs.label + "]"
}

func (bs *GoBranchStmt) VarType() *TypeData {
	panic("GoBranchStmt.VarType() unimplemented")
}

type GoCastType struct {
	target GoExpr
	casttype *TypeData
}

func (cast *GoCastType) Expr() ast.Expr {
	return cast.TypeExpr()
}

func (cast *GoCastType) hasVariable(govar GoVar) bool {
	if cast.target != nil && cast.target.hasVariable(govar) {
		return true
	}

	return false
}

func (cast *GoCastType) Init() ast.Stmt {
	return nil
}

func (cast *GoCastType) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if cast.target != nil {
		obj, is_nil := cast.target.RunTransform(xform, prog, cls, cast)
		if !is_nil {
			var err error
			if cast.target, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, cast)
}

func (cast *GoCastType) String() string {
	return "GoCastType[" + cast.target.String() + "|" +
		cast.casttype.String() + "]"
}

func (cast *GoCastType) TypeExpr() *ast.TypeAssertExpr {
	return &ast.TypeAssertExpr{X: cast.target.Expr(),
		Type: cast.casttype.Expr()}
}

func (cast *GoCastType) VarType() *TypeData {
	return cast.casttype
}

type GoClass interface {
	GoObject
	AddConstant(con *GoConstant)
	AddMethod(mthd GoMethod)
	Constants() []ast.Decl
	Decls() []ast.Decl
	finalize(gp *GoProgram)
	FindMethod(name string, args *GoMethodArguments) GoMethod
	findVariable(typename *grammar.JTypeName) GoVar
	IsNil() bool
	IsReference() bool
	Name() string
	Parent() GoMethodOwner
	Statics() []ast.Decl
	Super() GoMethodOwner
	String() string
	WriteString(out io.Writer, verbose bool)
}

func makeClassKey(cls GoClass) string {
	return makeClassKeyFromParts(cls.Parent(), cls.Name())
}

func makeClassKeyFromParts(parent GoMethodOwner, name string) string {
	if parent == nil {
		return name
	}

	return parent.Name() + "." + name
}

type GoClassAlloc struct {
	class GoClass
	method GoMethod
	type_args []*TypeData
	args []GoExpr
	body []GoStatement
}

func (gca *GoClassAlloc) Expr() ast.Expr {
	var funexpr ast.Expr
	var args []ast.Expr

	funexpr = ast.NewIdent(gca.method.GoName())

	if gca.args != nil && len(gca.args) > 0 {
		args = make([]ast.Expr, len(gca.args))

		for i, arg := range gca.args {
			args[i] = arg.Expr()
		}
	}

	// ignoring class body

	return &ast.CallExpr{Fun: funexpr, Args: args}
}

func (gca *GoClassAlloc) hasVariable(govar GoVar) bool {
	for _, a := range gca.args {
		if a.hasVariable(govar) {
			return true
		}
	}

	for _, b := range gca.body {
		if b.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (gca *GoClassAlloc) Init() ast.Stmt {
	return nil
}

func (gca *GoClassAlloc) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, a := range gca.args {
		obj, is_nil := a.RunTransform(xform, prog, cls, gca)
		if !is_nil {
			var err error
			if gca.args[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	for i, b := range gca.body {
		obj, is_nil := b.RunTransform(xform, prog, cls, gca)
		if !is_nil {
			var err error
			if gca.body[i], err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gca)
}

func (gca *GoClassAlloc) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoClassAlloc[")
	gca.class.WriteString(b, false)
	b.WriteString("|")
	gca.method.WriteString(b)
	b.WriteString("|")
	if gca.type_args != nil && len(gca.type_args) > 0 {
		for i, ta := range gca.type_args {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(ta.String())
		}
	}
	b.WriteString("|")
	if gca.args != nil && len(gca.args) > 0 {
		for i, aa := range gca.args {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(aa.String())
		}
	}
	b.WriteString("|")
	if gca.body != nil && len(gca.body) > 0 {
		for i, bb := range gca.body {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(bb.String())
		}
	}
	b.WriteString("]")
	return b.String()
}

func (gca *GoClassAlloc) VarType() *TypeData {
	return gca.method.VarType()
}

type GoClassDefinition struct {
	program *GoProgram
	parent GoMethodOwner
	super GoClass
	name string
	constants []*GoConstant
	statics []*GoStatic
	vars []*GoVarInit
	interfaces []GoInterface
	methods *classMethodMap
}

func NewGoClassDefinition(program *GoProgram, parent GoMethodOwner,
	name string) *GoClassDefinition {
	return &GoClassDefinition{program: program, parent: parent, name: name,
		methods: NewClassMethodMap()}
}

func (cls *GoClassDefinition) AddConstant(con *GoConstant) {
	if cls.constants == nil {
		cls.constants = make([]*GoConstant, 0)
	}
	cls.constants = append(cls.constants, con)
}

func (cls *GoClassDefinition) AddMethod(newmthd GoMethod) {
	cls.methods.AddMethod(newmthd, cls.methods)
}

func (cls *GoClassDefinition) AddNewMethod(gs *GoState, mth *grammar.JMethodDecl) {
	cls.AddMethod(NewGoClassMethod(cls, gs, mth))
}

func (cls *GoClassDefinition) addVar(v *GoVarInit) {
	if v.govar.IsFinal() && v.expr != nil {
		if _, is_literal := v.expr.(*GoLiteral); !is_literal {
			if cls.program.verbose {
				log.Printf("//ERR// Ignoring non-literal constant %v for %v\n",
					v.expr, v.govar)
			} else {
				log.Printf("//ERR// Ignoring non-literal constant\n")
			}
		} else {
			con := &GoConstant{name: v.govar.GoName(),
				typedata: v.govar.VarType(), init: v}
			if cls.constants == nil {
				cls.constants = make([]*GoConstant, 0)
			}
			cls.constants = append(cls.constants, con)
			return
		}
	}

	if v.govar.IsStatic() {
		stat := &GoStatic{init: v}
		if cls.statics == nil {
			cls.statics = make([]*GoStatic, 0)
		}
		cls.statics = append(cls.statics, stat)
		return
	}

	if cls.vars == nil {
		cls.vars = make([]*GoVarInit, 0)
	}
	cls.vars = append(cls.vars, v)
}

func (cls *GoClassDefinition) Constants() []ast.Decl {
	if cls.constants == nil || len(cls.constants) == 0 {
		return nil
	}

	decls := make([]ast.Decl, len(cls.constants))
	for i, con := range cls.constants {
		decls[i] = con.Decl()
	}

	return decls
}

func (cls *GoClassDefinition) createConstructor(gp *GoProgram) *GoClassMethod {
	gs := &GoState{program: gp, class: cls}

	// initialize receiver variable
	rcvr := gs.addVariable(gs.Receiver(), nil, 0, nil, false)

	// create reference to receiver
	rhs := make([]GoExpr, 1)
	rhs[0] = &GoReference{cls: cls}

	// create receiver assignment statement and final 'return'
	stmts := make([]GoStatement, 2)
	stmts[0] = &GoAssign{govar: rcvr, tok: token.ASSIGN, rhs: rhs}
	if cls.super != nil {
		if cls.program.verbose {
			log.Printf("//ERR// Not creating %v superclass %v initializer\n",
				cls.name, cls.super.Name())
		} else {
			log.Printf("//ERR// Not creating superclass initializer\n")
		}
	}
	stmts[1] = &GoReturn{}

	// create the constructor body
	body := &GoBlock{stmts: stmts}

	// create the constructor method
	m := &GoClassMethod{class: cls, name: cls.name, goname: "New" + cls.name,
		rcvr: rcvr, method_type: mt_constructor, body: body}

	// add new constructor
	cls.AddMethod(m)

	return m
}

func (cls *GoClassDefinition) Decls() []ast.Decl {
	specs := make([]ast.Spec, 1)
	specs[0] = &ast.TypeSpec{Name: ast.NewIdent(cls.name),
		Type: cls.struct_type()}

	decls := make([]ast.Decl, 1)
	decls[0] = &ast.GenDecl{Tok: token.TYPE, Specs: specs}

	for _, key := range cls.methods.SortedKeys() {
		for _, m := range cls.methods.MethodList(key) {
			d2 := m.Decl()
			if d2 != nil {
				decls = append(decls, d2)
			}
		}
	}

	return decls
}

func (cls *GoClassDefinition) finalize(gp *GoProgram) {
	// move variable initialization code inside constructors
	cls.internalizeVarInits(gp)
	// renumber duplicate methods
	cls.renumberDuplicateMethods(gp)
}

func findAssigned(unasgned_vars []*GoVarInit, gv GoVar) int {
	for i, v := range unasgned_vars {
		if v == nil {
			continue
		}

		if v.govar.Equals(gv) {
			unasgned_vars[i] = nil
			return i
		}
	}

	return -1
}

func (cls *GoClassDefinition) FindMethod(name string,
	args *GoMethodArguments) GoMethod {
	return cls.methods.FindMethod(name, args)
}

func (cls *GoClassDefinition) findVariable(typename *grammar.JTypeName) GoVar {
	for _, c := range cls.constants {
		if c.name == typename.String() {
			return c
		}
	}

	for _, i := range cls.interfaces {
		govar := i.findVariable(typename)
		if govar != nil {
			return govar
		}
	}

	if cls.super != nil {
		return cls.super.findVariable(typename)
	}

	return nil
}

func (cls *GoClassDefinition) internalizeVarInits(gp *GoProgram) {
	// build list of constructors
	ctors := make([]GoMethod, 0)
	for _, key := range cls.methods.SortedKeys() {
		for _, m := range cls.methods.MethodList(key) {
			if m.MethodType() == mt_constructor {
				ctors = append(ctors, m)
			}
		}
	}

	// fix this()/super() in all ctors
	for _, m := range ctors {
		stmts := m.Body().stmts

		sawCTOR := false
		for _, s := range stmts {
			if _, nok := s.(*GoNewStruct); nok {
				sawCTOR = true
			}
		}

		if cls.super != nil && !sawCTOR {
			if cls.program.verbose {
				log.Printf("//ERR// Not adding %v superclass %v initializer\n",
					cls.name, cls.super.Name())
			} else {
				log.Printf("//ERR// Not adding superclass initializer\n")
			}
		}

		m.Body().stmts = stmts
	}

	// if there are no constructors, create one
	if len(ctors) == 0 {
		m := cls.createConstructor(gp)
		ctors = append(ctors, m)
	}

	// build list of initialized variables
	init_vars := make([]*GoVarInit, 0)
	for _, v := range cls.vars {
		if v.expr != nil {
			init_vars = append(init_vars, v)
		}
	}

	// if nothing is initialized, we're done
	if len(init_vars) > 0 {
		// add initializers to all constructors
		for i, m := range ctors {
			if gp.verbose {
				log.Printf("//ERR// ## CTOR#%d: %v\n", i, m)
			}

			// initially assume all variables need to be initialized
			num_unasgned := len(init_vars)
			unasgned_vars := make([]*GoVarInit, num_unasgned)
			for i, v := range init_vars {
				unasgned_vars[i] = v
			}

			for _, s := range m.Body().stmts {
				if asgn, aok := s.(*GoAssign); aok {
					// ignore non-assignments
					if asgn.tok != token.ASSIGN {
						continue
					}

					// ignore receiver init
					if asgn.govar == m.Receiver() {
						continue
					}

					if i := findAssigned(unasgned_vars, asgn.govar); i >= 0 {
						num_unasgned -= 1
					}
				}
			}

			if num_unasgned == 0 {
				continue
			}

			for i, s := range m.Body().stmts {
				if asgn, aok := s.(*GoAssign); aok {
					if asgn.govar == m.Receiver() && asgn.tok == token.ASSIGN {
						//  create a temporary list with the initial statements
						tmp := append(make([]GoStatement, 0),
							m.Body().stmts[:i+1]...)

						// add all struct variable initialization
						for _, v := range unasgned_vars {
							if v != nil {
								tmp = append(tmp,
									v.createInitializer(m.Receiver()))
							}
						}

						// append remaining statements
						tmp = append(tmp, m.Body().stmts[i+1:]...)

						// use the updated list as the method's body
						m.Body().stmts = tmp

						break
					}
				}
			}
		}
	}
}

func (cls *GoClassDefinition) IsNil() bool {
	return false
}

func (cls *GoClassDefinition) IsReference() bool {
	return false
}

func (cls *GoClassDefinition) Name() string {
	return cls.name
}

func (cls *GoClassDefinition) Parent() GoMethodOwner {
	return cls.parent
}

func (cls *GoClassDefinition) renumberDuplicateMethods(gp *GoProgram) {
	cls.methods.renumberDuplicateMethods(gp)
}

func (cd *GoClassDefinition) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, c := range cd.constants {
		obj, is_nil := c.RunTransform(xform, prog, cd, cd)
		if !is_nil {
			if con, ok := obj.(*GoConstant); ok {
				cd.constants[i] = con
			} else {
				panic(fmt.Errorf("%v<%T> is not a *GoConstant", obj, obj))
			}
		}
	}

	for i, s := range cd.statics {
		obj, is_nil := s.RunTransform(xform, prog, cd, cd)
		if !is_nil {
			if sta, ok := obj.(*GoStatic); ok {
				cd.statics[i] = sta
			} else {
				panic(fmt.Errorf("%v<%T> is not a *GoStatic", obj, obj))
			}
		}
	}

	for i, v := range cd.vars {
		obj, is_nil := v.RunTransform(xform, prog, cd, cd)
		if !is_nil {
			var err error
			if cd.vars[i], err = convertToVarInit(obj); err != nil {
				panic(err)
			}
		}
	}

	for _, key := range cd.methods.SortedKeys() {
		mlist := cd.methods.MethodList(key)
		for i, m := range mlist {
			obj, is_nil := m.RunTransform(xform, prog, cd, cd)
			if !is_nil {
				var err error
				if mlist[i], err = convertToMethod(obj); err != nil {
					panic(err)
				}
			}
		}
	}

	return xform(parent, prog, cd, cd)
}

func (cls *GoClassDefinition) Statics() []ast.Decl {
	if cls.statics == nil || len(cls.statics) == 0 {
		return nil
	}

	decls := make([]ast.Decl, len(cls.statics))
	for i, stat := range cls.statics {
		decls[i] = stat.Decl()
	}

	return decls
}

func (cls *GoClassDefinition) String() string {
    return fmt.Sprintf("%s{%d methods}", cls.name, cls.methods.Length())
}

func (cls *GoClassDefinition) struct_type() *ast.StructType {
	flds := make([]*ast.Field, 0)
	if cls.super != nil {
		stype := &ast.StarExpr{X: ast.NewIdent(cls.super.Name())}
		flds = append(flds, &ast.Field{Type: stype})
	}
	for _, v := range cls.vars {
		flds = append(flds, makeField(v.govar.GoName(), v.govar.Type()))
	}

	return &ast.StructType{Fields: &ast.FieldList{List: flds}}
}

func (cls *GoClassDefinition) Super() GoMethodOwner {
	return cls.super
}

func (cls *GoClassDefinition) WriteString(out io.Writer, verbose bool) {
	io.WriteString(out, "GoClassDefinition[")
	io.WriteString(out, cls.name)

	io.WriteString(out, "|")
	if !verbose {
		io.WriteString(out, fmt.Sprintf("%d vars", len(cls.vars)))
	} else {
		for i, v := range cls.vars {
			if i > 0 {
				io.WriteString(out, ",")
			}
			io.WriteString(out, v.String())
		}
	}

	io.WriteString(out, "|")
	cls.methods.WriteString(out, verbose)
	io.WriteString(out, "]")
}

type GoClassMethod struct {
	class GoMethodOwner
	name string
	goname string
	typedata *TypeData
	rcvr GoVar
	method_type methodType
	params []GoVar
	body *GoBlock
}

func NewGoClassMethod(class GoMethodOwner, gs *GoState, jmth *grammar.JMethodDecl) *GoClassMethod {
	var mtype methodType
	if jmth.Modifiers.HasAnnotation("Test") {
		mtype = mt_test
	} else if jmth.Name == class.Name() {
		mtype = mt_constructor
	} else if jmth.Modifiers.IsSet(grammar.ModStatic) {
		mtype = mt_static
	} else {
		mtype = mt_method
	}

	name := jmth.Name
	goname := fixName(jmth.Name, jmth.Modifiers)
	if mtype == mt_constructor {
		name = "New" + name
		goname = "New" + goname
	} else if mtype == mt_static && goname == "Main" {
		goname = "main"
		mtype = mt_main
	}

	gs2 := NewGoState(gs)

	var params []GoVar
	if mtype == mt_test {
		if jmth.FormalParams != nil && len(jmth.FormalParams) > 0 {
			if gs.Program().verbose {
				log.Printf("//ERR// Test method %s.%s should not have %d params\n",
					class.Name(), jmth.Name, len(jmth.FormalParams))
			} else {
				log.Printf("//ERR// Ignoring test method params\n")
			}
		}
	} else if jmth.FormalParams != nil && len(jmth.FormalParams) > 0 {
		params = make([]GoVar, len(jmth.FormalParams))

		for i, fp := range jmth.FormalParams {
			if fp.TypeSpec != nil {
				if fp.Dims != 0 {
					if gs.Program().verbose {
						log.Printf("//ERR// Ignoring %s dims=%d for %s.%s\n",
							fp.Name, fp.Dims, class.Name(), jmth.Name)
					} else {
						log.Printf("//ERR// Ignoring mthd param dims\n")
					}
				} else if fp.DotDotDot {
					if gs.Program().verbose {
						log.Printf("//ERR// Ignoring %s DotDotDot=true for %s.%s\n",
							fp.Name, class.Name(), jmth.Name)
					} else {
						log.Printf("//ERR// Ignoring mthd param...\n")
					}
				}

				govar := gs2.addVariable(fp.Name, fp.Modifiers, fp.Dims,
					fp.TypeSpec, false)



				params[i] = govar
			}
		}
	}

	var rvar GoVar
	if mtype == mt_constructor || mtype == mt_method {
		rvar = gs2.addVariable(gs.Receiver(), nil, 0, nil, false)
	}

	var typedata *TypeData
	if jmth.TypeSpec != nil &&
		(mtype == mt_method || mtype == mt_static) {
		typedata = gs.Program().createTypeData(jmth.TypeSpec.Name,
			jmth.TypeSpec.TypeArgs, jmth.TypeSpec.Dims)
	}

	body := analyzeBlock(gs2, class, jmth.Block)

	mthd := &GoClassMethod{class: class, name: name, goname: goname,
		typedata: typedata, rcvr: rvar, method_type: mtype, params: params,
		body: body}

	if mtype == mt_test {
		// make sure program imports 'testing' package
		gs.Program().addImport("testing", "")
	}

	if mtype == mt_constructor {
		// fix this()/super()
		has_this := false
		if len(mthd.body.stmts) >= 1 {
			if xstmt, xok := mthd.body.stmts[0].(*GoExprStmt); xok {
				if macc, mok := xstmt.x.(*GoMethodAccessKeyword); mok {
					if !macc.is_super {
						newstmt := NewGoNewStruct(rvar, class,
							false, macc.args)
						mthd.body.stmts[0] = newstmt
						has_this = true
					} else if class.Super() != nil {
						newstmt := NewGoNewStruct(rvar, class.Super(),
							true, macc.args)
						mthd.body.stmts[0] = newstmt
					} else {
						// super() called on class without a superclass!
						mthd.body.stmts[0] = assignNewStruct(rvar, class)
						has_this = true
					}
				}
			}
		}

		if !has_this {
			list := []GoStatement { assignNewStruct(rvar, class) }

			if mthd.body == nil {
				mthd.body = &GoBlock{stmts: list}
			} else {
				mthd.body.stmts = append(list, mthd.body.stmts...)
			}
		}

		// append final 'return'
		mthd.body.stmts = append(mthd.body.stmts, &GoReturn{})
	}

	return mthd
}

func (mthd *GoClassMethod) Arguments() []GoVar {
	return mthd.params
}

func (mthd *GoClassMethod) Body() *GoBlock {
	return mthd.body
}

func (mthd *GoClassMethod) Class() GoMethodOwner {
	return mthd.class
}

func (mthd *GoClassMethod) Decl() ast.Decl {
	mtype := &ast.FuncType{Params: mthd.paramList(), Results: mthd.results()}

	return &ast.FuncDecl{Name: ast.NewIdent(mthd.goname),
		Recv: mthd.recv(), Type: mtype, Body: mthd.body.BlockStmt()}
}

func (mthd *GoClassMethod)  Field() *ast.Field {
	return nil
}

func (mthd *GoClassMethod) GoName() string {
	return mthd.goname
}

func (mthd *GoClassMethod) HasArguments(args *GoMethodArguments) bool {
	if len(mthd.Arguments()) != args.Length() {
		return false
	}

	for i, arg := range mthd.Arguments() {
		if !arg.VarType().Equals(args.args[i].VarType()) {
			return false
		}
	}

	return true
}

func (gcm *GoClassMethod) IsMethod(mthd GoMethod) bool {
	return mthd.Name() == gcm.name &&
		len(mthd.Arguments()) == len(gcm.params)
}

func (mthd *GoClassMethod) MethodType() methodType {
	return mthd.method_type
}

func (mthd *GoClassMethod) Name() string {
	return mthd.name
}

func (mthd *GoClassMethod) NumParameters() int {
	return len(mthd.params)
}

func (mthd *GoClassMethod) paramList() *ast.FieldList {
	if mthd.method_type == mt_test {
		selexp := &ast.SelectorExpr{X: ast.NewIdent("testing"),
			Sel: ast.NewIdent("T")}

		pfld := makeField("t", &ast.StarExpr{X: selexp})

		paramList := make([]*ast.Field, 1)
		paramList[0] = pfld

		return &ast.FieldList{List: paramList}
	}

	if mthd.params != nil && len(mthd.params) > 0 {
		flist := make([]*ast.Field, len(mthd.params))

		for i, fp := range mthd.params {
			flist[i] = makeField(fp.Name(), fp.Type())
		}

		return &ast.FieldList{List: flist}
	}

	return nil
}

func (mthd *GoClassMethod) Receiver() GoVar {
	return mthd.rcvr
}

func (mthd *GoClassMethod) recv() *ast.FieldList {
	if mthd.method_type != mt_method {
		// non-methods don't have a receiver
		return nil
	}

	rlist := make([]*ast.Field, 1)
	rlist[0] = makeField(mthd.rcvr.GoName(),
		&ast.StarExpr{X: ast.NewIdent(mthd.class.Name())})
	return &ast.FieldList{List: rlist}
}

func (mthd *GoClassMethod) results() *ast.FieldList {
	switch mthd.method_type {
	case mt_constructor:
		// return the result receiver
		rlist := make([]*ast.Field, 1)
		rlist[0] = makeField(mthd.rcvr.GoName(),
			&ast.StarExpr{X: ast.NewIdent(mthd.class.Name())})
		return &ast.FieldList{List: rlist}
	case mt_main:
		fallthrough
	case mt_static:
		fallthrough
	case mt_method:
		if mthd.typedata != nil {
			if typename, is_nil := mthd.typedata.TypeName(); !is_nil {
				rlist := make([]*ast.Field, 1)
				rlist[0] = makeField("", typename)
				return &ast.FieldList{List: rlist}
			}
		}
	}

	return nil
}

func (mthd *GoClassMethod) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, p := range mthd.params {
		obj, is_nil := p.RunTransform(xform, prog, cls, mthd)
		if !is_nil {
			var err error
			if mthd.params[i], err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if mthd.body != nil {
		obj, is_nil := mthd.body.RunTransform(xform, prog, cls, mthd)
		if !is_nil {
			var err error
			if mthd.body, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, mthd)
}

func (gcm *GoClassMethod) SetGoName(newname string) {
	gcm.goname = newname
}

func (gcm *GoClassMethod) SetOriginal(gcm2 *GoClassMethod) {
	panic("Unimplemented")
}

func (mthd *GoClassMethod) String() string {
	var pstr string
	if mthd.params != nil && len(mthd.params) > 0 {
		b := &bytes.Buffer{}
		for i, prm := range mthd.params {
			if i > 0 {
				io.WriteString(b, ",")
			}
			io.WriteString(b, prm.String())
		}
		pstr = b.String()
	}

	return fmt.Sprintf("GoClassMethod[%s|%s|%s]", mthd.name, mthd.goname, pstr)
}

func (mthd *GoClassMethod) VarType() *TypeData {
	return mthd.typedata
}

func (mthd *GoClassMethod) WriteString(out io.Writer) {
	io.WriteString(out, "GoClassMethod[")
	io.WriteString(out, mthd.name)
	io.WriteString(out, "|")
	io.WriteString(out, mthd.goname)
	io.WriteString(out, "|")
	if mthd.typedata != nil {
		io.WriteString(out, mthd.typedata.String())
	}
	io.WriteString(out, "|")
	if mthd.rcvr != nil {
		io.WriteString(out, mthd.rcvr.String())
	}
	io.WriteString(out, "|")
	io.WriteString(out, mthd.method_type.String())
	io.WriteString(out, "|")
	for i, prm := range mthd.params {
		if i > 0 {
			io.WriteString(out, ",")
		}
		io.WriteString(out, prm.String())
	}
	io.WriteString(out, "|")
	io.WriteString(out, mthd.body.String())
	io.WriteString(out, "]")
}

type GoClassReference struct {
	name string
	parent GoMethodOwner
	cls *GoClassDefinition
}

func (cref *GoClassReference) AddConstant(con *GoConstant) {
	if cref.parent == nil {
		panic(fmt.Sprintf("ClassReference %v has no parent", cref.name))
	}

	cref.parent.AddConstant(con)
}

func (cref *GoClassReference) AddMethod(mthd GoMethod) {
	if cref.parent == nil {
		panic(fmt.Sprintf("ClassReference %v has no parent", cref.name))
	}

	cref.parent.AddMethod(mthd)
}

func (cref *GoClassReference) Constants() []ast.Decl {
	if cref.parent == nil {
		panic(fmt.Sprintf("ClassReference %v has no parent", cref.name))
	}

	return cref.parent.Constants()
}

func (cref *GoClassReference) Decls() []ast.Decl {
	if cref.cls == nil {
		return nil
	}

	return cref.cls.Decls()
}

func (cref *GoClassReference) finalize(gp *GoProgram) {
	// do nothing
}

func (cref *GoClassReference) FindMethod(name string,
	args *GoMethodArguments) GoMethod {
	if cref.parent == nil {
		panic(fmt.Sprintf("ClassReference %v has no parent", cref.name))
	}

	return cref.parent.FindMethod(name, args)
}

func (cref *GoClassReference) findVariable(typename *grammar.JTypeName) GoVar {
	panic("Unimplemented")
}

func (cref *GoClassReference) IsNil() bool {
	return false
}

func (cref *GoClassReference) IsReference() bool {
	return true
}

func (cref *GoClassReference) Name() string {
	return cref.name
}

func (cref *GoClassReference) Parent() GoMethodOwner {
	return cref.parent
}

func (cref *GoClassReference) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, cref)
}

func (cref *GoClassReference) Statics() []ast.Decl {
	if cref.parent == nil {
		panic(fmt.Sprintf("ClassReference %v has no parent", cref.name))
	}

	return cref.parent.Statics()
}

func (cref *GoClassReference) String() string {
	b := &bytes.Buffer{}
	cref.WriteString(b, false)
	return b.String()
}

func (cref *GoClassReference) Super() GoMethodOwner {
	if cref.cls == nil {
		return nil
	}

	return cref.cls.Super()
}

func (cref *GoClassReference) WriteString(out io.Writer, verbose bool) {
	io.WriteString(out, "GoClassReference[")
	io.WriteString(out, cref.name)
	io.WriteString(out, "|")
	if cref.cls != nil {
		cref.cls.WriteString(out, verbose)
	}
	io.WriteString(out, "]")
}

type GoConstant struct {
	name string
	typedata *TypeData
	init *GoVarInit
}

func (con *GoConstant) Decl() ast.Decl {
	vals := []ast.Expr{ con.init.Expr(), }
	vspec := &ast.ValueSpec{Names: []*ast.Ident{ ast.NewIdent(con.name), },
		Values: vals}
	return &ast.GenDecl{Tok: token.CONST, Specs: []ast.Spec{ vspec, }}
}

func (con *GoConstant) Expr() ast.Expr {
	return ast.NewIdent(con.name)
}

func (con *GoConstant) GoName() string {
	panic("Unimplemented")
}

func (con *GoConstant) Equals(govar GoVar) bool {
	panic("Unimplemented")
}

func (con *GoConstant) hasVariable(govar GoVar) bool {
	return false
}

func (con *GoConstant) Ident() *ast.Ident {
	panic("Unimplemented")
}

func (con *GoConstant) Init() ast.Stmt {
	panic("Unimplemented")
}

func (con *GoConstant) IsClassField() bool {
	panic("Unimplemented")
}

func (con *GoConstant) IsFinal() bool {
	panic("Unimplemented")
}

func (con *GoConstant) IsStatic() bool {
	panic("Unimplemented")
}

func (con *GoConstant) Name() string {
	return con.name
}

func (con *GoConstant) Receiver() string {
	return ""
}

func (con *GoConstant) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if con.init != nil {
		obj, is_nil := con.init.RunTransform(xform, prog, cls, con)
		if !is_nil {
			var err error
			if con.init, err = convertToVarInit(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, con)
}

func (con *GoConstant) SetGoName(newname string) {
	panic("Unimplemented")
}

func (con *GoConstant) String() string {
	b := &bytes.Buffer{}
	con.WriteString(b)
	return b.String()
}

func (con *GoConstant) Type() ast.Expr {
	panic("Unimplemented")
}

func (con *GoConstant) VarType() *TypeData {
	return con.typedata
}

func (con *GoConstant) WriteString(out io.Writer) {
	io.WriteString(out, "GoConstant[")
	io.WriteString(out, con.name)

	io.WriteString(out, "|")
	io.WriteString(out, con.typedata.String())

	io.WriteString(out, "|")
	if con.init != nil {
		io.WriteString(out, con.init.String())
	}

	io.WriteString(out, "]")
}

type GoNewStruct struct {
	rcvr GoVar
	cls GoMethodOwner
	is_super bool
	args *GoMethodArguments
}

func NewGoNewStruct(rcvr GoVar, cls GoMethodOwner, is_super bool,
	args *GoMethodArguments) *GoNewStruct {
	return &GoNewStruct{rcvr: rcvr, cls: cls, is_super: is_super, args: args}
}

func (gsc *GoNewStruct) hasVariable(govar GoVar) bool {
	if gsc.args != nil && gsc.args.hasVariable(govar) {
		return true
	}

	return false
}

func (gsc *GoNewStruct) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gsc.args != nil {
		obj, is_nil := gsc.args.RunTransform(xform, prog, cls, gsc)
		if !is_nil {
			var err error
			if gsc.args, err = convertToMethodArgs(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gsc)
}

func (gsc *GoNewStruct) Stmts() []ast.Stmt {
	lhs := make([]ast.Expr, 1)
	if gsc.is_super {
		lhs[0] = &ast.SelectorExpr{X: gsc.rcvr.Ident(),
			Sel: ast.NewIdent(gsc.cls.Name())}
	} else {
		lhs[0] = gsc.rcvr.Ident()
	}

	var args []ast.Expr
	if gsc.args != nil {
		args = gsc.args.ExprList()
	}

	rhs := make([]ast.Expr, 1)
	rhs[0] = &ast.CallExpr{Fun: ast.NewIdent("New" + gsc.cls.Name()),
		Args: args}

	return []ast.Stmt { &ast.AssignStmt{Lhs: lhs, Tok: token.ASSIGN,
		Rhs: rhs}, }
}

func (gsc *GoNewStruct) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoNewStruct[")
	b.WriteString(gsc.cls.String())
	b.WriteString("|")
	if gsc.is_super {
		b.WriteString("is_super")
	}
	b.WriteString("|")
	gsc.args.WriteString(b)
	b.WriteString("]")
	return b.String()
}

type GoEmpty struct {
}

func NewGoEmpty() *GoEmpty {
	return &GoEmpty{}
}

type GoEnumConstant struct {
	name string
}

func (ec *GoEnumConstant) ValueSpec(num int, typeident *ast.Ident) *ast.ValueSpec {
	var vals []ast.Expr
	if num == 0 {
		vals = []ast.Expr{ ast.NewIdent("iota"), }
	}

	if typeident == nil {
		return &ast.ValueSpec{Names: []*ast.Ident{ ast.NewIdent(ec.name), },
			Values: vals}
	}

	return &ast.ValueSpec{Names: []*ast.Ident{ ast.NewIdent(ec.name), },
		Type: typeident, Values: vals}
}

func (ec *GoEnumConstant) WriteString(out io.Writer) {
	io.WriteString(out, "GoEnumConstant[")
	io.WriteString(out, ec.name)
	io.WriteString(out, "]")
}

// enumSlice is a []*GoEnumDefinition wrapper for sort.Sort()
type enumSlice []*GoEnumDefinition
func (p enumSlice) Len() int { return len(p) }
func (p enumSlice) Less(i, j int) bool { return p[i].name < p[j].name }
func (p enumSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type GoEnumDefinition struct {
	name string
	constants []*GoEnumConstant
}

func (enum *GoEnumDefinition) Decls(gp *GoProgram) []ast.Decl {
	typeident := ast.NewIdent(enum.name)

	tspecs := []ast.Spec{ &ast.TypeSpec{Name: typeident,
		Type: ast.NewIdent("int")}}
	tdecl := &ast.GenDecl{Tok: token.TYPE, Specs: tspecs}

	decls := []ast.Decl{ tdecl, }

	specs := make([]ast.Spec, len(enum.constants))

	lpos := gp.mgr.NextPos()
	for i, e := range enum.constants {
		ti := typeident
		if i > 0 {
			ti = nil
		}

		specs[i] = e.ValueSpec(i, ti)
		gp.mgr.NextPos()
	}
	rpos := gp.mgr.NextPos()

	gen := &ast.GenDecl{Tok: token.CONST, Lparen: lpos, Specs: specs,
		Rparen: rpos}
	decls = append(decls, gen)

	return decls
}

func (enum *GoEnumDefinition) WriteString(out io.Writer) {
	io.WriteString(out, "GoEnumDefinition[")
	io.WriteString(out, enum.name)
	io.WriteString(out, "|")

	if enum.constants != nil && len(enum.constants) > 0 {
		for i, c := range enum.constants {
			if i > 0 {
				io.WriteString(out, ",")
			}

			c.WriteString(out)
		}
	}

	io.WriteString(out, "]")
}

type GoExpr interface {
	GoObject
	Expr() ast.Expr
	hasVariable(govar GoVar) bool
	Init() ast.Stmt
	String() string
	VarType() *TypeData
}

type GoExprStmt struct {
	x GoExpr
}

func (exst *GoExprStmt) hasVariable(govar GoVar) bool {
	return exst.x.hasVariable(govar)
}

func (exst *GoExprStmt) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	obj, is_nil := exst.x.RunTransform(xform, prog, cls, exst)
	if !is_nil {
		var err error
		if exst.x, err = convertToExpr(obj); err != nil {
			panic(err)
		}
	}

	return xform(parent, prog, cls, exst)
}

func (exst *GoExprStmt) Stmts() []ast.Stmt {
	return []ast.Stmt { &ast.ExprStmt{X: exst.x.Expr()}, }
}

func (exst *GoExprStmt) String() string {
	return "GoExprStmt[" + exst.x.String() + "]"
}

func (exst *GoExprStmt) VarType() *TypeData {
	return exst.x.VarType()
}

type GoFakeClass struct {
	pkg string
	name string
	methods map[string]GoMethod
}

func NewGoFakeClass(name string) *GoFakeClass {
	return &GoFakeClass{name: name}
}

func (gfc *GoFakeClass) AddConstant(con *GoConstant) {
	panic("unimplemented")
}

func (gfc *GoFakeClass) AddMethod(mthd GoMethod) {
	if gfc.methods == nil {
		gfc.methods = make(map[string]GoMethod)
	}

	if _, ok := gfc.methods[mthd.Name()]; ok {
		if gfc.name != "fmt" {
			// don't warn about variadic functions
			log.Printf("//ERR// Adding multiple %v methods to %v\n",
				mthd.Name(), gfc.name)
		}
	}

	gfc.methods[mthd.Name()] = mthd
}

func (gfc *GoFakeClass) Constants() []ast.Decl {
	return nil
}

func (gfc *GoFakeClass) finalize(gp *GoProgram) {
	// do nothing
}

func (gfc *GoFakeClass) FindMethod(name string,
	args *GoMethodArguments) GoMethod {
	if gfc.methods == nil {
		return nil
	}

	mthd, _ := gfc.methods[name]
	return mthd
}

func (gfc *GoFakeClass) findVariable(typename *grammar.JTypeName) GoVar {
	return nil
}

func (gfc *GoFakeClass) Decls() []ast.Decl {
	return nil
}

func (gfc *GoFakeClass) IsNil() bool {
	return false
}

func (gfc *GoFakeClass) IsReference() bool {
	panic("unimplemented")
}

func (gfc *GoFakeClass) Name() string {
	return gfc.name
}

func (gfc *GoFakeClass) Parent() GoMethodOwner {
	return nil
}

func (gfc *GoFakeClass) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, m := range gfc.methods {
		obj, is_nil := m.RunTransform(xform, prog, cls, gfc)
		if !is_nil {
			var err error
			if gfc.methods[i], err = convertToMethod(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gfc)
}

func (gfc *GoFakeClass) SetPackage(pkg string) {
	gfc.pkg = pkg
}

func (gfc *GoFakeClass) Statics() []ast.Decl {
	return nil
}

func (gfc *GoFakeClass) String() string {
	b := &bytes.Buffer{}
	gfc.WriteString(b, false)
	return b.String()
}

func (gfc *GoFakeClass) Super() GoMethodOwner {
	return nil
}

func (gfc *GoFakeClass) WriteString(out io.Writer, verbose bool) {
	io.WriteString(out, "GoFakeClass[")
	io.WriteString(out, gfc.pkg)
	io.WriteString(out, "|")
	io.WriteString(out, gfc.name)
	io.WriteString(out, "]")
}

type GoFakeMethod struct {
	class GoMethodOwner
	name string
	goname string
	rtntype *TypeData
}

func NewGoFakeMethod(cls GoMethodOwner, name string, rtntype *TypeData) *GoFakeMethod {
	return &GoFakeMethod{class: cls, name: name, goname: name, rtntype: rtntype}
}

func (gfm *GoFakeMethod) Arguments() []GoVar {
	return nil
}

func (gfm *GoFakeMethod) Body() *GoBlock {
	panic("unimplemented")
}

func (gfm *GoFakeMethod) Class() GoMethodOwner {
	return gfm.class
}

func (gfm *GoFakeMethod) Decl() ast.Decl {
	return nil
}

func (gfm *GoFakeMethod) Field() *ast.Field {
	return nil
}

func (gfm *GoFakeMethod) GoName() string {
	return gfm.goname
}

func (gfm *GoFakeMethod) HasArguments(args *GoMethodArguments) bool {
	panic("Unimplemented")
}

func (gfm *GoFakeMethod) IsMethod(mthd GoMethod) bool {
	return false
}

func (gfm *GoFakeMethod) MethodType() methodType {
	return mt_method
}

func (gfm *GoFakeMethod) Name() string {
	return gfm.name
}

func (mthd *GoFakeMethod) NumParameters() int {
	panic("unimplemented")
}

func (gfm *GoFakeMethod) Receiver() GoVar {
	if gfm.class == nil {
		return nil
	}

	return NewFakeVar(gfm.class.Name(), nil, 0)
}

func (gfm *GoFakeMethod) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, gfm)
}

func (gfm *GoFakeMethod) SetGoName(newname string) {
	gfm.goname = newname
}

func (gfm *GoFakeMethod) SetOriginal(gcm *GoClassMethod) {
	panic("Unimplemented")
}

func (gfm *GoFakeMethod) VarType() *TypeData {
	return gfm.rtntype
}

func (gfm *GoFakeMethod) WriteString(out io.Writer) {
	io.WriteString(out, "GoFakeMethod[")
	if gfm.class != nil && !gfm.class.IsNil() {
		gfm.class.WriteString(out, false)
	}
	io.WriteString(out, "|")
	io.WriteString(out, gfm.name)
	io.WriteString(out, "]")
}

type GoForColon struct {
	govar GoVar
	expr GoExpr
	body *GoBlock
}

func (fc *GoForColon) hasVariable(govar GoVar) bool {
	if fc.govar.Equals(govar) {
		return true
	}

	return fc.body.hasVariable(govar)
}

func (fc *GoForColon) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if fc.govar != nil {
		obj, is_nil := fc.govar.RunTransform(xform, prog, cls, fc)
		if !is_nil {
			var err error
			if fc.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if fc.expr != nil {
		obj, is_nil := fc.expr.RunTransform(xform, prog, cls, fc)
		if !is_nil {
			var err error
			if fc.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if fc.body != nil {
		obj, is_nil := fc.body.RunTransform(xform, prog, cls, fc)
		if !is_nil {
			var err error
			if fc.body, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, fc)
}

func (fc *GoForColon) Stmts() []ast.Stmt {
	var expr ast.Expr
	if fc.expr != nil {
		expr = fc.expr.Expr()
	}

	rs := &ast.RangeStmt{Key: ast.NewIdent("_"),
		Value: ast.NewIdent(fc.govar.GoName()),
		Tok: token.DEFINE, X: expr, Body: fc.body.BlockStmt()}
	return []ast.Stmt { rs, }
}

func (fc *GoForColon) String() string {
	var estr string
	if fc.expr != nil {
		estr = fc.expr.String()
	}

	return "GoForColon[" + fc.govar.String() + "|" + estr + "|" +
		fc.body.String()
}

type GoForExpr struct {
	init []GoExpr
	cond GoExpr
	incr []GoExpr
	block *GoBlock
}

func (fe *GoForExpr) hasVariable(govar GoVar) bool {
	if fe.init != nil {
		for _, expr := range fe.init {
			if expr.hasVariable(govar) {
				return true
			}
		}
	}

	if fe.cond != nil && fe.cond.hasVariable(govar) {
		return true
	}

	if fe.incr != nil {
		for _, expr := range fe.incr {
			if expr.hasVariable(govar) {
				return true
			}
		}
	}

	if fe.block != nil && fe.block.hasVariable(govar) {
		return true
	}

	return false
}

func (fe *GoForExpr) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, in := range fe.init {
		obj, is_nil := in.RunTransform(xform, prog, cls, fe)
		if !is_nil {
			var err error
			if fe.init[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if fe.cond != nil {
		obj, is_nil := fe.cond.RunTransform(xform, prog, cls, fe)
		if !is_nil {
			var err error
			if fe.cond, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	for i, in := range fe.incr {
		obj, is_nil := in.RunTransform(xform, prog, cls, fe)
		if !is_nil {
			var err error
			if fe.incr[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if fe.block != nil {
		obj, is_nil := fe.block.RunTransform(xform, prog, cls, fe)
		if !is_nil {
			var err error
			if fe.block, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, fe)
}

func (fe *GoForExpr) Stmts() []ast.Stmt {
	var init ast.Stmt
	if fe.init != nil && len(fe.init) > 0 {
		init = &ast.ExprStmt{X: fe.init[0].Expr()}

		if len(fe.init) > 1 {
			log.Printf("//ERR// ignoring extra forexpr init (%d stmts)\n",
				len(fe.init))
		}
	}

	var cond ast.Expr
	if fe.cond != nil {
		cond = fe.cond.Expr()
	}

	var post ast.Stmt
	if fe.incr != nil && len(fe.incr) > 0 {
		incr := fe.incr[0].Expr()
		if incr != nil {
			post = &ast.ExprStmt{X: incr}
		}

		if len(fe.incr) > 1 {
			log.Printf("//ERR// ignoring extra forexpr incr (%d stmts)\n",
				len(fe.incr))
		}
	}

	var block *ast.BlockStmt
	if fe.block != nil {
		block = fe.block.BlockStmt()
	}

	return []ast.Stmt { &ast.ForStmt{Init: init, Cond: cond, Post: post,
		Body: block}, }
}

func (fe *GoForExpr) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoArrayAlloc[")
	if fe.init != nil && len(fe.init) > 0 {
		for i, v := range fe.init {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(v.String())
		}
	}
	b.WriteString("|")
	if fe.cond != nil {
		b.WriteString(fe.cond.String())
	}
	b.WriteString("|")
	if fe.incr != nil && len(fe.incr) > 0 {
		for i, v := range fe.incr {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(v.String())
		}
	}
	b.WriteString("|")
	b.WriteString(fe.block.String())
	b.WriteString("]")
	return b.String()
}

type GoForVar struct {
	govar GoVar
	init GoExpr
	cond GoExpr
	incr []GoStatement
	block *GoBlock
}

func (gfv *GoForVar) hasVariable(govar GoVar) bool {
	if gfv.govar.Equals(govar) {
		return true
	}

	if gfv.init != nil && gfv.init.hasVariable(govar) {
		return true
	}

	if gfv.cond != nil && gfv.cond.hasVariable(govar) {
		return true
	}

	if gfv.incr != nil {
		for _, expr := range gfv.incr {
			if expr.hasVariable(govar) {
				return true
			}
		}
	}

	if gfv.block != nil {
		return gfv.block.hasVariable(govar)
	}

	return false
}

func (gfv *GoForVar) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gfv.govar != nil {
		obj, is_nil := gfv.govar.RunTransform(xform, prog, cls, gfv)
		if !is_nil {
			var err error
			if gfv.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if gfv.init != nil {
		obj, is_nil := gfv.init.RunTransform(xform, prog, cls, gfv)
		if !is_nil {
			var err error
			if gfv.init, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if gfv.cond != nil {
		obj, is_nil := gfv.cond.RunTransform(xform, prog, cls, gfv)
		if !is_nil {
			var err error
			if gfv.cond, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	for i, in := range gfv.incr {
		obj, is_nil := in.RunTransform(xform, prog, cls, gfv)
		if !is_nil {
			var err error
			if gfv.incr[i], err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	if gfv.block != nil {
		obj, is_nil := gfv.block.RunTransform(xform, prog, cls, gfv)
		if !is_nil {
			var err error
			if gfv.block, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gfv)
}

func (gfv *GoForVar) Stmts() []ast.Stmt {
	var init ast.Stmt
	if gfv.init == nil {
		names := make([]*ast.Ident, 1)
		names[0] = gfv.govar.Ident()

		valspec := &ast.ValueSpec{Names: names, Type: gfv.govar.Type()}

		specs := make([]ast.Spec, 1)
		specs[0] = valspec
		preinit := &ast.GenDecl{Tok: token.VAR, Specs: specs}

		log.Printf("//ERR// need to preinit forvar %T\n", preinit)
	} else {
		lhs := make([]ast.Expr, 1)
		lhs[0] = ast.NewIdent(gfv.govar.Name())

		rhs := make([]ast.Expr, 1)
		rhs[0] = gfv.init.Expr()

		init = &ast.AssignStmt{Lhs: lhs, Tok: token.DEFINE, Rhs: rhs}
	}

	var cond ast.Expr
	if gfv.cond != nil {
		cond = gfv.cond.Expr()
	}

	var post ast.Stmt
	if gfv.incr != nil && len(gfv.incr) > 0 {
		if len(gfv.incr) == 1 {
			var is_nil bool
			post, is_nil = singleStatement("ForVar incr", gfv.incr[0].Stmts())
			if is_nil {
				post = nil
			}
		} else {
			log.Printf("//ERR// ignoring forvar incr (%d stmts)\n",
				len(gfv.incr))
		}
	}

	var block *ast.BlockStmt
	if gfv.block != nil {
		stmts := gfv.block.Stmts()
		if stmts == nil || len(stmts) == 0 {
			block = nil
		} else if len(stmts) == 1 {
			if blk, ok := stmts[0].(*ast.BlockStmt); ok {
				block = blk
			}
		}
		if block == nil && len(stmts) > 0 {
			block = &ast.BlockStmt{List: stmts}
		}
	}

	if block == nil {
		log.Printf("//ERR// adding empty forvar body\n")
		list := make([]ast.Stmt, 0)
		block = &ast.BlockStmt{List: list}
	}

	return []ast.Stmt { &ast.ForStmt{Init: init, Cond: cond, Post: post,
		Body: block}, }
}

func (gfv *GoForVar) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoForVar[")
	if gfv.govar != nil {
		b.WriteString(gfv.govar.String())
	}
	b.WriteString("|")
	if gfv.init != nil {
		b.WriteString(gfv.init.String())
	}
	b.WriteString("|")
	if gfv.cond != nil {
		b.WriteString(gfv.cond.String())
	}
	b.WriteString("|")
	if gfv.incr != nil && len(gfv.incr) > 0 {
		for i, v := range gfv.incr {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(v.String())
		}
	}
	b.WriteString("|")
	if gfv.block != nil {
		b.WriteString(gfv.block.String())
	}
	b.WriteString("]")
	return b.String()
}

type GoIfaceMethod struct {
	name string
	goname string
	param_list []GoVar
	result_type *TypeData
}

func NewGoInterfaceMethod(gp *GoProgram, iface_name string,
	imth *grammar.JInterfaceMethodDecl) *GoIfaceMethod {
	gm := &GoIfaceMethod{}

	gm.name = imth.Name
	gm.goname = strings.ToUpper(imth.Name[:1]) + imth.Name[1:]

	var gs *GoState

	if imth.FormalParams != nil && len(imth.FormalParams) > 0 {
		gm.param_list = make([]GoVar, len(imth.FormalParams))
		for i, fp := range imth.FormalParams {
			if fp.TypeSpec != nil {
				if fp.Dims != 0 {
					if gp.verbose {
						log.Printf("//ERR// Ignoring %s dims=%d for %s.%s\n",
							fp.Name, fp.Dims, iface_name, imth.Name)
					} else {
						log.Printf("//ERR// Ignoring non-zero interface dims\n")
					}
				} else if fp.DotDotDot {
					if gp.verbose {
						log.Printf("//ERR// Ignoring %s DotDotDot=true for %s.%s\n",
							fp.Name, iface_name, imth.Name)
					} else {
						log.Printf("//ERR// Ignoring interface DotDotDot\n")
					}
				}

				if gs == nil {
					gs = &GoState{program: gp}
				}

				govar := gs.addVariable(fp.Name, fp.Modifiers, fp.Dims,
					fp.TypeSpec, false)

				gm.param_list[i] = govar
			}
		}
	}

	if imth.TypeSpec != nil {
		gm.result_type = gp.createTypeData(imth.TypeSpec.Name,
			imth.TypeSpec.TypeArgs, imth.TypeSpec.Dims)
	}

	return gm
}

func (gm *GoIfaceMethod) Arguments() []GoVar {
	return gm.param_list
}

func (gm *GoIfaceMethod) Body() *GoBlock {
	return nil
}

func (gm *GoIfaceMethod) Class() GoMethodOwner {
	return nilMethodOwner
}

func (gm *GoIfaceMethod) Decl() ast.Decl {
	return nil
}

func (gm *GoIfaceMethod) Field() *ast.Field {
	itype := &ast.FuncType{Params: gm.params(), Results: gm.results()}

	iname := make([]*ast.Ident, 1)
	iname[0] = ast.NewIdent(gm.goname)

	return &ast.Field{Names: iname, Type: itype}
}

func (gm *GoIfaceMethod) GoName() string {
	return gm.goname
}

func (gm *GoIfaceMethod) HasArguments(args *GoMethodArguments) bool {
	panic("Unimplemented")
}

func (gm *GoIfaceMethod) IsMethod(mthd GoMethod) bool {
	return false
}

func (gm *GoIfaceMethod) MethodType() methodType {
	return mt_interface
}

func (gm *GoIfaceMethod) Name() string {
	return gm.name
}

func (gm *GoIfaceMethod) NumParameters() int {
	return len(gm.param_list)
}

func (gm *GoIfaceMethod) params() *ast.FieldList {
	var flist []*ast.Field
	if gm.param_list != nil {
		flist = make([]*ast.Field, len(gm.param_list))

		for i, fp := range gm.param_list {
			flist[i] = makeField(fp.Name(), fp.Type())
		}
	}

	return &ast.FieldList{List: flist}
}

func (gm *GoIfaceMethod) Receiver() GoVar {
	return nil
}

func (gm *GoIfaceMethod) results() *ast.FieldList {
	if gm.result_type != nil {
		if typename, is_nil := gm.result_type.TypeName(); !is_nil {
			rlist := make([]*ast.Field, 1)
			rlist[0] = &ast.Field{Names: make([]*ast.Ident, 0), Type: typename}

			return &ast.FieldList{List: rlist}
		}
	}

	return nil
}

func (gm *GoIfaceMethod) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, p := range gm.param_list {
		obj, is_nil := p.RunTransform(xform, prog, cls, gm)
		if !is_nil {
			var err error
			if gm.param_list[i], err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gm)
}

func (gm *GoIfaceMethod) SetGoName(newname string) {
	gm.goname = newname
}

func (gm *GoIfaceMethod) SetOriginal(gcm *GoClassMethod) {
	panic("Unimplemented")
}

func (gm *GoIfaceMethod) VarType() *TypeData {
	return gm.result_type
}

func (gm *GoIfaceMethod) WriteString(out io.Writer) {
	io.WriteString(out, "GoIfaceMethod[")
	io.WriteString(out, gm.name)
	io.WriteString(out, "|")
	io.WriteString(out, gm.goname)
	io.WriteString(out, "|")
	for i, prm := range gm.param_list {
		if i > 0 {
			io.WriteString(out, ",")
		}
		io.WriteString(out, prm.String())
	}
	io.WriteString(out, "|")
	io.WriteString(out, gm.result_type.String())
	io.WriteString(out, "]")
}

type GoIfElse struct {
    cond GoExpr
	ifblk GoStatement
	elseblk GoStatement
}

func (gie *GoIfElse) hasVariable(govar GoVar) bool {
	if gie.cond != nil && gie.cond.hasVariable(govar) {
		return true
	}

	if gie.ifblk != nil && gie.ifblk.hasVariable(govar) {
		return true
	}

	if gie.elseblk != nil && gie.elseblk.hasVariable(govar) {
		return true
	}

	return false
}

func (gie *GoIfElse) IfStmt() *ast.IfStmt {
	init := gie.cond.Init()
	cond := gie.cond.Expr()

	ifblk := gie.ifblk.Stmts()
	if ifblk == nil {
		panic("If-block cannot return nil")
	}

	var body *ast.BlockStmt
	if len(ifblk) == 1 {
		if b, ok := ifblk[0].(*ast.BlockStmt); ok {
			body = b
		}
	}
	if body == nil {
		body = &ast.BlockStmt{List: ifblk}
	}

	ifstmt := &ast.IfStmt{Init: init, Cond: cond, Body: body}
	if gie.elseblk != nil {
		stmts := gie.elseblk.Stmts()
		if stmts == nil {
			panic("Else-block cannot return nil")
		} else if len(stmts) > 1 {
			panic("Else-block cannot return multiple statements")
		} else if len(stmts) == 1 {
			ifstmt.Else = stmts[0]
		}
	}

	return ifstmt
}

func (gie *GoIfElse) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gie.cond != nil {
		obj, is_nil := gie.cond.RunTransform(xform, prog, cls, gie)
		if !is_nil {
			var err error
			if gie.cond, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if gie.ifblk != nil {
		obj, is_nil := gie.ifblk.RunTransform(xform, prog, cls, gie)
		if !is_nil {
			var err error
			if gie.ifblk, err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	if gie.elseblk != nil {
		obj, is_nil := gie.elseblk.RunTransform(xform, prog, cls, gie)
		if !is_nil {
			var err error
			if gie.elseblk, err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gie)
}

func (gie *GoIfElse) Stmts() []ast.Stmt {
	return []ast.Stmt { gie.IfStmt(), }
}

func (gie *GoIfElse) String() string {
	var estr string
	if gie.elseblk != nil {
		estr = gie.elseblk.String()
	}

	return "GoIfElse[" + gie.cond.String() + "|" + gie.ifblk.String() + "|" +
		estr + "]"
}

type GoImportPackage struct {
	name string
	classes map[string]*GoImportClass
}

func NewGoImportPackage(pkgname string) *GoImportPackage {
	return &GoImportPackage{name: pkgname,
		classes: make(map[string]*GoImportClass)}
}

func (gip *GoImportPackage) addClass(clsname string) (cls *GoImportClass) {
	if c, ok := gip.classes[clsname]; ok {
		return c
	}

	cls = &GoImportClass{pkg: gip, name: clsname}
	gip.classes[clsname] = cls
	return
}

func (gip *GoImportPackage) ImportSpec() *ast.ImportSpec {
	return &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING,
		Value: fmt.Sprintf("\"%s\"", gip.name)}}
}

type GoImportClass struct {
	pkg *GoImportPackage
	name string
}

func (gic *GoImportClass) FullName() string {
	return gic.pkg.name + "." + gic.name
}

type GoInstanceOf struct {
	expr GoExpr
	vartype GoVar
}

func (gio *GoInstanceOf) Expr() ast.Expr {
	return ast.NewIdent("ok")
}

func (gio *GoInstanceOf) hasVariable(govar GoVar) bool {
	if gio.vartype != nil && gio.vartype.Equals(govar) {
		return true
	}

	if gio.expr != nil && gio.expr.hasVariable(govar) {
		return true
	}

	return false
}

func (gio *GoInstanceOf) Init() ast.Stmt {
	lhs := []ast.Expr { ast.NewIdent("_"), ast.NewIdent("ok") }
	asrt := &ast.TypeAssertExpr{X: gio.expr.Expr(), Type: gio.vartype.Ident() }

	return &ast.AssignStmt{Lhs: lhs, Tok: token.DEFINE,
		Rhs: []ast.Expr { asrt, }}
}

func (gio *GoInstanceOf) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gio.expr != nil {
		obj, is_nil := gio.expr.RunTransform(xform, prog, cls, gio)
		if !is_nil {
			var err error
			if gio.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if gio.vartype != nil {
		obj, is_nil := gio.vartype.RunTransform(xform, prog, cls, gio)
		if !is_nil {
			var err error
			if gio.vartype, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gio)
}

func (gio *GoInstanceOf) String() string {
	var estr string
	if gio.expr != nil {
		estr = gio.expr.String()
	}

	var vstr string
	if gio.vartype != nil {
		vstr = gio.vartype.String()
	}

	return "GoInstanceOf[" + estr + "|" + vstr + "]"
}

func (gio *GoInstanceOf) VarType() *TypeData {
	return boolType
}

type GoInterface interface {
	Constants() []ast.Decl
	Decl() ast.Decl
	finalize(*GoProgram)
	findVariable(*grammar.JTypeName) GoVar
	IsInterface() bool
	Matches(*grammar.JTypeName) bool
	Name() string
	String() string
	WriteString(io.Writer, bool)
}

// InterfaceSlice is a GoInterface wrapper for sort.Sort()
type InterfaceSlice []GoInterface
func (p InterfaceSlice) Len() int { return len(p) }
func (p InterfaceSlice) Less(i, j int) bool { return p[i].Name() < p[j].Name() }
func (p InterfaceSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type GoInterfaceDefinition struct {
	name string

	methods *interfaceMethodMap
	constants []*GoConstant
}

func NewGoInterfaceDefinition(name string) *GoInterfaceDefinition {
	return &GoInterfaceDefinition{name: name, methods: NewInterfaceMethodMap()}
}

func (gi *GoInterfaceDefinition) AddConstant(con *GoConstant) {
	if gi.constants == nil {
		gi.constants = make([]*GoConstant, 0)
	}
	gi.constants = append(gi.constants, con)
}

func (gi *GoInterfaceDefinition) AddMethod(newmthd GoMethod) {
	gi.methods.AddMethod(newmthd, gi.methods)
}

func (gi *GoInterfaceDefinition) Constants() []ast.Decl {
	if gi.constants == nil || len(gi.constants) == 0 {
		return nil
	}

	decllist := make([]ast.Decl, 0)
	for _, c := range gi.constants {
		names := []*ast.Ident{ ast.NewIdent(c.name), }
		var vals []ast.Expr
		if c.init != nil {
			vals = []ast.Expr{ c.init.Expr() }
		}

		spec := &ast.ValueSpec{Names: names, Type: c.typedata.Expr(),
			Values: vals }
		decllist = append(decllist, &ast.GenDecl{Tok: token.CONST,
			Specs: []ast.Spec{ spec, }})
	}

	return decllist
}

func (gi *GoInterfaceDefinition) Decl() ast.Decl {
	list := make([]*ast.Field, 0)
	for _, key := range gi.methods.SortedKeys() {
		for _, m := range gi.methods.MethodList(key) {
			fld := m.Field()
			if fld != nil {
				list = append(list, fld)
			}
		}
	}

	sttype := &ast.InterfaceType{Methods: &ast.FieldList{List: list}}

	specs := make([]ast.Spec, 1)
	specs[0] = &ast.TypeSpec{Name: ast.NewIdent(gi.name),
		Type: sttype}

	return &ast.GenDecl{Tok: token.TYPE, Specs: specs}
}

func (gi *GoInterfaceDefinition) finalize(gp *GoProgram) {
	// renumber duplicate methods
	gi.renumberDuplicateMethods(gp)
}

func (gi *GoInterfaceDefinition) findVariable(name *grammar.JTypeName) GoVar {
	for _, c := range gi.constants {
		if c.name == name.String() {
			return c
		}
	}

	return nil
}

func (gi *GoInterfaceDefinition) FindMethod(name string,
	args *GoMethodArguments) GoMethod {
	return gi.methods.FindMethod(name, args)
}

func (gi *GoInterfaceDefinition) IsInterface() bool {
	return false
}

func (gi *GoInterfaceDefinition) IsNil() bool {
	return false
}

func (gi *GoInterfaceDefinition) Matches(name *grammar.JTypeName) bool {
	return name.String() == gi.name || name.LastType() == gi.name
}

func (gi *GoInterfaceDefinition) Name() string {
    return gi.name
}

func (gi *GoInterfaceDefinition) renumberDuplicateMethods(gp *GoProgram) {
	gi.methods.renumberDuplicateMethods(gp)
}

func (gi *GoInterfaceDefinition) Statics() []ast.Decl {
	return nil
}

func (gi *GoInterfaceDefinition) String() string {
	b := &bytes.Buffer{}
	gi.WriteString(b, false)
	return b.String()
}

func (gi *GoInterfaceDefinition) Super() GoMethodOwner {
	panic("Unimplemented")
}

func (gi *GoInterfaceDefinition) WriteString(out io.Writer, verbose bool) {
	io.WriteString(out, "GoInterfaceDefinition[")
	io.WriteString(out, gi.name)

	io.WriteString(out, "|")
	for i, c := range gi.constants {
		if i > 0 {
			io.WriteString(out, ",")
		}
		c.WriteString(out)
	}

	io.WriteString(out, "|")
	if !verbose {
		io.WriteString(out, fmt.Sprintf("%d methods", gi.methods.Length()))
	} else {
		gi.methods.WriteString(out, verbose)
	}

	io.WriteString(out, "]")
}

type GoInterfaceReference struct {
	name *grammar.JTypeName
	gp *GoProgram
	realiface *GoInterfaceDefinition
}

func (ref *GoInterfaceReference) Constants() []ast.Decl {
	if ref.realiface != nil {
		return ref.realiface.Constants()
	}

	return nil
}

func (ref *GoInterfaceReference) Decl() ast.Decl {
	if ref.realiface != nil {
		return ref.realiface.Decl()
	}

	return nil
}

func (ref *GoInterfaceReference) finalize(gp *GoProgram) {
	if ref.realiface == nil {
		iface := ref.gp.findInterface(ref.name)
		if realiface, ok := iface.(*GoInterfaceDefinition); ok {
			ref.realiface = realiface
		}
	}

	if ref.realiface != nil {
		ref.realiface.finalize(gp)
	}
}

func (ref *GoInterfaceReference) findVariable(name *grammar.JTypeName) GoVar {
	if ref.realiface != nil {
		return ref.realiface.findVariable(name)
	}

	return nil
}

func (ref *GoInterfaceReference) IsInterface() bool {
	return true
}

func (ref *GoInterfaceReference) Matches(name *grammar.JTypeName) bool {
	return name.LastType() == ref.name.LastType() ||
		name.String() == ref.name.String()
}

func (ref *GoInterfaceReference) Name() string {
	return ref.name.String()
}

func (ref *GoInterfaceReference) String() string {
	if ref.realiface != nil {
		return ref.realiface.String()
	}

	b := &bytes.Buffer{}
	ref.WriteString(b, false)
	return b.String()
}

func (ref *GoInterfaceReference) WriteString(out io.Writer, verbose bool) {
	io.WriteString(out, "GoInterfaceReference[")
	io.WriteString(out, ref.name.String())
	io.WriteString(out, "]")
}

type GoJumpToLabel struct {
	is_continue bool
	label string
}

func NewGoJumpToLabel(label string, is_continue bool) *GoJumpToLabel {
	return &GoJumpToLabel{is_continue: is_continue, label: label}
}

func (j2l *GoJumpToLabel) hasVariable(govar GoVar) bool {
	return false
}

func (j2l *GoJumpToLabel) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, j2l)
}

func (j2l *GoJumpToLabel) Stmts() []ast.Stmt {
	var tok token.Token
	if j2l.is_continue {
		tok = token.CONTINUE
	} else {
		tok = token.BREAK
	}

	var lbl *ast.Ident
	if j2l.label != "" {
		lbl = ast.NewIdent(j2l.label)
	}

	return []ast.Stmt { &ast.BranchStmt{Tok: tok, Label: lbl}, }
}

func (j2l *GoJumpToLabel) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoJumpToLabel[")
	if j2l.is_continue {
		b.WriteString("continue")
	} else {
		b.WriteString("break")
	}
	b.WriteString("|")
	b.WriteString(j2l.label)
	b.WriteString("]")
	return b.String()
}

type GoKeyword struct {
	token int
	name string
}

func NewGoKeyword(token int, name string) *GoKeyword {
	if name == "" {
		grammar.ReportError("Keyword name cannot be empty")
	}

	return &GoKeyword{token: token, name: name}
}

func (key *GoKeyword) Expr() ast.Expr {
	return key.Ident()
}

func (key *GoKeyword) hasVariable(govar GoVar) bool {
	return false
}

func (key *GoKeyword) Ident() *ast.Ident {
	if key.name == "null" {
		return ast.NewIdent("nil")
	} else if key.name == "true" || key.name == "false" {
		return ast.NewIdent(key.name)
	} else if key.name == "this" {
		return ast.NewIdent("this")
	}
	log.Printf("//ERR// Not converting keyword %s\n", key.name)
	return ast.NewIdent(fmt.Sprintf("<<unimp_key_%s>>", key.name))
}

func (key *GoKeyword) Init() ast.Stmt {
	return nil
}

func (key *GoKeyword) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, key)
}

func (key *GoKeyword) String() string {
	return "GoKeyword[" + grammar.JulyTokname(key.token) + "|" + key.name + "]"
}

func (key *GoKeyword) VarType() *TypeData {
	if key.name == "null" {
		return voidType
	} else if key.name == "true" || key.name == "false" {
		return boolType
	}
	log.Printf("//ERR// Not returning type for keyword %s\n", key.name)
	return nil
}

type GoLabeledStmt struct {
	label string
	stmt GoStatement
}

func (gl *GoLabeledStmt) hasVariable(govar GoVar) bool {
	return gl.stmt.hasVariable(govar)
}

func (gl *GoLabeledStmt) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gl.stmt != nil {
		obj, is_nil := gl.stmt.RunTransform(xform, prog, cls, gl)
		if !is_nil {
			var err error
			if gl.stmt, err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gl)
}

func (gl *GoLabeledStmt) Stmts() []ast.Stmt {
	stmt, is_nil := singleStatement("labeled statement", gl.stmt.Stmts())
	if is_nil {
		return nil
	}
	return []ast.Stmt { &ast.LabeledStmt{Label: ast.NewIdent(gl.label),
		Stmt: stmt}, }
}

func (gl *GoLabeledStmt) Init() ast.Stmt {
	return nil
}

func (gl *GoLabeledStmt) String() string {
	return "GoLabeledStmt[" + gl.label + "|" + gl.stmt.String() + "]"
}

type GoLiteral struct {
	text string
}

func NewGoLiteral(text string) *GoLiteral {
	if text == "" {
		grammar.ReportError("Literal text cannot be empty")
	}

	return &GoLiteral{text: text}
}

func (gl *GoLiteral) Expr() ast.Expr {
	var kind token.Token
	var value string
	if gl.text[0] == '"' {
		kind = token.STRING
		value = gl.text
	} else if gl.text[0] == '\'' {
		kind = token.CHAR
		value = gl.text
	} else if strings.Contains(gl.text, ".") {
		kind = token.FLOAT
		value = gl.text
	} else {
		kind = token.INT
		if !strings.HasSuffix(gl.text, "L") {
			value = gl.text
		} else {
			value = gl.text[0:len(gl.text)-1]
		}
	}

	return &ast.BasicLit{Kind: kind, Value: value}
}

func (gl *GoLiteral) hasVariable(govar GoVar) bool {
	return false
}

func (gl *GoLiteral) Init() ast.Stmt {
	return nil
}

func (gl *GoLiteral) isString() bool {
	return len(gl.text) > 1 && gl.text[0] == '"'
}

func (gl *GoLiteral) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, gl)
}

func (gl *GoLiteral) String() string {
	return "GoLiteral[" + gl.text + "]"
}

func (gl *GoLiteral) VarType() *TypeData {
	if gl.text[0] == '"' {
		return stringType
	} else if gl.text[0] == '\'' {
		return charType
	} else if strings.Contains(gl.text, ".") {
		return doubleType
	}

	return intType
}

type GoLocalVarNoInit struct {
	govar GoVar
}

func NewGoLocalVarNoInit(govar GoVar) *GoLocalVarNoInit {
	return &GoLocalVarNoInit{govar: govar};
}

func (glv *GoLocalVarNoInit) hasVariable(govar GoVar) bool {
	return glv.govar.Equals(govar)
}

func (glv *GoLocalVarNoInit) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if glv.govar != nil {
		obj, is_nil := glv.govar.RunTransform(xform, prog, cls, glv)
		if !is_nil {
			var err error
			if glv.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, glv)
}

func (glv *GoLocalVarNoInit) Stmts() []ast.Stmt {
	names := make([]*ast.Ident, 1)
	names[0] = ast.NewIdent(glv.govar.Name())

	specs := make([]ast.Spec, 1)
	specs[0] = &ast.ValueSpec{Names: names, Type: glv.govar.Type()}

	return []ast.Stmt { &ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR,
		Specs: specs}}, }
}

func (glv *GoLocalVarNoInit) String() string {
	return "GoLocalVarNoInit[" + glv.govar.String() + "]"
}

type GoLocalVarInit struct {
	govar GoVar
	init GoExpr
}

func NewGoLocalVarInit(govar GoVar, init GoExpr) *GoLocalVarInit {
	if init == nil {
		panic("Local variable init expression cannot be nil")
	}

	return &GoLocalVarInit{govar: govar, init: init};
}

func (glv *GoLocalVarInit) hasVariable(govar GoVar) bool {
	if glv.govar.Equals(govar) {
		return true
	}

	return glv.init.hasVariable(govar)
}

func (glv *GoLocalVarInit) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if glv.govar != nil {
		obj, is_nil := glv.govar.RunTransform(xform, prog, cls, glv)
		if !is_nil {
			var err error
			if glv.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if glv.init != nil {
		obj, is_nil := glv.init.RunTransform(xform, prog, cls, glv)
		if !is_nil {
			var err error
			if glv.init, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, glv)
}

func (glv *GoLocalVarInit) Stmts() []ast.Stmt {
	rhs := make([]ast.Expr, 1)
	rhs[0] = glv.init.Expr()

	lhs := make([]ast.Expr, 1)
	lhs[0] = ast.NewIdent(glv.govar.Name())

	tok := token.DEFINE

	return []ast.Stmt { &ast.AssignStmt{Lhs: lhs, Tok: tok, Rhs: rhs}, }
}

func (glv *GoLocalVarInit) String() string {
	return "GoLocalVarInit[" + glv.govar.String() + "|" +
		glv.init.String() + "]"
}

type GoLocalVarCast struct {
	govar GoVar
	cast GoExpr
}

func NewGoLocalVarCast(govar GoVar, cast GoExpr) *GoLocalVarCast {
	if cast == nil {
		panic("Local variable cast expression cannot be nil")
	}

	return &GoLocalVarCast{govar: govar, cast: cast}
}

func (glv *GoLocalVarCast) hasVariable(govar GoVar) bool {
	if glv.govar.Equals(govar) {
		return true
	}

	return glv.cast.hasVariable(govar)
}

func (glv *GoLocalVarCast) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if glv.govar != nil {
		obj, is_nil := glv.govar.RunTransform(xform, prog, cls, glv)
		if !is_nil {
			var err error
			if glv.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if glv.cast != nil {
		obj, is_nil := glv.cast.RunTransform(xform, prog, cls, glv)
		if !is_nil {
			var err error
			if glv.cast, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, glv)
}

func (glv *GoLocalVarCast) Stmts() []ast.Stmt {
	blklist := make([]ast.Stmt, 2)

	rhs := make([]ast.Expr, 1)
	rhs[0] = glv.cast.Expr()

	ok_ident := ast.NewIdent("ok")

	lhs := make([]ast.Expr, 2)
	lhs[0] = ast.NewIdent(glv.govar.Name())
	lhs[1] = ok_ident

	tok := token.DEFINE

	blklist[0] = &ast.AssignStmt{Lhs: lhs, Tok: tok, Rhs: rhs}

	cond := &ast.UnaryExpr{Op: token.NOT, X: ok_ident}

	args := make([]ast.Expr, 1)
	args[0] = &ast.BasicLit{Kind: token.STRING,
		//Value: fmt.Sprintf("\"Cannot cast %v to %v\"", glv.cast.target,
		//	glv.cast.vartype)}
		Value: fmt.Sprintf("\"XXX Cast fail for %T\"", glv.cast)}

	panic_expr := &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: args}

	list := make([]ast.Stmt, 1)
	list[0] = &ast.ExprStmt{X: panic_expr}

	blklist[1] = &ast.IfStmt{Cond: cond, Body: &ast.BlockStmt{List: list}}

	return blklist
}

func (glc *GoLocalVarCast) String() string {
	return "GoLocalVarCast[" + glc.govar.String() + "|" +
		glc.cast.String() + "]"
}

type methodType int

const (
	mt_test methodType = iota
	mt_constructor
	mt_static
	mt_method
	mt_interface
	mt_main
)

func (mt methodType) String() string {
	switch (mt) {
	case mt_test: return "mt_test"
	case mt_constructor: return "mt_constructor"
	case mt_static: return "mt_static"
	case mt_method: return "mt_method"
	case mt_main: return "mt_main"
	}

	return "mt_??unknown??"
}

type GoMethod interface {
	GoObject
	Arguments() []GoVar
	Body() *GoBlock
	Class() GoMethodOwner
	Decl() ast.Decl
	Field() *ast.Field
	GoName() string
	HasArguments(args *GoMethodArguments) bool
	IsMethod(mthd GoMethod) bool
	MethodType() methodType
	Name() string
	NumParameters() int
	Receiver() GoVar
	SetGoName(name string)
	SetOriginal(gcm *GoClassMethod)
	VarType() *TypeData
	WriteString(out io.Writer)
}

type GoMethodAccess struct {
	obj GoExpr
	method GoMethod
	args *GoMethodArguments
}

func (ma *GoMethodAccess) Expr() ast.Expr {
	var fun ast.Expr
	if ma.obj != nil {
		fun = ma.obj.Expr()
	} else {
		if ma.method.Receiver() != nil {
			fun = &ast.SelectorExpr{X: ma.method.Receiver().Ident(),
				Sel: ast.NewIdent(ma.method.GoName())}
		} else if ma.method.Class() != nil &&
			!ma.method.Class().IsNil() {
			fun = &ast.SelectorExpr{X: ast.NewIdent(ma.method.Class().Name()),
				Sel: ast.NewIdent(ma.method.GoName())}
		} else {
			fun = ast.NewIdent(ma.method.GoName())
		}
	}

	return &ast.CallExpr{Fun: fun, Args: ma.args.ExprList()}
}

func (ma *GoMethodAccess) hasVariable(govar GoVar) bool {
	if ma.obj != nil && ma.obj.hasVariable(govar) {
		return true
	}

	if ma.args != nil && ma.args.hasVariable(govar) {
		return true
	}

	return false
}

func (ma *GoMethodAccess) Init() ast.Stmt {
	return nil
}

func (ma *GoMethodAccess) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if ma.obj != nil {
		obj, is_nil := ma.obj.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.obj, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if ma.method != nil {
		obj, is_nil := ma.method.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.method, err = convertToMethod(obj); err != nil {
				panic(err)
			}
		}
	}

	if ma.args != nil {
		obj, is_nil := ma.args.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.args, err = convertToMethodArgs(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, ma)
}

func (ma *GoMethodAccess) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoMethodAccess[")
	if ma.obj != nil {
		b.WriteString(ma.obj.String())
	}
	b.WriteString("|")
	if ma.method != nil {
		ma.method.WriteString(b)
	}
	b.WriteString("|")
	ma.args.WriteString(b)
	b.WriteString("]<GMA>")
	return b.String()
}

func (ma *GoMethodAccess) VarType() *TypeData {
	if ma.method != nil {
		return ma.method.VarType()
	}

	panic("GoMethodAccess.VarType() unimplemented")
}

type GoMethodAccessExpr struct {
	expr GoExpr
	method GoMethod
	args *GoMethodArguments
}

func (ma *GoMethodAccessExpr) Expr() ast.Expr {
	var fun ast.Expr
	if ma.method == nil {
		fun = ma.expr.Expr()
	} else {
		fun = &ast.SelectorExpr{X: ma.expr.Expr(),
			Sel: ast.NewIdent(ma.method.Name())}
	}

	return &ast.CallExpr{Fun: fun, Args: ma.args.ExprList()}
}

func (ma *GoMethodAccessExpr) hasVariable(govar GoVar) bool {
	if ma.expr != nil && ma.expr.hasVariable(govar) {
		return true
	}

	if ma.args != nil && ma.args.hasVariable(govar) {
		return true
	}

	return false
}

func (ma *GoMethodAccessExpr) Init() ast.Stmt {
	return nil
}

func (ma *GoMethodAccessExpr) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if ma.expr != nil {
		obj, is_nil := ma.expr.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if ma.method != nil {
		obj, is_nil := ma.method.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.method, err = convertToMethod(obj); err != nil {
				panic(err)
			}
		}
	}

	if ma.args != nil {
		obj, is_nil := ma.args.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.args, err = convertToMethodArgs(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, ma)
}

func (ma *GoMethodAccessExpr) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoMethodAccessExpr[")
	b.WriteString(ma.expr.String())
	b.WriteString("|")
	if ma.method != nil {
		var name string
		if ma.method.Receiver() != nil {
			name = ma.method.Receiver().Name()
		}
		fmt.Fprintf(b, "GoMethod[%v|%v#%d]", name, ma.method.Name(),
			ma.method.NumParameters())
	}
	b.WriteString("|")
	ma.args.WriteString(b)
	b.WriteString("]")
	return b.String()
}

func (ma *GoMethodAccessExpr) VarType() *TypeData {
	if ma.method != nil {
		return ma.method.VarType()
	}

	panic("GoMethodAccessExpr.VarType() unimplemented")
}

type GoMethodAccessKeyword struct {
	is_super bool
	args *GoMethodArguments
}

func (ma *GoMethodAccessKeyword) Expr() ast.Expr {
	var str string
	if ma.is_super {
		str = "super"
	} else {
		str = "this"
	}

	panic(fmt.Sprintf("GoMethodAccessKeyword.Expr() unimplemented for %v(%v)",
		str, ma.args))
}

func (ma *GoMethodAccessKeyword) hasVariable(govar GoVar) bool {
	if ma.args != nil && ma.args.hasVariable(govar) {
		return true
	}

	return false
}

func (ma *GoMethodAccessKeyword) Init() ast.Stmt {
	return nil
}

func (ma *GoMethodAccessKeyword) RunTransform(xform TransformFunc,
	prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if ma.args != nil {
		obj, is_nil := ma.args.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.args, err = convertToMethodArgs(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, ma)
}

func (ma *GoMethodAccessKeyword) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoMethodAccessKeyword[")
	if ma.is_super {
		b.WriteString("super")
	} else {
		b.WriteString("this")
	}
	b.WriteString("|")
	ma.args.WriteString(b)
	b.WriteString("]")
	return b.String()
}

func (ma *GoMethodAccessKeyword) VarType() *TypeData {
	panic("GoMethodAccessKeyword.VarType() unimplemented")
}

type GoMethodAccessVar struct {
	govar GoVar
	method GoMethod
	args *GoMethodArguments
}

func (ma *GoMethodAccessVar) Expr() ast.Expr {
	fun := &ast.SelectorExpr{X: ma.govar.Expr(),
		Sel: ast.NewIdent(ma.method.Name())}

	return &ast.CallExpr{Fun: fun, Args: ma.args.ExprList()}
}

func (ma *GoMethodAccessVar) hasVariable(govar GoVar) bool {
	if ma.govar != nil && ma.govar.Equals(govar) {
		return true
	}

	if ma.args != nil && ma.args.hasVariable(govar) {
		return true
	}

	return false
}

func (ma *GoMethodAccessVar) Init() ast.Stmt {
	return nil
}

func (ma *GoMethodAccessVar) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if ma.govar != nil {
		obj, is_nil := ma.govar.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if ma.method != nil {
		obj, is_nil := ma.method.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.method, err = convertToMethod(obj); err != nil {
				panic(err)
			}
		}
	}

	if ma.args != nil {
		obj, is_nil := ma.args.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.args, err = convertToMethodArgs(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, ma)
}

func (ma *GoMethodAccessVar) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoMethodAccessVar[")
	b.WriteString(ma.govar.String())
	b.WriteString("|")
	if ma.method != nil {
		var name string
		if ma.method.Receiver() != nil {
			name = ma.method.Receiver().Name()
		}
		fmt.Fprintf(b, "GoMethod[%v|%v#%d]", name, ma.method.Name(),
			ma.method.NumParameters())
	}
	b.WriteString("|")
	ma.args.WriteString(b)
	b.WriteString("]")
	return b.String()
}

func (ma *GoMethodAccessVar) VarType() *TypeData {
	return ma.method.VarType()
}

type GoMethodArguments struct {
	args []GoExpr
}

func NewGoMethodArguments(gs *GoState, owner GoMethodOwner,
	args []grammar.JObject) *GoMethodArguments {
	ma := &GoMethodArguments{}
	if args != nil && len(args) > 0 {
		ma.args = make([]GoExpr, len(args))
		for i, arg := range args {
			ma.args[i] = analyzeExpr(gs, owner, arg)
		}
	}
	return ma
}

func (ma *GoMethodArguments) ExprList() []ast.Expr {
	var args []ast.Expr

	if ma.args != nil && len(ma.args) > 0 {
		args = make([]ast.Expr, len(ma.args))
		for i, a := range ma.args {
			args[i] = a.Expr()
		}
	}

	return args
}

func (ma *GoMethodArguments) hasVariable(govar GoVar) bool {
	for _, a := range ma.args {
		if a.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (ma *GoMethodArguments) Length() int {
	return len(ma.args)
}

func (ma *GoMethodArguments) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, a := range ma.args {
		obj, is_nil := a.RunTransform(xform, prog, cls, ma)
		if !is_nil {
			var err error
			if ma.args[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, ma)
}

func (ma *GoMethodArguments) String() string {
	b := &bytes.Buffer{}
	ma.WriteString(b)
	return b.String()
}

func (ma *GoMethodArguments) WriteString(out io.Writer) {
	io.WriteString(out, "GoMethodArguments[")
	if ma.args != nil && len(ma.args) > 0 {
		for i, a := range ma.args {
			if i > 0 {
				io.WriteString(out, ", ")
			}
			io.WriteString(out, a.String())
		}
	}
	io.WriteString(out, "]")
}

type GoMethodOwner interface {
	AddConstant(con *GoConstant)
	AddMethod(mthd GoMethod)
	Constants() []ast.Decl
	FindMethod(name string, args *GoMethodArguments) GoMethod
	IsNil() bool
	Name() string
	Statics() []ast.Decl
	String() string
	Super() GoMethodOwner
	WriteString(out io.Writer, verbose bool)
}

var refcnt int

type GoMethodReference struct {
	class GoMethodOwner
	name string
	goname string
	args *GoMethodArguments
	ref *GoClassMethod
	cnt int
}

func NewGoMethodReference(class GoMethodOwner, name string,
	args *GoMethodArguments, verbose bool) *GoMethodReference {
	mthd := &GoMethodReference{class: class, name: name, goname: name,
		args: args, cnt: refcnt}
	refcnt++
	if class != nil && !class.IsNil() {
		class.AddMethod(mthd)
	} else if verbose {
		log.Printf("//ERR// No class for method ref %v\n", name)
	} else {
		log.Printf("//ERR// No class for method ref\n")
	}

	return mthd
}

func (mref *GoMethodReference) Arguments() []GoVar {
	if mref.ref != nil {
		return mref.ref.Arguments()
	}

	panic("Unimplemented")
}

func (mref *GoMethodReference) Body() *GoBlock {
	if mref.ref != nil {
		return mref.ref.Body()
	}

	panic("Unimplemented")
}

func (mref *GoMethodReference) Class() GoMethodOwner {
	if mref.ref != nil {
		return mref.ref.Class()
	}

	return mref.class
}

func (mref *GoMethodReference) Decl() ast.Decl {
	if mref.ref != nil {
		return mref.ref.Decl()
	}

	return nil
}

func (mref *GoMethodReference) Field() *ast.Field {
	return nil
}

func (mref *GoMethodReference) GoName() string {
	if mref.ref != nil {
		return mref.ref.GoName()
	}

	return mref.goname
}

func (mref *GoMethodReference) HasArguments(args *GoMethodArguments) bool {
	if mref.ref != nil {
		return mref.ref.HasArguments(args)
	}

	if mref.args.Length() != args.Length() {
		return false
	}

	for i, arg := range mref.args.args {
		at := arg.VarType()
		aat := args.args[i].VarType()

		if at == nil {
			if aat != nil {
				return false
			}
		}

		if !at.Equals(aat) {
			return false
		}
	}

	return true
}

func (mref *GoMethodReference) IsMethod(mthd GoMethod) bool {
	if mref.ref != nil {
		return mref.ref.IsMethod(mthd)
	}

	return mthd.Name() == mref.Name() &&
		mthd.NumParameters() == mref.NumParameters()
}

func (mref *GoMethodReference) MethodType() methodType {
	if mref.ref != nil {
		return mref.ref.MethodType()
	}

	return mt_method
}

func (mref *GoMethodReference) Name() string {
	if mref.ref != nil {
		return mref.ref.Name()
	}

	return mref.name
}

func (mref *GoMethodReference) NumParameters() int {
	if mref.ref != nil {
		return mref.ref.NumParameters()
	}

	if mref.args == nil {
		return 0
	}

	return mref.args.Length()
}

func (mref *GoMethodReference) Receiver() GoVar {
	if mref.ref != nil {
		return mref.ref.Receiver()
	}

	return nil
}

func (mref *GoMethodReference) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, mref)
}

func (mref *GoMethodReference) SetGoName(newname string) {
	if mref.ref != nil {
		mref.ref.SetGoName(newname)
	}

	mref.goname = newname
}

func (mref *GoMethodReference) SetOriginal(mthd *GoClassMethod) {
	if mref.ref != mthd {
		if mref.ref != nil {
			log.Printf("//ERR// found multiple method replacements");
		}

		mref.ref = mthd
	}
}

func (mref *GoMethodReference) String() string {
	b := &bytes.Buffer{}
	mref.WriteString(b)
	return b.String()
}

func (mref *GoMethodReference) VarType() *TypeData {
	if mref.ref != nil {
		return mref.ref.VarType()
	}

	return nil
}

func (mref *GoMethodReference) WriteString(out io.Writer) {
	io.WriteString(out, "GoMethodReference[")
	io.WriteString(out, mref.name)
	fmt.Fprintf(out, "#%d", mref.cnt)
	io.WriteString(out, "|")
	if mref.class != nil {
		mref.class.WriteString(out, false)
	}
	io.WriteString(out, "|")
	if mref.args != nil {
		io.WriteString(out, "Args[")
		for i, arg := range mref.args.args {
			if i > 0 {
				io.WriteString(out, ",")
			}

			if arg == nil {
				io.WriteString(out, "nil")
			} else if arg.VarType() == nil {
				fmt.Fprintf(out, "/* %v<%T> */", arg, arg)
			} else {
				io.WriteString(out, arg.VarType().Name())
			}
		}
		io.WriteString(out, "]")
	}
	io.WriteString(out, "|")
	if mref.ref != nil {
		fmt.Fprintf(out, "%T[%s/%s]", mref.ref, mref.ref.name, mref.ref.goname)
	}
	io.WriteString(out, "]")
}

type GoObjectDotName struct {
	obj GoExpr
	ref GoVar
}

func NewObjectDotName(odn *grammar.JObjectDotName, obj GoExpr, gs *GoState) *GoObjectDotName {
	govar := gs.findOrFakeVariable(odn.Name, "objdotname")

	log.Printf("//ERR// Inadequately wrapping odnobj %T\n", odn.Obj)
	return &GoObjectDotName{obj: obj, ref: govar}
}

func (odn *GoObjectDotName) Equals(govar GoVar) bool {
	if odn == govar || (odn.ref != nil && odn.ref.Equals(govar)) {
		return true
	}

	return false
}

func (odn *GoObjectDotName) Expr() ast.Expr {
	return odn.Ident()
}

func (odn *GoObjectDotName) GoName() string {
	return fmt.Sprintf("<<unimp_obj.nm_%T>>", odn.obj)
}

func (odn *GoObjectDotName) hasVariable(govar GoVar) bool {
	if odn.ref != nil && odn.ref.Equals(govar) {
		return true
	}

	if odn.obj != nil && odn.obj.hasVariable(govar) {
		return true
	}

	return false
}

func (odn *GoObjectDotName) Ident() *ast.Ident {
	return ast.NewIdent(odn.GoName())
}

func (odn *GoObjectDotName) Init() ast.Stmt {
	return nil
}

func (odn *GoObjectDotName) IsClassField() bool {
	return false
}

func (odn *GoObjectDotName) IsFinal() bool {
	return false
}

func (odn *GoObjectDotName) IsStatic() bool {
	return false
}

func (odn *GoObjectDotName) Name() string {
	return odn.GoName()
}

func (odn *GoObjectDotName) Receiver() string {
	return ""
}

func (odn *GoObjectDotName) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if odn.obj != nil {
		obj, is_nil := odn.obj.RunTransform(xform, prog, cls, odn)
		if !is_nil {
			var err error
			if odn.obj, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if odn.ref != nil {
		obj, is_nil := odn.ref.RunTransform(xform, prog, cls, odn)
		if !is_nil {
			var err error
			if odn.ref, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, odn)
}

func (odn *GoObjectDotName) SetGoName(newname string) {
	panic("Cannot change Go name for GoObjectDotName")
}

func (odn *GoObjectDotName) String() string {
	return fmt.Sprintf("GoObjectDotName[%s|%s]", odn.obj.String(),
		odn.ref.String())
}

func (odn *GoObjectDotName) Type() ast.Expr {
	return nil
}

func (odn *GoObjectDotName) VarType() *TypeData {
	return nil
}

type GoPrimitiveType struct {
	typedata *TypeData
}

func NewGoPrimitiveType(name string, dims int) *GoPrimitiveType {
	if name == "" {
		grammar.ReportError("PrimitiveType name cannot be empty")
	}

	return &GoPrimitiveType{typedata: NewTypeDataPrimitive(name, dims)}
}

func (pt *GoPrimitiveType) Expr() ast.Expr {
	return pt.typedata.Expr()
}

func (pt *GoPrimitiveType) hasVariable(govar GoVar) bool {
	return false
}

func (pt *GoPrimitiveType) Init() ast.Stmt {
	return nil
}

func (pt *GoPrimitiveType) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, pt)
}

func (pt *GoPrimitiveType) String() string {
	return "GoPrimitiveType[" + pt.typedata.String() + "]"
}

func (pt *GoPrimitiveType) VarType() *TypeData {
	return pt.typedata
}

type GoProgram struct {
	name string
	config *Config
	verbose bool

	import_map map[string]*GoImportPackage
	import_types map[string]*GoImportClass

	pkgname string
	enums []*GoEnumDefinition
	interfaces []GoInterface
	classes map[string]GoClass

	mgr *FileManager
	file *ast.File
}

func NewGoProgram(name string, config *Config, verbose bool) *GoProgram {
	return &GoProgram{name: name, config: config, verbose: verbose}
}

func (gp *GoProgram) addClass(cls GoClass) {
	key := makeClassKeyFromParts(cls.Parent(), cls.Name())

	if gp.classes == nil {
		gp.classes = make(map[string]GoClass)
	}
	gp.classes[key] = cls
}

func (gp *GoProgram) addEnum(enm *grammar.JEnumDecl) {
	if enm.Interfaces != nil && len(enm.Interfaces) > 0 {
		if gp.verbose {
			log.Printf("//ERR// Ignoring enum %s %d interfaces in %s\n",
				enm.Name, len(enm.Interfaces), gp.name)
		} else {
			log.Printf("//ERR// Ignoring enum interfaces\n")
		}
	}

	var constants []*GoEnumConstant
	if enm.Constants != nil && len(enm.Constants) > 0 {
		constants = make([]*GoEnumConstant, len(enm.Constants))
		for i, v := range enm.Constants {
			constants[i] = gp.analyzeEnumConstant(enm.Name, v)
		}
	}

	if enm.BodyDecl != nil && len(enm.BodyDecl) > 0 {
		log.Printf("//ERR// Ignoring enum body\n")
	}

	enum := &GoEnumDefinition{name: enm.Name, constants: constants}

	if gp.enums == nil {
		gp.enums = make([]*GoEnumDefinition, 0)
	}
	gp.enums = append(gp.enums, enum)
}

func (gp *GoProgram) addImport(pkgname string, clsname string) {
	if pkgname == gp.pkgname {
		return
	}

	if gp.import_map == nil {
		gp.import_map = make(map[string]*GoImportPackage)
	}

	var pkg *GoImportPackage
	if p, ok := gp.import_map[pkgname]; ok {
		pkg = p
	} else {
		pkg = NewGoImportPackage(pkgname)
		gp.import_map[pkgname] = pkg
	}

	if clsname != "" {
		cls := pkg.addClass(clsname)

		if gp.import_types == nil {
			gp.import_types = make(map[string]*GoImportClass)
		}

		if c, o := gp.import_types[clsname]; !o {
			gp.import_types[clsname] = cls
		} else if c.pkg != pkg {
			log.Printf("//ERR// Found multiple entries for %v\n", clsname)
		}
	}
}

func (gp *GoProgram) addInterface(iface *grammar.JInterfaceDecl) {
	if gp.interfaces == nil {
		gp.interfaces = make([]GoInterface, 0)
	}

	gi := NewGoInterfaceDefinition(iface.Name.String())

	added := false
	for i, x := range gp.interfaces {
		if x.Matches(iface.Name) {
			if _, ok := x.(*GoInterfaceDefinition); !ok {
				gp.interfaces[i] = gi
				added = true
			}
		}
	}
	if !added {
		gp.interfaces = append(gp.interfaces, gi)
	}

	for _, jobj := range iface.Body {
		switch j := jobj.(type) {
		case *grammar.JConstantDecl:
			analyzeConstant(&GoState{program: gp}, gi, j)
		case *grammar.JClassDecl:
			if gp.verbose {
				log.Printf("//ERR// Not adding %T to interface %s\n", j, gi.name)
			} else {
				log.Printf("//ERR// Not adding %T to interface\n", j)
			}
		case *grammar.JEnumDecl:
			if gp.verbose {
				log.Printf("//ERR// Not adding %T to interface %s\n", j, gi.name)
			} else {
				log.Printf("//ERR// Not adding %T to interface\n", j)
			}
		case *grammar.JInterfaceDecl:
			if gp.verbose {
				log.Printf("//ERR// Not adding %T to interface %s\n", j, gi.name)
			} else {
				log.Printf("//ERR// Not adding %T to interface\n", j)
			}
		case *grammar.JInterfaceMethodDecl:
			gi.AddMethod(NewGoInterfaceMethod(gp, gi.name, j))
		default:
			grammar.ReportCastError("InterfaceDecl", jobj)
		}
	}
}

func (gp *GoProgram) addInterfaceReference(name *grammar.JTypeName) *GoInterfaceReference {
	iface := &GoInterfaceReference{name: name, gp: gp}

	if gp.interfaces == nil {
		gp.interfaces = make([]GoInterface, 0)
	}
	gp.interfaces = append(gp.interfaces, iface)

	return iface
}

func (gp *GoProgram) Analyze(pgm *grammar.JProgramFile) {
	gp.setPackage(pgm)

	gp.analyzeImports(pgm)

	gp.analyzeCode(pgm)

	gp.finalize()
}

func (gp *GoProgram) analyzeCode(pgm *grammar.JProgramFile) {
	if pgm == nil || pgm.TypeDecls == nil || len(pgm.TypeDecls) == 0 {
		return
	}

	var gs *GoState

	for _, tobj := range pgm.TypeDecls {
		switch t := tobj.(type) {
		case *grammar.JClassDecl:
			if gs == nil {
				gs = &GoState{program: gp}
			}
			gs.addClassDecl(nil, t)
		case *grammar.JEnumDecl:
			gp.addEnum(t)
		case *grammar.JInterfaceDecl:
			gp.addInterface(t)
		case *grammar.JUnimplemented:
			log.Printf("//ERR// Ignoring unimplemented object %s\n", t.TypeStr)
		default:
			log.Printf("//ERR// Ignoring unknown type_decl %T\n", tobj)
		}

		if gs != nil && gs.classes != nil {
			if gp.classes == nil {
				gp.classes = make(map[string]GoClass)
			}

			for k, v := range gs.classes {
				if kcls, ok := gp.classes[k]; ok {
					if kcls != v {
						if gp.verbose {
							log.Printf("GoState class %v<%T> overrides" +
								" GoProgram class %v<%T>\n", v, v, kcls, kcls)
						} else {
							log.Printf("GoState class overrides" +
								" GoProgram class\n")
						}
					}
				}
				gp.classes[k] = v
			}
		}
	}
}

func (gp *GoProgram) analyzeEnumConstant(name string, con *grammar.JEnumConstant) *GoEnumConstant {
	if con.Annotations != nil && len(con.Annotations) > 0 {
		if gp.verbose {
			log.Printf("//ERR// ignoring enumconst %v.%v annotations\n",
				name, con.Name)
		} else {
			log.Printf("//ERR// ignoring enumconst annotations\n")
		}
	}

	if con.ArgList != nil && len(con.ArgList) > 0 {
		if gp.verbose {
			log.Printf("//ERR// ignoring enumconst %v.%v arguments\n",
				name, con.Name)
		} else {
			log.Printf("//ERR// ignoring enumconst arguments\n")
		}
	}

	if con.Body != nil && len(con.Body) > 0 {
		if gp.verbose {
			log.Printf("//ERR// ignoring enumconst %v.%v body\n", name, con.Name)
		} else {
			log.Printf("//ERR// ignoring enumconst body\n")
		}
	}

	return &GoEnumConstant{name: con.Name}
}

func (gp *GoProgram) analyzeImports(pgm *grammar.JProgramFile) {
	if pgm == nil || pgm.Imports == nil {
		return
	}

	for _, iobj := range pgm.Imports {
		if jimp, ok := iobj.(*grammar.JImportStmt); !ok {
			grammar.ReportCastError("JImportStmt", iobj)
		} else {
			pkgstr := jimp.Name.PackageString()
			iname := gp.config.findPackage(pkgstr)
			if iname == "" {
				gfc := NewGoFakeClass(jimp.Name.LastType())
				gfc.SetPackage(pkgstr)
				gp.addClass(gfc)
				log.Printf("Faking import for %v\n", jimp.Name.String())
				continue
			}

			gp.addImport(iname, jimp.Name.LastType())
		}
	}
}

func (gp *GoProgram) createTypeData(typename *grammar.JTypeName,
	type_args []*grammar.JTypeArgument, dims int) *TypeData {

	typestr := typename.String()
	if typename.IsPrimitive() || typestr == "String" {
		if type_args != nil && len(type_args) > 0 {
			panic(fmt.Sprintf("Found type_args for %v\n", typestr))
		}

		return NewTypeDataPrimitive(typestr, dims)
	}

	return NewTypeDataObject(gp, typestr, dims)
}

func (gp *GoProgram) Decls() []ast.Decl {
	decls := make([]ast.Decl, 0)

	if len(gp.import_map) > 0 {
		keys := make([]string, len(gp.import_map))
		i := 0
		for k, _ := range gp.import_map {
			keys[i] = k
			i++
		}
		sort.Strings(keys)

		// add imports
		if true {
			// add each import on a separate line
			for _, k := range keys {
				specs := make([]ast.Spec, 1)
				specs[0] = gp.import_map[k].ImportSpec()
				decls = append(decls, &ast.GenDecl{Tok: token.IMPORT,
					Specs: specs})
			}
		} else {
			// add an import block -- this doesn't work!
			specs := make([]ast.Spec, len(gp.import_map))

			lpos := gp.mgr.NextPos()
			for i, k := range keys {
				specs[i] = gp.import_map[k].ImportSpec()
				gp.mgr.NextPos()
			}
			rpos := gp.mgr.NextPos()

			decls = append(decls, &ast.GenDecl{Tok: token.IMPORT, Lparen: lpos,
				Specs: specs, Rparen: rpos})
		}
	}

	if gp.enums != nil && len(gp.enums) > 0 {
		sort.Sort(enumSlice(gp.enums))

		for _, enum := range gp.enums {
			edecls := enum.Decls(gp)
			if edecls != nil {
				decls = append(decls, edecls...)
			}
		}
	}

	if gp.interfaces != nil && len(gp.interfaces) > 0 {
		sort.Sort(InterfaceSlice(gp.interfaces))

		for _, iface := range gp.interfaces {
			consts := iface.Constants()
			if consts != nil {
				decls = append(decls, consts...)
			}

			idecl := iface.Decl()
			if idecl != nil {
				decls = append(decls, idecl)
			}
		}
	}

	if gp.classes != nil && len(gp.classes) > 0 {
		keys := make([]string, len(gp.classes))
		i := 0
		for k, _ := range gp.classes {
			keys[i] = k
			i++
		}
		sort.Strings(keys)

		for _, k := range keys {
			class := gp.classes[k]
			consts := class.Constants()
			if consts != nil {
				decls = append(decls, consts...)
			}

			stats := class.Statics()
			if stats != nil {
				decls = append(decls, stats...)
			}

			cd := class.Decls()
			// GoClassReference.Decls() can return nil
			if cd != nil {
				decls = append(decls, class.Decls()...)
			}
		}
	}

	return decls
}

func (gp *GoProgram) Dump(out io.Writer) {
	format.Node(out, gp.FileSet(), gp.File())
}

func (gp *GoProgram) DumpTree() {
	dumper.Dump("Parse Tree", gp.File())
}

func (gp *GoProgram) File() *ast.File {
	if gp.mgr == nil {
		gp.mgr = NewFileManager(gp.name)
	}

	if gp.file == nil {
		gp.file = &ast.File{Name: ast.NewIdent(gp.pkgname), Decls: gp.Decls(),
			Imports: gp.Imports(), Scope: nil, Unresolved: nil, Comments: nil}
	}
	return gp.file
}

func (gp *GoProgram) FileSet() *token.FileSet {
	if gp.mgr == nil {
		gp.mgr = NewFileManager(gp.name)
	}

	return gp.mgr.FileSet()
}

func (gp *GoProgram) finalize() {
	// finalize all interfaces
	for _, iface := range gp.interfaces {
		iface.finalize(gp)
	}
	// finalize all classes
	for _, cls := range gp.classes {
		cls.finalize(gp)
	}
}

func (gp *GoProgram) findClass(name string) GoClass {
	key := makeClassKeyFromParts(nil, name)

	if gp.classes != nil {
		if ocls, ok := gp.classes[key]; ok {
			return ocls
		}
	}

	return nil
}

func (gp *GoProgram) findInterface(name *grammar.JTypeName) GoInterface {
	if gp.interfaces != nil {
		for _, iface := range gp.interfaces {
			if iface.Matches(name) || iface.Matches(name) {
				return iface
			}
		}
	}

	return nil
}

func (gp *GoProgram) Imports() [] *ast.ImportSpec {
	imports := make([]*ast.ImportSpec, len(gp.import_map))

	var i int
	for _, v := range gp.import_map {
		imports[i] = v.ImportSpec()
		i += 1
	}

	return imports
}

func (gp *GoProgram) ImportedType(name string) string {
	if cls, ok := gp.import_types[name]; ok {
		return cls.FullName()
	}

	return ""
}

func (gp *GoProgram) IsInterface(name string) bool {
	if gp.config == nil {
		return false
	}

	return gp.config.isInterface(name)
}

func (gp *GoProgram) Name() string {
	return gp.name
}

func (gp *GoProgram) Receiver(class string) string {
	var rcvr string
	if gp.config != nil {
		var rkey string
		if gp.pkgname == "" {
			rkey = class
		} else {
			rkey = gp.pkgname + "." + class
		}

		rcvr = gp.config.receiver(rkey)
		if rcvr != "" {
			return rcvr
		}
	}

	return "rcvr"
}

func (gp *GoProgram) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, cls := range gp.classes {
		obj, is_nil := cls.RunTransform(xform, gp, cls, gp)
		if !is_nil {
			if cla, ok := obj.(GoClass); ok {
				gp.classes[i] = cla
			} else {
				panic(fmt.Errorf("%v<%T> is not a GoClass", obj, obj))
			}
		}
	}

	return xform(parent, prog, cls, gp)
}

func (gp *GoProgram) setPackage(pgm *grammar.JProgramFile) {
	if pgm != nil && pgm.Pkg != nil {
		gp.pkgname = gp.config.findPackage(pgm.Pkg.Name.String())
		if gp.pkgname == "" {
			gp.pkgname = pgm.Pkg.Name.String()
		}
	}
	if gp.pkgname == "" {
		gp.pkgname = "main"
	}
}

func (gp *GoProgram) Write(topdir string) error {
	var dirpath string
	if gp.pkgname == "" || gp.pkgname == "main" {
	    dirpath = topdir
	} else {
	    dirpath = path.Join(topdir, gp.pkgname)
	}

	if err := os.MkdirAll(dirpath, os.ModeDir|0755); err != nil {
		return err
	}

	path := path.Join(dirpath, gp.name)
	fd, err := os.Create(path)
	if err != nil {
		return err
	}

	format.Node(fd, gp.FileSet(), gp.File())

	fd.Close()

	return nil
}

func (gp *GoProgram) WriteString(out io.Writer) {
	io.WriteString(out, "GoProgram[")
	io.WriteString(out, gp.pkgname)
	io.WriteString(out, "|")

	for i, enum := range gp.enums {
		if i > 0 {
			io.WriteString(out, ",")
		}
		enum.WriteString(out)
	}

	for i, iface := range gp.interfaces {
		if i > 0 {
			io.WriteString(out, ",")
		}
		iface.WriteString(out, true)
	}

	i := 0
	for _, cls := range gp.classes {
		if i > 0 {
			io.WriteString(out, ",")
		}
		i += 1

		cls.WriteString(out, true)
	}

	io.WriteString(out, "]")
}

type GoReference struct {
	cls GoMethodOwner
}

func (ref *GoReference) Expr() ast.Expr {
	return &ast.UnaryExpr{Op: token.AND,
		X: &ast.CompositeLit{Type: ast.NewIdent(ref.cls.Name())}}
}

func (ref *GoReference) hasVariable(govar GoVar) bool {
	return false
}

func (ref *GoReference) Init() ast.Stmt {
	return nil
}

func (ref *GoReference) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, ref)
}

func (ref *GoReference) String() string {
	return "GoReference[" + ref.cls.String() + "]"
}

func (ref *GoReference) VarType() *TypeData {
	panic("GoReference.VarType() unimplemented")
}

type GoReturn struct {
	expr GoExpr
}

func (rtn *GoReturn) hasVariable(govar GoVar) bool {
	return rtn.expr != nil && rtn.expr.hasVariable(govar)
}

func (rtn *GoReturn) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if rtn.expr != nil {
		obj, is_nil := rtn.expr.RunTransform(xform, prog, cls, rtn)
		if !is_nil {
			var err error
			if rtn.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, rtn)
}

func (rtn *GoReturn) Stmts() []ast.Stmt {

	var results []ast.Expr
	if rtn.expr != nil {
		results = make([]ast.Expr, 1)
		results[0] = rtn.expr.Expr()
	}

	return []ast.Stmt { &ast.ReturnStmt{Results: results}, }
}

func (rtn *GoReturn) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoReturn[")
	if rtn.expr != nil {
		b.WriteString(rtn.expr.String())
	}
	b.WriteString("]")
	return b.String()
}

type GoSelector struct {
	expr GoExpr
	sel GoVar
}

func NewGoSelector(expr GoExpr, sel GoVar) *GoSelector {
	return &GoSelector{expr: expr, sel: sel}
}

func (sel *GoSelector) Equals(govar GoVar) bool {
	if gs, ok := govar.(*GoSelector); ok {
		if gs.sel.Equals(sel.sel) {
			return gs.expr == sel.expr
		}
	}

	return false
}

func (sel *GoSelector) Expr() ast.Expr {
	return &ast.SelectorExpr{X: sel.expr.Expr(), Sel: sel.sel.Ident()}
}

func (sel *GoSelector) GoName() string {
	panic("GoSelector.GoName() unimplemented")
}

func (sel *GoSelector) hasVariable(govar GoVar) bool {
	if sel.sel != nil && sel.sel.Equals(govar) {
		return true
	}

	if sel.expr != nil && sel.expr.hasVariable(govar) {
		return true
	}

	return false
}

func (fv *GoSelector) Ident() *ast.Ident {
	panic("GoSelector.Ident() unimplemented")
}

func (sel *GoSelector) Init() ast.Stmt {
	return nil
}

func (sel *GoSelector) IsClassField() bool {
	return false
}

func (sel *GoSelector) IsFinal() bool {
	return false
}

func (sel *GoSelector) IsStatic() bool {
	return false
}

func (sel *GoSelector) Name() string {
	if gv, ok := sel.sel.(*GoVarData); ok {
		return gv.Name()
	}

	panic(fmt.Sprintf("GoSelector.Name() unimplemented for %v", sel.sel))
}

func (sel *GoSelector) Receiver() string {
	if gv, ok := sel.expr.(*GoVarData); ok {
		return gv.Name()
	}

	panic(fmt.Sprintf("GoSelector.Receiver() unimplemented for %v", sel.expr))
}

func (sel *GoSelector) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if sel.expr != nil {
		obj, is_nil := sel.expr.RunTransform(xform, prog, cls, sel)
		if !is_nil {
			var err error
			if sel.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if sel.sel != nil {
		obj, is_nil := sel.sel.RunTransform(xform, prog, cls, sel)
		if !is_nil {
			var err error
			if sel.sel, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, sel)
}

func (sel *GoSelector) SetGoName(newname string) {
	panic("GoSelector.SetGoName() unimplemented")
}

func (sel *GoSelector) String() string {
	return fmt.Sprintf("GoSelector[expr %v, sel %v]", sel.expr, sel.sel)
}

func (sel *GoSelector) Type() ast.Expr {
	panic("GoSelector.Type() unimplemented")
}

func (sel *GoSelector) VarType() *TypeData {
	return sel.expr.VarType()
}

type GoState struct {
	parent *GoState
	program *GoProgram
	class *GoClassDefinition
	vars map[string]GoVar
	classes map[string]GoClass
}

func NewGoState(parent *GoState) *GoState {
	return &GoState{parent: parent}
}

func (gs *GoState) addClass(nref GoClass) {
	if gs.classes == nil {
		gs.classes = make(map[string]GoClass, 0)
	}

	key := makeClassKey(nref)

	// if this isn't first reference, we need to do some work
	if ocls, ok := gs.classes[key]; !ok {
		gs.classes[key] = nref
	} else {
		// panic if this is a class reference
		if !ocls.IsReference() {
			panic(fmt.Sprintf("GoProgram already contains class reference" +
				" for %v", nref.Name()))
		}

		// grab previous class reference
		var ref *GoClassReference
		if ref, ok = ocls.(*GoClassReference); !ok {
			panic(fmt.Sprintf("GoProgram contains multiple instances of %v",
				nref.Name()))
		}

		// make sure new class is a class definition
		var cls *GoClassDefinition
		if cls, ok = nref.(*GoClassDefinition); !ok {
			panic(fmt.Sprintf("New %v should be a class definition, not %T",
				nref, nref))
		}

		// if reference points to a class, make sure it's this class definition
		if ref.cls != nil && ref.cls != cls {
			panic(fmt.Sprintf("Class reference refers to %v, not %v",
				ref.cls, cls))
		}

		// point to class definition
		ref.cls = cls

		// update map
		gs.classes[key] = cls
	}
}

func (gs *GoState) addClassDecl(parent GoMethodOwner, jcls *grammar.JClassDecl) {
	cls := NewGoClassDefinition(gs.Program(), parent, jcls.Name)
	gs.addClass(cls)

	if jcls.Extends != nil {
		if jcls.Extends.Dims != 0 {
			log.Printf("Class %v cannot extend array %v with dim=%d",
				jcls.Name, jcls.Extends.Name, jcls.Extends.Dims)
			jcls.Extends.Dims = 0
		}

		if jcls.Extends.TypeArgs != nil && len(jcls.Extends.TypeArgs) > 0 {
			if gs.Program().verbose {
				log.Printf("//ERR// Not handling %v type_args in extended" +
					" class %v\n", jcls.Extends.Name, jcls.Name)
			} else {
				log.Printf("//ERR// Not handling type_args in extended class\n")
			}
		}

		extname := jcls.Extends.Name.LastType()
		cls.super = gs.Program().findClass(extname)
		if cls.super == nil {
			cls.super = NewGoFakeClass(extname)
			gs.Program().addClass(cls.super)
		}
	}

	if jcls.Interfaces != nil && len(jcls.Interfaces) > 0 {
		ifaces := make([]GoInterface, len(jcls.Interfaces))
		for i, iname := range jcls.Interfaces {
			iface := gs.Program().findInterface(iname)
			if iface == nil {
				iface = gs.Program().addInterfaceReference(iname)
			}
			ifaces[i] = iface
		}
		cls.interfaces = ifaces
	}

	gs2 := NewGoState(gs)
	gs2.class = cls

	for _, jobj := range jcls.Body {
		switch j := jobj.(type) {
		case *grammar.JClassBody:
			analyzeClassBody(gs2, cls, j)
		case *grammar.JBlock:
			blk := analyzeBlock(gs2, cls, j)
			if blk != nil {
				m := &GoClassMethod{class: cls, name: "init", goname: "init",
					rcvr: nil, method_type: mt_static, body: blk}
				cls.AddMethod(m)
			}
		case *grammar.JEmpty:
			; // do nothing
		default:
			grammar.ReportCastError("JClassDecl", jobj)
		}
	}
}

func (gs *GoState) addVariable(name string, modifiers *grammar.JModifiers, dims int,
	typespec *grammar.JReferenceType, class_field bool) GoVar {
	goname := fixName(name, modifiers)

	var vartype *TypeData
	if typespec != nil {
		// "String[] a[]" creates a 2D array of Strings
		vartype = gs.Program().createTypeData(typespec.Name,
			typespec.TypeArgs, typespec.Dims + dims)
	}

	if gs.vars == nil {
		gs.vars = make(map[string]GoVar)
	}

	val, ok := gs.vars[name]
	if ok {
		return val
	}

	var rcvr string
	if class_field {
		rcvr = gs.Receiver()
	}

	var is_static bool
	if modifiers != nil {
		is_static = modifiers.IsSet(grammar.ModStatic)
	}

	var is_final bool
	if modifiers != nil {
		is_final = modifiers.IsSet(grammar.ModFinal)
	}

	govar := &GoVarData{rcvr: rcvr, name: name, goname: goname,
		vartype: vartype, class_field: class_field, is_static: is_static,
		is_final: is_final}
	gs.vars[name] = govar

	return govar
}

func (gs *GoState) addVariableDecl(vd *grammar.JVariableDecl, class_field bool) GoVar {
	return gs.addVariable(vd.Name, vd.Modifiers, vd.Dims, vd.TypeSpec,
		class_field)
}

func (gs *GoState) Class() *GoClassDefinition {
	if gs.class != nil {
		return gs.class
	}

	if gs.parent != nil {
		return gs.parent.Class()
	}

	return nil
}

func (gs *GoState) ClassName() string {
	cls := gs.Class()
	if cls != nil {
		return cls.name
	}

	return ""
}

func (gs *GoState) dumpVariables(fd io.Writer, indentLevel int) {
	var indent string
	for i := 0; i < indentLevel; i++ {
		indent = indent + " "
	}

	for _, v := range gs.vars {
		fmt.Fprintf(fd, indent + v.String() + "\n")
	}
	if gs.parent != nil {
		gs.parent.dumpVariables(fd, indentLevel + 1)
	}
}

func (gs *GoState) findClass(parent GoMethodOwner, name string) GoClass {
	key := makeClassKeyFromParts(parent, name)

	if gs.classes != nil {
		if ocls, ok := gs.classes[key]; ok {
			return ocls
		}
	}

	if gs.parent != nil {
		return gs.parent.findClass(parent, name)
	}

	if gs.Program() != nil {
		return gs.Program().findClass(name)
	}

	return nil
}

func (gs *GoState) findOrFakeVariable(typename *grammar.JTypeName,
	origtype string) GoVar {
	govar := gs.findVariable(typename)
	if govar != nil {
		return govar
	}

	if gs.Program().verbose {
		log.Printf("//ERR// Cannot find %v var %v\n", origtype, typename)
	}

	return NewFakeVar(fmt.Sprintf("<<unimp_%v_%v>>", origtype, typename), nil,
		0)
}

func (gs *GoState) findVariable(typename *grammar.JTypeName) GoVar {
	if !typename.IsDotted() {
		val, ok := gs.vars[typename.String()]
		if ok {
			return val
		}
	} else {
		// find class.attribute instance
		val, ok := gs.vars[typename.FirstType()]
		if ok {
			return &GoClassAttribute{govar: val,
				suffix: typename.NotFirst().String()}
		}
	}

	if gs.parent != nil {
		if govar := gs.parent.findVariable(typename); govar != nil {
			return govar
		}
	}

	cls := gs.Class()
	if cls != nil {
		return cls.findVariable(typename)
	}

	return nil
}

func (gs *GoState) findVariableString(name string) GoVar {
	val, ok := gs.vars[name]
	if ok {
		return val
	}

	if gs.parent != nil {
		return gs.parent.findVariableString(name)
	}

	return nil
}

func (gs *GoState) Receiver() string {
	return gs.Program().Receiver(gs.ClassName())
}

func (gs *GoState) Program() *GoProgram {
	if gs.program != nil {
		return gs.program
	}

	if gs.parent != nil {
		return gs.parent.Program()
	}

	return nil
}

type GoStatement interface {
	GoObject
	hasVariable(govar GoVar) bool
	Stmts() []ast.Stmt
	String() string
}

type GoStatic struct {
	init *GoVarInit
}

func (stat *GoStatic) Decl() ast.Decl {
	var vtype ast.Expr
	var vals []ast.Expr
	if stat.init.expr == nil {
		vtype = stat.init.govar.Type()
		vals = nil
	} else {
		vtype = nil
		vals = []ast.Expr{ stat.init.Expr(), }
	}
	vspec := &ast.ValueSpec{Names: []*ast.Ident{ stat.init.govar.Ident(), },
		Type: vtype, Values: vals}
	return &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{ vspec, }}
}

func (stat *GoStatic) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if stat.init != nil {
		obj, is_nil := stat.init.RunTransform(xform, prog, cls, stat)
		if !is_nil {
			var err error
			if stat.init, err = convertToVarInit(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, stat)
}

func (stat *GoStatic) String() string {
	return "GoStatic[" + stat.init.String() + "]"
}

type GoSwitch struct {
	expr GoExpr
	cases []*GoSwitchCase
}

func (gsw *GoSwitch) hasVariable(govar GoVar) bool {
	if gsw.expr != nil && gsw.expr.hasVariable(govar) {
		return true
	}

	if gsw.cases != nil {
		for _, c := range gsw.cases {
			if c.hasVariable(govar) {
				return true
			}
		}
	}

	return false
}

func (gsw *GoSwitch) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gsw.expr != nil {
		obj, is_nil := gsw.expr.RunTransform(xform, prog, cls, gsw)
		if !is_nil {
			var err error
			if gsw.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	for i, c := range gsw.cases {
		obj, is_nil := c.RunTransform(xform, prog, cls, gsw)
		if !is_nil {
			if ca, ok := obj.(*GoSwitchCase); ok {
				gsw.cases[i] = ca
			} else {
				panic(fmt.Errorf("%v<%T> is not a *GoSwitchCase", obj, obj))
			}
		}
	}

	return xform(parent, prog, cls, gsw)
}

func (gsw *GoSwitch) Stmts() []ast.Stmt {
	var totLabels int
	for _, c := range gsw.cases {
		totLabels += len(c.labels)
	}

	cases := make([]ast.Stmt, totLabels)

	nextCase := 0
	for _, c := range gsw.cases {
		var label *GoSwitchLabel
		if len(c.labels) == 1 {
			label = c.labels[0]
		} else {
			for _, l := range c.labels {
				if label != nil {
					body := make([]ast.Stmt, 1)
					body[0] = &ast.BranchStmt{Tok: token.FALLTHROUGH}
					cases[nextCase] = &ast.CaseClause{List: label.List(),
						Body: body}
					nextCase++
				}
				label = l
			}
		}

		body := make([]ast.Stmt, 0)
		if c.stmts == nil || len(c.stmts) == 0 {
			body = append(body, &ast.EmptyStmt{})
		} else {
			for _, s := range c.stmts {
				body = append(body, s.Stmts()...)
			}
		}

		cases[nextCase] = &ast.CaseClause{List: label.List(), Body: body}
		nextCase++
	}

	return []ast.Stmt { &ast.SwitchStmt{Tag: gsw.expr.Expr(),
		Body: &ast.BlockStmt{List: cases}}, }
}

func (gsw *GoSwitch) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoSwitch[")
	b.WriteString(gsw.expr.String())
	b.WriteString("|")
	for i, c := range gsw.cases {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(c.String())
	}
	b.WriteString("]")
	return b.String()
}

type GoSwitchCase struct {
	labels []*GoSwitchLabel
	stmts []GoStatement
}

func (gsc *GoSwitchCase) hasVariable(govar GoVar) bool {
	for _, l := range gsc.labels {
		if l.hasVariable(govar) {
			return true
		}
	}

	for _, s := range gsc.stmts {
		if s.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (gsc *GoSwitchCase) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	for i, l := range gsc.labels {
		obj, is_nil := l.RunTransform(xform, prog, cls, gsc)
		if !is_nil {
			if lab, ok := obj.(*GoSwitchLabel); ok {
				gsc.labels[i] = lab
			} else {
				panic(fmt.Errorf("%v<%T> is not a *GoSwitchLabel", obj, obj))
			}
		}
	}

	for i, s := range gsc.stmts {
		obj, is_nil := s.RunTransform(xform, prog, cls, gsc)
		if !is_nil {
			var err error
			if gsc.stmts[i], err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gsc)
}

func (gsc *GoSwitchCase) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoSwitchCase[")
	for i, c := range gsc.labels {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(c.String())
	}
	b.WriteString("|")
	for i, s := range gsc.stmts {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(s.String())
	}
	b.WriteString("]")
	return b.String()
}

type GoSwitchLabel struct {
	is_default bool
	expr GoExpr
}

func (gsl *GoSwitchLabel) hasVariable(govar GoVar) bool {
	if gsl.expr != nil && gsl.expr.hasVariable(govar) {
		return true
	}

	return false
}

func (gsl *GoSwitchLabel) List() []ast.Expr {
	var list []ast.Expr

	if !gsl.is_default {
		list = make([]ast.Expr, 1)
		list[0] = gsl.expr.Expr()
	}

	return list
}

func (gsl *GoSwitchLabel) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gsl.expr != nil {
		obj, is_nil := gsl.expr.RunTransform(xform, prog, cls, gsl)
		if !is_nil {
			var err error
			if gsl.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gsl)
}

func (gsl *GoSwitchLabel) String() string {
	if gsl.is_default {
		return "default"
	}

	return gsl.expr.String()
}

type GoSynchronized struct {
	expr GoExpr
	block *GoBlock
}

func (sync *GoSynchronized) hasVariable(govar GoVar) bool {
	if sync.expr != nil && sync.expr.hasVariable(govar) {
		return true
	}

	if sync.block != nil && sync.block.hasVariable(govar) {
		return true
	}

	return false
}

func (sync *GoSynchronized) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if sync.expr != nil {
		obj, is_nil := sync.expr.RunTransform(xform, prog, cls, sync)
		if !is_nil {
			var err error
			if sync.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if sync.block != nil {
		obj, is_nil := sync.block.RunTransform(xform, prog, cls, sync)
		if !is_nil {
			var err error
			if sync.block, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, sync)
}

func (sync *GoSynchronized) Stmts() []ast.Stmt {
	sargs := make([]ast.Expr, 1)
	sargs[0] = sync.expr.Expr()

	sexpr := &ast.CallExpr{Fun: ast.NewIdent("synchronized"),
		Args: sargs}

	return []ast.Stmt { &ast.IfStmt{Cond: sexpr,
		Body: sync.block.BlockStmt()}, }
}

func (sync *GoSynchronized) String() string {
	return "GoSynchronized[" + sync.expr.String() + "|" +
		sync.block.String() + "]"
}

type GoThrow struct {
	expr GoExpr
}

func (thr *GoThrow) hasVariable(govar GoVar) bool {
	return thr.expr != nil && thr.expr.hasVariable(govar)
}

func (thr *GoThrow) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if thr.expr != nil {
		obj, is_nil := thr.expr.RunTransform(xform, prog, cls, thr)
		if !is_nil {
			var err error
			if thr.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, thr)
}

func (thr *GoThrow) Stmts() []ast.Stmt {
	targs := []ast.Expr { thr.expr.Expr(), }

	x := &ast.CallExpr{Fun: ast.NewIdent("throw"), Args: targs}

	return []ast.Stmt { &ast.ExprStmt{X: x }, }
}

func (thr *GoThrow) String() string {
	return "GoThrow[" + thr.expr.String() + "]"
}

type GoTry struct {
	block *GoBlock
	catches []*GoTryCatch
	finally *GoBlock
}

func (try *GoTry) hasVariable(govar GoVar) bool {
	if try.block != nil && try.block.hasVariable(govar) {
		return true
	}

	for _, c := range try.catches {
		if c.hasVariable(govar) {
			return true
		}
	}

	if try.finally != nil && try.finally.hasVariable(govar) {
		return true
	}

	return false
}

func (try *GoTry) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if try.block != nil {
		obj, is_nil := try.block.RunTransform(xform, prog, cls, try)
		if !is_nil {
			var err error
			if try.block, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	for i, c := range try.catches {
		obj, is_nil := c.RunTransform(xform, prog, cls, try)
		if !is_nil {
			if ca, ok := obj.(*GoTryCatch); ok {
				try.catches[i] = ca
			} else {
				panic(fmt.Errorf("%v<%T> is not a *GoTryCatch", obj, obj))
			}
		}
	}

	if try.finally != nil {
		obj, is_nil := try.finally.RunTransform(xform, prog, cls, try)
		if !is_nil {
			var err error
			if try.finally, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, try)
}

func (try *GoTry) Stmts() []ast.Stmt {
	tfake := &ast.CallExpr{Fun: ast.NewIdent("try"), Args: make([]ast.Expr, 0)}
	top_if := &ast.IfStmt{Cond: tfake, Body: try.block.BlockStmt()}

	cur_if := top_if

	if try.catches != nil && len(try.catches) > 0 {
		for _, c := range try.catches {
			cfake := fmt.Sprintf("catch_%s", c.govar.VarType().Name())
			cargs := make([]ast.Expr, 1)
			cargs[0] = ast.NewIdent(c.govar.GoName())

			expr := &ast.CallExpr{Fun: ast.NewIdent(cfake), Args: cargs}

			c_if := &ast.IfStmt{Cond: expr, Body: c.block.BlockStmt()}

			cur_if.Else = c_if
			cur_if = c_if
		}
	}

	if try.finally != nil {
		expr := &ast.CallExpr{Fun: ast.NewIdent("finally"),
			Args: make([]ast.Expr, 0)}
		f_if := &ast.IfStmt{Cond: expr, Body: try.finally.BlockStmt()}

		cur_if.Else = f_if
		cur_if = f_if
	}

	return []ast.Stmt { top_if, }
}

func (try *GoTry) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoTry[")
	b.WriteString(try.block.String())
	b.WriteString("|")
	if try.catches != nil && len(try.catches) > 0 {
		for i, c := range try.catches {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(c.String())
		}
	}
	b.WriteString("|")
	if try.finally != nil {
		b.WriteString(try.finally.String())
	}
	b.WriteString("]")
	return b.String()
}

type GoTryCatch struct {
	govar GoVar
	block *GoBlock
}

func (gtc *GoTryCatch) hasVariable(govar GoVar) bool {
	if gtc.govar != nil && gtc.govar.Equals(govar) {
		return true
	}

	if gtc.block != nil && gtc.block.hasVariable(govar) {
		return true
	}

	return false
}

func (gtc *GoTryCatch) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gtc.govar != nil {
		obj, is_nil := gtc.govar.RunTransform(xform, prog, cls, gtc)
		if !is_nil {
			var err error
			if gtc.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if gtc.block != nil {
		obj, is_nil := gtc.block.RunTransform(xform, prog, cls, gtc)
		if !is_nil {
			var err error
			if gtc.block, err = convertToBlock(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gtc)
}

func (gtc *GoTryCatch) String() string {
	return "GoTryCatch[" + gtc.govar.String() + "|" + gtc.block.String() + "]"
}

type GoUnaryExpr struct {
	op token.Token
	x GoExpr
}

func (uex *GoUnaryExpr) Expr() ast.Expr {
	return uex.UnaryExpr()
}

func (uex *GoUnaryExpr) hasVariable(govar GoVar) bool {
	if uex.x != nil && uex.x.hasVariable(govar) {
		return true
	}

	return false
}

func (uex *GoUnaryExpr) Init() ast.Stmt {
	return nil
}

func (uex *GoUnaryExpr) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	obj, is_nil := uex.x.RunTransform(xform, prog, cls, uex)
	if !is_nil {
		var err error
		if uex.x, err = convertToExpr(obj); err != nil {
			panic(err)
		}
	}

	return xform(parent, prog, cls, uex)
}

func (uex *GoUnaryExpr) Stmts() []ast.Stmt {
	var stmt ast.Stmt
	if uex.op == token.INC || uex.op == token.DEC {
		stmt = &ast.IncDecStmt{X: uex.x.Expr(), Tok: uex.op}
	} else {
		stmt = &ast.ExprStmt{X: uex.Expr()}
	}

	return []ast.Stmt {stmt, }
}

func (uex *GoUnaryExpr) String() string {
	return "GoUnaryExpr[" + uex.op.String() + "|" + uex.x.String() + "]"
}

func (uex *GoUnaryExpr) UnaryExpr() *ast.UnaryExpr {
	return &ast.UnaryExpr{Op: uex.op, X: uex.x.Expr()}
}

func (uex *GoUnaryExpr) VarType() *TypeData {
	return uex.x.VarType()
}

type GoUnimplemented struct {
	fname string
	text string
}

func (un *GoUnimplemented) hasVariable(govar GoVar) bool {
	return false
}

func (un *GoUnimplemented) Expr() ast.Expr {
	return ast.NewIdent(un.String())
}

func (un *GoUnimplemented) Init() ast.Stmt {
	return nil
}

func (un *GoUnimplemented) GoName() string {
	return un.String()
}

func (un *GoUnimplemented) IsClassField() bool {
	return false
}

func (un *GoUnimplemented) IsFinal() bool {
	return false
}

func (un *GoUnimplemented) IsStatic() bool {
	return false
}

func (un *GoUnimplemented) Name() string {
	return un.String()
}

func (un *GoUnimplemented) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, un)
}

func (un *GoUnimplemented) Stmts() []ast.Stmt {
	return []ast.Stmt{ &ast.ExprStmt{X: un.Expr()}, }
}

func (un *GoUnimplemented) String() string {
	return fmt.Sprintf("<<unimp_%s[%s]>>", un.fname, un.text)
}

func (un *GoUnimplemented) Type() ast.Expr {
	panic("unimplemented")
}

func (un *GoUnimplemented) VarType() *TypeData {
	return nil //panic("GoUnimplemented.VarType() unimplemented")
}

type GoVarInit struct {
	govar GoVar
	expr GoExpr
	elements []GoExpr
}

func (gvi *GoVarInit) createInitializer(rcvr GoVar) GoStatement {
	sel := NewGoSelector(rcvr, gvi.govar)
	return &GoAssign{govar: sel, tok: token.ASSIGN,
		rhs: []GoExpr{gvi}}
}

func (gvi *GoVarInit) Expr() ast.Expr {
	if gvi.expr != nil {
		return gvi.expr.Expr()
	}

	elements := make([]ast.Expr, len(gvi.elements))
	for i, v := range gvi.elements {
		elements[i] = v.Expr()
	}

	var vartype ast.Expr
	if gvi.govar != nil {
		vartype = gvi.govar.Type()
	}

	return &ast.CompositeLit{Type: vartype, Elts: elements}
}

func (gvi *GoVarInit) hasVariable(govar GoVar) bool {
	if gvi.govar != nil && gvi.govar.Equals(govar) {
		return true
	}

	if gvi.expr != nil && gvi.expr.hasVariable(govar) {
		return true
	}

	for _, e := range gvi.elements {
		if e.hasVariable(govar) {
			return true
		}
	}

	return false
}

func (gvi *GoVarInit) Init() ast.Stmt {
	return nil
}

func (gvi *GoVarInit) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if gvi.govar != nil {
		obj, is_nil := gvi.govar.RunTransform(xform, prog, cls, gvi)
		if !is_nil {
			var err error
			if gvi.govar, err = convertToVar(obj); err != nil {
				panic(err)
			}
		}
	}

	if gvi.expr != nil {
		obj, is_nil := gvi.expr.RunTransform(xform, prog, cls, gvi)
		if !is_nil {
			var err error
			if gvi.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	for i, e := range gvi.elements {
		obj, is_nil := e.RunTransform(xform, prog, cls, gvi)
		if !is_nil {
			var err error
			if gvi.elements[i], err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, gvi)
}

func (gvi *GoVarInit) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoVarInit[")
	if gvi.govar != nil {
		b.WriteString(gvi.govar.String())
	}
	b.WriteString("|")
	if gvi.expr != nil {
		b.WriteString(gvi.expr.String())
	}
	b.WriteString("|")
	if gvi.elements != nil && len(gvi.elements) > 0 {
		for i, v := range gvi.elements {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(v.String())
		}
	}
	b.WriteString("]")
	return b.String()
}

func (gvi *GoVarInit) VarType() *TypeData {
	panic("GoVarInit.VarType() unimplemented")
}

type GoWhile struct {
	expr GoExpr
	stmt GoStatement
	is_do_while bool
}

func (while *GoWhile) hasVariable(govar GoVar) bool {
	if while.expr != nil && while.expr.hasVariable(govar) {
		return true
	}

	if while.stmt != nil && while.stmt.hasVariable(govar) {
		return true
	}

	return false
}

func (while *GoWhile) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	if while.expr != nil {
		obj, is_nil := while.expr.RunTransform(xform, prog, cls, while)
		if !is_nil {
			var err error
			if while.expr, err = convertToExpr(obj); err != nil {
				panic(err)
			}
		}
	}

	if while.stmt != nil {
		obj, is_nil := while.stmt.RunTransform(xform, prog, cls, while)
		if !is_nil {
			var err error
			if while.stmt, err = convertToStmt(obj); err != nil {
				panic(err)
			}
		}
	}

	return xform(parent, prog, cls, while)
}

func (while *GoWhile) Stmts() []ast.Stmt {
	expr := while.expr.Expr()

	var stmt ast.Stmt
	if while.stmt == nil {
		stmt = &ast.EmptyStmt{}
	} else {
		var is_nil bool
		if stmt, is_nil = singleStatement("While expression",
			while.stmt.Stmts()); is_nil {
				stmt = nil
			}
	}

	var body *ast.BlockStmt
	if blk, ok := stmt.(*ast.BlockStmt); ok {
		body = blk
	} else {
		list := make([]ast.Stmt, 1)
		list[0] = stmt
		body = &ast.BlockStmt{List: list}
	}

	if while.is_do_while {
		cond := &ast.UnaryExpr{Op: token.NOT, X: &ast.ParenExpr{X: expr}}

		list := make([]ast.Stmt, 1)
		list[0] = &ast.BranchStmt{Tok: token.BREAK}

		dobrk := &ast.IfStmt{Cond: cond, Body: &ast.BlockStmt{List: list}}

		body.List = append(body.List, dobrk)

		// wipe out condition since we're checking it at the end
		expr = nil
	}

	return []ast.Stmt{ &ast.ForStmt{Cond: expr, Body: body}, }
}

func (while *GoWhile) String() string {
	var bstr string
	if while.is_do_while {
		bstr = "true"
	} else {
		bstr = "false"
	}

	return "GoWhile[" + while.expr.String() + "|" + while.stmt.String() + "|" +
		bstr + "]"
}

type NilMethodOwner struct {
}

func (nm *NilMethodOwner) AddConstant(con *GoConstant) {
	panic("Unimplemented")
}

func (nm *NilMethodOwner) AddMethod(mthd GoMethod) {
	panic("Unimplemented")
}

func (nm *NilMethodOwner) Constants() []ast.Decl {
	panic("Unimplemented")
}

func (nm *NilMethodOwner) FindMethod(name string,
	args *GoMethodArguments) GoMethod {
	panic("Unimplemented")
}

func (nm *NilMethodOwner) IsNil() bool {
	return true
}

func (nm *NilMethodOwner) Name() string {
	panic("Unimplemented")
}

func (nm *NilMethodOwner) Statics() []ast.Decl {
	panic("Unimplemented")
}

func (nm *NilMethodOwner) String() string {
	return "NilMethodOwner[]"
}

func (nm *NilMethodOwner) Super() GoMethodOwner {
	panic("Unimplemented")
}

func (nm *NilMethodOwner) WriteString(out io.Writer, verbose bool) {
	io.WriteString(out, nm.String())
}

var nilMethodOwner = &NilMethodOwner{}
