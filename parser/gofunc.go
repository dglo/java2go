package parser

import (
	"go/ast"
	"go/token"

	"java2go/grammar"
)

type GoFunc interface {
	Create() ast.Stmt
	//Expr(gs *GoState) *ast.FuncDecl
	GoName() string
	Receiver() string
}

type GoFuncData struct {
	class     string
	name      string
	goname    string
	func_type methodType
	rcvr      string
	params    []*grammar.JFormalParameter
}

func (gf *GoFuncData) Create() ast.Stmt {
	if gf.func_type != mt_constructor {
		return nil
	}

	// left side of first statement in NewFoo()
	//   (assign new struct to the receiver)
	lhs := make([]ast.Expr, 1)
	lhs[0] = ast.NewIdent(gf.rcvr)

	// right side of first statement in NewFoo()
	//   (create new struct)
	rhs := make([]ast.Expr, 1)
	rhs[0] = &ast.UnaryExpr{Op: token.AND,
		X: &ast.CompositeLit{Type: ast.NewIdent(gf.class)}}

	// build creation statement
	return &ast.AssignStmt{Lhs: lhs, Tok: token.DEFINE, Rhs: rhs}
}

/*
func (gf *GoFuncData) Expr(gs *GoState) *ast.FuncDecl {
	mtype := &ast.FuncType{Params: gf.paramList(gs), Results: gf.results()}

	return &ast.FuncDecl{Name: ast.NewIdent(gf.goname),
		Recv: gf.recv(), Type: mtype}
}
*/

func (gf *GoFuncData) GoName() string {
	return gf.goname
}

/*
func (gf *GoFuncData) paramList(gs *GoState) *ast.FieldList {
	if gf.func_type == mt_test {
		if gf.params != nil && len(gf.params) > 0 {
			fmt.Fprintf(os.Stderr,
				"//ERR// Test method %s.%s should not have %d params\n",
				gf.class, gf.name, len(gf.params))
		}

		selexp := &ast.SelectorExpr{X: ast.NewIdent("testing"),
			Sel: ast.NewIdent("T")}

		pfld := makeField("t", &ast.StarExpr{X: selexp})

		paramList := make([]*ast.Field, 1)
		paramList[0] = pfld

		return &ast.FieldList{List: paramList}
	}

	if gf.params != nil && len(gf.params) > 0 {
		flist := make([]*ast.Field, len(gf.params))

		for i, fp := range gf.params {
			if fp.typespec != nil {
				if fp.dims != 0 {
					fmt.Fprintf(os.Stderr,
						"//ERR// Ignoring %s dims=%d for %s.%s\n",
						fp.name, fp.dims, gf.class, gf.name)
				} else if fp.dotdotdot {
					fmt.Fprintf(os.Stderr,
						"//ERR// Ignoring %s dotdotdot=true for %s.%s\n",
						fp.name, gf.class, gf.name)
				}

				ftype := gs.createTypeData(fp.typespec.name,
					fp.typespec.type_args, fp.typespec.dims)

				flist[i] = makeField(fp.name, ftype.Expr())
			}
		}

		return &ast.FieldList{List: flist}
	}

	return nil
}
*/

func (gf *GoFuncData) Receiver() string {
	return gf.rcvr
}

func (gf *GoFuncData) recv() *ast.FieldList {
	if gf.func_type != mt_method {
		// non-methods don't have a receiver
		return nil
	}

	rlist := make([]*ast.Field, 1)
	rlist[0] = makeField(gf.rcvr, &ast.StarExpr{X: ast.NewIdent(gf.class)})
	return &ast.FieldList{List: rlist}
}

func (gf *GoFuncData) results() *ast.FieldList {
	if gf.func_type != mt_constructor {
		return nil
	}

	// return the receiver
	rlist := make([]*ast.Field, 1)
	rlist[0] = makeField(gf.rcvr, &ast.StarExpr{X: ast.NewIdent(gf.class)})
	return &ast.FieldList{List: rlist}
}

type GoFuncRef struct {
	gofunc GoFunc
	suffix *grammar.JTypeName
}

func (gfr *GoFuncRef) Create() ast.Stmt {
	return gfr.Create()
}

/*
func (gfr *GoFuncRef) Expr(gs *GoState) *ast.FuncDecl {
	return gfr.Expr(gs)
}
*/

func (gfr *GoFuncRef) GoName() string {
	return gfr.GoName() + "." + gfr.suffix.String()
}

func (gfr *GoFuncRef) Receiver() string {
	return gfr.Receiver()
}
