package parser

import (
	"bytes"
	"fmt"
	"go/ast"
)

type GoVar interface {
	GoObject
	Expr() ast.Expr
	GoName() string
	Equals(govar GoVar) bool
	hasVariable(govar GoVar) bool
	Ident() *ast.Ident
	Init() ast.Stmt
	IsClassField() bool
	IsFinal() bool
	IsStatic() bool
	Name() string
	Receiver() string
	SetGoName(newname string)
	String() string
	Type() ast.Expr
	VarType() *TypeData
}

type GoVarData struct {
	rcvr string
	name string
	goname string
	vartype *TypeData
	class_field bool
	is_static bool
	is_final bool
}

func (gvd *GoVarData) Expr() ast.Expr {
	ident := gvd.Ident()
	if gvd.rcvr == "" || gvd.is_static {
		return ident
	}

	return &ast.SelectorExpr{X: ast.NewIdent(gvd.rcvr), Sel: ident}
}

func (gvd *GoVarData) GoName() string {
	return gvd.goname
}

func (gvd *GoVarData) Equals(govar GoVar) bool {
	if gvd == govar {
		return true
	} else if gvd == nil || govar == nil {
		return false
	}

	return gvd.rcvr == govar.Receiver() && gvd.name == govar.Name()
}

func (gvd *GoVarData) hasVariable(govar GoVar) bool {
	return gvd.Equals(govar)
}

func (gvd *GoVarData) Ident() *ast.Ident {
	return ast.NewIdent(gvd.goname)
}

func (gvd *GoVarData) Init() ast.Stmt {
	return nil
}

func (gvd *GoVarData) IsClassField() bool {
	return gvd.class_field
}

func (gvd *GoVarData) IsFinal() bool {
	return gvd.is_final
}

func (gvd *GoVarData) IsStatic() bool {
	return gvd.is_static
}

func (gvd *GoVarData) Name() string {
	return gvd.name
}

func (gvd *GoVarData) Receiver() string {
	return gvd.rcvr
}

func (gvd *GoVarData) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, gvd)
}

func (gvd *GoVarData) SetGoName(newname string) {
	gvd.goname = newname
}

func (gvd *GoVarData) String() string {
	b := &bytes.Buffer{}
	b.WriteString("GoVarData[")
	if gvd.rcvr != "" {
		b.WriteString(gvd.rcvr)
		b.WriteString(".")
	}
	b.WriteString(gvd.name)
	b.WriteString(fmt.Sprintf("<%v>->", gvd.vartype))
	b.WriteString(gvd.goname)

	if gvd.class_field || gvd.is_final || gvd.is_static {
		b.WriteString(" (")

		need_sep := false

		if gvd.class_field {
			if need_sep {
				b.WriteString("|")
			}
			b.WriteString("clsfld")
			need_sep = true
		}

		if gvd.is_final {
			if need_sep {
				b.WriteString("|")
			}
			b.WriteString("final")
			need_sep = true
		}

		if gvd.is_static {
			if need_sep {
				b.WriteString("|")
			}
			b.WriteString("static")
			need_sep = true
		}

		b.WriteString(")")
	}
	b.WriteString("]")

	return b.String()
}

func (gvd *GoVarData) Type() ast.Expr {
	if gvd.vartype == nil {
		return nil
	}

	return gvd.vartype.Expr()
}

func (gvd *GoVarData) VarType() *TypeData {
	return gvd.vartype
}

type GoClassAttribute struct {
	govar GoVar
	suffix string
}

func (gvr *GoClassAttribute) Equals(govar GoVar) bool {
	return gvr.govar.Equals(govar)
}

func (gvr *GoClassAttribute) Expr() ast.Expr {
	return &ast.SelectorExpr{X: gvr.govar.Expr(),
		Sel: ast.NewIdent(gvr.suffix)}
}

func (gvr *GoClassAttribute) hasVariable(govar GoVar) bool {
	return gvr.Equals(govar)
}

func (gvr *GoClassAttribute) GoName() string {
	return gvr.govar.GoName() + "." + gvr.suffix
}

func (gvr *GoClassAttribute) Ident() *ast.Ident {
	return gvr.govar.Ident()
}

func (gvr *GoClassAttribute) Init() ast.Stmt {
	return nil
}

func (gvr *GoClassAttribute) IsClassField() bool {
	return gvr.govar.IsClassField()
}

func (gvr *GoClassAttribute) IsFinal() bool {
	return gvr.govar.IsFinal()
}

func (gvr *GoClassAttribute) IsStatic() bool {
	return gvr.govar.IsStatic()
}

func (gvr *GoClassAttribute) Name() string {
	return gvr.govar.Name() + "." + gvr.suffix
}

func (gvr *GoClassAttribute) Receiver() string {
	return gvr.govar.Receiver()
}

func (gvr *GoClassAttribute) RunTransform(xform TransformFunc, prog *GoProgram, cls GoClass, parent GoObject) (GoObject, bool) {
	return xform(parent, prog, cls, gvr)
}

func (gvr *GoClassAttribute) SetGoName(newname string) {
	gvr.govar.SetGoName(newname)
}

func (gvr *GoClassAttribute) String() string {
	return fmt.Sprintf("GoClassAttribute[%s|%s]", gvr.govar.String(), gvr.suffix)
}

func (gvr *GoClassAttribute) Suffix() string {
	return gvr.suffix
}

func (gvr *GoClassAttribute) Type() ast.Expr {
	return gvr.govar.Type()
}

func (gvr *GoClassAttribute) VarType() *TypeData {
	return gvr.govar.VarType()
}
