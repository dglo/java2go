package dumper

import (
	"fmt"
	"go/ast"
)

func buildIndent(indentNum int) string {
	var indent string
	for i := 0; i < indentNum; i++ {
		indent += "  "
	}
	return indent
}

func Dump(srcname string, f *ast.File) {
	fmt.Printf("%s returned <%T> %v\n---- Scope ----\n", srcname, f, f)
	dumpScope(f.Scope, 0)
	if f.Imports != nil && len(f.Imports) > 0 {
		fmt.Printf("--- Imports ---\n")
		for _, spec := range f.Imports {
			dumpImportSpec(spec, 0)
		}
	}
	if f.Decls != nil && len(f.Decls) > 0 {
		fmt.Printf("---- Decls ----\n")
		for _, decl := range f.Decls {
			dumpDecl(decl, 0)
		}
	}
	fmt.Printf("---------------\n")
}

func dumpArrayType(atyp *ast.ArrayType, indentNum int) {
	indent := buildIndent(indentNum)
	fmt.Printf("%s-- ArrayType\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if atyp.Len != nil {
		fmt.Printf("%s-- Len\n", indent1)
		dumpExpr(atyp.Len, indentNum + 2)
	}

	fmt.Printf("%s-- Elt\n", indent1)
	dumpExpr(atyp.Elt, indentNum + 2)
}

func dumpAssignStmt(asgn *ast.AssignStmt, indentNum int) {
	fmt.Printf("%s-- AssignStmt\n", buildIndent(indentNum))

	indent1 := buildIndent(indentNum + 1)

	if asgn.Lhs == nil || len(asgn.Lhs) == 0 {
		fmt.Printf("%s<no lhs>\n", indent1)
	} else {
		fmt.Printf("%s-- Lhs\n", indent1)
		for _, expr := range asgn.Lhs {
			dumpExpr(expr, indentNum + 2)
		}
	}

	fmt.Printf("%s-- Token -> %s\n", indent1, asgn.Tok.String())

	if asgn.Rhs == nil || len(asgn.Rhs) == 0 {
		fmt.Printf("%s<no rhs>\n", indent1)
	} else {
		fmt.Printf("%s-- Rhs\n", indent1)
		for _, expr := range asgn.Rhs {
			dumpExpr(expr, indentNum + 2)
		}
	}
}

func dumpBinaryExpr(bin *ast.BinaryExpr, indentNum int) {
	indent := buildIndent(indentNum)

	if bin == nil {
		fmt.Printf("%sBinaryExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- BinaryExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if bin.X != nil {
		fmt.Printf("%s-- X\n", indent1)
		dumpExpr(bin.X, indentNum + 2)
	}

	fmt.Printf("%sOp -> \"%s\"\n", indent1, bin.Op.String())

	if bin.Y != nil {
		fmt.Printf("%s-- Y\n", indent1)
		dumpExpr(bin.Y, indentNum + 2)
	}
}

func dumpBlockStmt(blk *ast.BlockStmt, indentNum int) {
	if blk.List == nil || len(blk.List) == 0 {
		fmt.Printf("%s<empty BlockStmt>\n", buildIndent(indentNum))
		return
	}

	for _, stmt := range blk.List {
		dumpStmt(stmt, indentNum)
	}
}

func dumpBranchStmt(bstmt *ast.BranchStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if bstmt == nil {
		fmt.Printf("%sBranchStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sBranchStmt -> %s\n", indent, bstmt.Tok.String())

	indent1 := buildIndent(indentNum + 1)

	if bstmt.Label == nil {
		fmt.Printf("%s<empty label>\n", indent1)
	} else {
		fmt.Printf("%sLabel -> %s\n", indent1, bstmt.Label.Name)
	}
}

func dumpCallExpr(call *ast.CallExpr, indentNum int) {
	fmt.Printf("%s-- CallExpr\n", buildIndent(indentNum))

	dumpExpr(call.Fun, indentNum + 1)

	indent1 := buildIndent(indentNum + 1)
	if call.Args == nil || len(call.Args) == 0 {
		fmt.Printf("%s<no args>\n", indent1)
	} else {
		fmt.Printf("%s-- Args\n", indent1)
		for _, expr := range call.Args {
			dumpExpr(expr, indentNum + 2)
		}
	}

}

func dumpCaseClause(cc *ast.CaseClause, indentNum int) {
	indent := buildIndent(indentNum)

	if cc == nil {
		fmt.Printf("%sCaseClause -> <nil>\n", indent)
		return
	}

	indent1 := buildIndent(indentNum + 1)

	fmt.Printf("%sCaseClause ->\n", indent)

	if cc.List == nil || len(cc.List) == 0 {
		fmt.Printf("%s<empty List>\n", indent1)
	} else {
		fmt.Printf("%s-- List\n", indent1)
		for _, expr := range cc.List {
			dumpExpr(expr, indentNum + 2)
		}
	}

	if cc.Body == nil || len(cc.Body) == 0 {
		fmt.Printf("%s<empty Body>\n", indent1)
	} else {
		fmt.Printf("%s-- Body\n", indent1)
		for _, stmt := range cc.Body {
			dumpStmt(stmt, indentNum + 2)
		}
	}
}

func dumpChanType(ch *ast.ChanType, indentNum int) {
	indent := buildIndent(indentNum)

	if ch == nil {
		fmt.Printf("%sChanType -> <nil>\n", indent)
		return
	}

	var dstr string
	switch ch.Dir {
	case ast.SEND:
		dstr = "SEND"
	case ast.RECV:
		dstr = "RECV"
	case ast.SEND|ast.RECV:
		dstr = "SEND|RECV"
	default:
		dstr = fmt.Sprintf("??ChanDir#%d??", ch.Dir)
	}

	fmt.Printf("%s-- ChanType -> Dir %s\n", indent, dstr)

	indent1 := buildIndent(indentNum + 1)

	if ch.Value != nil {
		fmt.Printf("%s-- Value\n", indent1)
		dumpExpr(ch.Value, indentNum + 2)
	}
}

func dumpCompositeLit(clit *ast.CompositeLit, indentNum int) {
	indent := buildIndent(indentNum)

	if clit == nil {
		fmt.Printf("%sCompositeLit -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sCompositeLit\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if clit.Type != nil {
		fmt.Printf("%s-- Type\n", indent1)
		dumpExpr(clit.Type, indentNum + 2)
	}

	if clit.Elts == nil || len(clit.Elts) == 0 {
		fmt.Printf("%s<no elts>\n", indent1)
	} else {
		fmt.Printf("%s-- Elts\n", indent1)
		for _, expr := range clit.Elts {
			dumpExpr(expr, indentNum + 2)
		}
	}
}

func dumpDecl(decl ast.Decl, indentNum int) {
	if decl == nil {
		fmt.Printf("%sDecl -> <nil>\n", buildIndent(indentNum))
		return
	}

	switch d := decl.(type) {
	case *ast.GenDecl:
		dumpGenDecl(d, indentNum)
	case *ast.FuncDecl:
		dumpFuncDecl(d, indentNum)
	default:
		fmt.Printf("%s??Unknown Decl %T??\n", buildIndent(indentNum), decl)
	}
}

func dumpDeferStmt(dstmt *ast.DeferStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if dstmt == nil {
		fmt.Printf("%sDeferStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sDeferStmt ->\n", indent)
	dumpCallExpr(dstmt.Call, indentNum + 1)
}

func dumpEllipsis(elp *ast.Ellipsis, indentNum int) {
	indent := buildIndent(indentNum)

	if elp == nil {
		fmt.Printf("%sEllipsis -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sEllipsis\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if elp.Elt != nil {
		fmt.Printf("%s-- Elt\n", indent1)
		dumpExpr(elp.Elt, indentNum + 2)
	}
}

func dumpExpr(expr ast.Expr, indentNum int) {
	if expr == nil {
		fmt.Printf("%sExpr -> <nil>\n", buildIndent(indentNum))
		return
	}

	switch e := expr.(type) {
	case *ast.ArrayType:
		dumpArrayType(e, indentNum)
	case *ast.BasicLit:
		fmt.Printf("%sBasicLit -> %s (tok %s)\n", buildIndent(indentNum),
			e.Value, e.Kind)
	case *ast.BinaryExpr:
		dumpBinaryExpr(e, indentNum)
	case *ast.CallExpr:
		dumpCallExpr(e, indentNum)
	case *ast.ChanType:
		dumpChanType(e, indentNum)
	case *ast.CompositeLit:
		dumpCompositeLit(e, indentNum)
	case *ast.Ellipsis:
		dumpEllipsis(e, indentNum)
	case *ast.FuncLit:
		dumpFuncLit(e, indentNum)
	case *ast.FuncType:
		dumpFuncType(e, indentNum)
	case *ast.Ident:
		fmt.Printf("%sIdent -> %s\n", buildIndent(indentNum), e.Name)
	case *ast.IndexExpr:
		dumpIndexExpr(e, indentNum)
	case *ast.InterfaceType:
		dumpInterfaceType(e, indentNum)
	case *ast.KeyValueExpr:
		dumpKeyValueExpr(e, indentNum)
	case *ast.MapType:
		fmt.Printf("%sMapType\n", buildIndent(indentNum))
		dumpExpr(e.Key, indentNum + 1)
		dumpExpr(e.Value, indentNum + 1)
	case *ast.ParenExpr:
		dumpParenExpr(e, indentNum)
	case *ast.SelectorExpr:
		var sstr string
		if e.Sel == nil {
			sstr = "<nil>"
		} else {
			sstr = e.Sel.Name
		}

		fmt.Printf("%sSelectorExpr -> %s\n", buildIndent(indentNum), sstr)

		dumpExpr(e.X, indentNum + 1)
	case *ast.SliceExpr:
		dumpSliceExpr(e, indentNum)
	case *ast.StarExpr:
		fmt.Printf("%sStarExpr -> *\n", buildIndent(indentNum))
		dumpExpr(e.X, indentNum + 1)
	case *ast.StructType:
		dumpStructType(e, indentNum + 1)
	case *ast.TypeAssertExpr:
		dumpTypeAssertExpr(e, indentNum + 1)
	case *ast.UnaryExpr:
		dumpUnaryExpr(e, indentNum)
	default:
		fmt.Printf("%s??Expr -> %T\n", buildIndent(indentNum), expr)
	}
}

func dumpField(fld *ast.Field, indentNum int) {
	if fld == nil {
		fmt.Printf("%sField -> <nil>\n", buildIndent(indentNum))
		return
	}

	var names string
	if fld.Names == nil || len(fld.Names) == 0 {
		names = ""
	} else {
		for _, name := range fld.Names {
			if names != "" {
				names += ", "
			}

			if name == nil {
				names += "<nil>"
			} else {
				names += name.Name
			}
		}
	}

	var tag string
	if fld.Tag != nil {
		tag = " tag " + fld.Tag.Value
	}

	fmt.Printf("%sField -> %s%s\n", buildIndent(indentNum), names, tag)
	dumpExpr(fld.Type, indentNum + 1)
}

func dumpFieldList(list *ast.FieldList, indentNum int) {
	if list == nil {
		fmt.Printf("%sFieldList -> <nil>\n", buildIndent(indentNum))
		return
	}

	if list.List == nil || len(list.List) == 0 {
		fmt.Printf("%s<empty FieldList>\n", buildIndent(indentNum))
		return
	}

	for _, fld := range list.List {
		dumpField(fld, indentNum)
	}
}

func dumpForStmt(fstmt *ast.ForStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if fstmt == nil {
		fmt.Printf("%sForStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- ForStmt\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if fstmt.Init != nil {
		fmt.Printf("%s-- Init\n", indent1)
		dumpStmt(fstmt.Init, indentNum + 2)
	}
	if fstmt.Cond != nil {
		fmt.Printf("%s-- Cond\n", indent1)
		dumpExpr(fstmt.Cond, indentNum + 2)
	}
	if fstmt.Post != nil {
		fmt.Printf("%s-- Post\n", indent1)
		dumpStmt(fstmt.Post, indentNum + 2)
	}
	if fstmt.Body != nil {
		fmt.Printf("%s-- Body\n", indent1)
		dumpBlockStmt(fstmt.Body, indentNum + 2)
	}
}

func dumpFuncDecl(fun *ast.FuncDecl, indentNum int) {
	indent := buildIndent(indentNum)

	if fun == nil {
		fmt.Printf("%sFuncDecl -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sFuncDecl %s\n", indent, fun.Name.Name)

	if fun.Recv != nil {
		fmt.Printf("%s-- Recv\n", indent)
		dumpFieldList(fun.Recv, indentNum + 1)
	}

	if fun.Type != nil {
		dumpFuncType(fun.Type, indentNum + 1)
	}

	if fun.Body != nil {
		fmt.Printf("%s-- Body\n", indent)
		dumpBlockStmt(fun.Body, indentNum + 1)
	}
}

func dumpFuncType(ftyp *ast.FuncType, indentNum int) {
	indent := buildIndent(indentNum)

	if ftyp == nil || (ftyp.Params == nil && ftyp.Results == nil) {
		fmt.Printf("%s-- <empty type>\n", indent)
	} else {
		if ftyp.Params != nil {
			fmt.Printf("%s-- FuncType.Params\n", indent)
			dumpFieldList(ftyp.Params, indentNum + 1)
		}
		if ftyp.Results != nil {
			fmt.Printf("%s-- FuncType.Results\n", indent)
			dumpFieldList(ftyp.Results, indentNum + 1)
		}
	}
}

func dumpFuncLit(flit *ast.FuncLit, indentNum int) {
	indent := buildIndent(indentNum)

	if flit == nil {
		fmt.Printf("%sFuncLit -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sFuncLit\n", indent)

	if flit.Type != nil {
		dumpFuncType(flit.Type, indentNum + 1)
	}

	if flit.Body != nil {
		fmt.Printf("%s-- Body\n", indent)
		dumpBlockStmt(flit.Body, indentNum + 1)
	}
}

func dumpGenDecl(gen *ast.GenDecl, indentNum int) {
	fmt.Printf("%sGenDecl -> \"%s\"\n", buildIndent(indentNum),
		gen.Tok.String())
	if gen.Specs != nil {
		for _, spec := range gen.Specs {
			dumpSpec(spec, indentNum + 1)
		}
	}
}

func dumpGoStmt(gostmt *ast.GoStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if gostmt == nil {
		fmt.Printf("%sGoStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sGoStmt ->\n", indent)
	dumpCallExpr(gostmt.Call, indentNum + 1)
}

func dumpIfStmt(ifstmt *ast.IfStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if ifstmt == nil {
		fmt.Printf("%sIfStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- IfStmt\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if ifstmt.Init != nil {
		fmt.Printf("%s-- Init\n", indent1)
		dumpStmt(ifstmt.Init, indentNum + 2)
	}
	if ifstmt.Cond != nil {
		fmt.Printf("%s-- Cond\n", indent1)
		dumpExpr(ifstmt.Cond, indentNum + 2)
	}
	if ifstmt.Body != nil {
		fmt.Printf("%s-- Body\n", indent1)
		dumpBlockStmt(ifstmt.Body, indentNum + 2)
	}
	if ifstmt.Else != nil {
		fmt.Printf("%s-- Else\n", indent1)
		dumpStmt(ifstmt.Else, indentNum + 2)
	}
}

func dumpImportSpec(spec *ast.ImportSpec, indentNum int) {
	if spec == nil {
		fmt.Printf("%sImportSpec -> <nil>\n", buildIndent(indentNum))
		return
	}

	var nstr string
	if spec.Name != nil {
		nstr = fmt.Sprintf(" name %s", spec.Name.Name)
	}

	var pstr string
	if spec.Path != nil {
		pstr = fmt.Sprintf(" path %s", spec.Path.Value)
	}

	fmt.Printf("%sImportSpec ->%s%s\n", buildIndent(indentNum), nstr, pstr)
}

func dumpIncDecStmt(ids *ast.IncDecStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if ids == nil {
		fmt.Printf("%sIncDecStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sIncDecStmt -> %s\n", indent, ids.Tok.String())

	if ids.X != nil {
		fmt.Printf("%s-- X\n", buildIndent(indentNum + 1))
		dumpExpr(ids.X, indentNum + 2)
	}
}

func dumpIndexExpr(idx *ast.IndexExpr, indentNum int) {
	indent := buildIndent(indentNum)

	if idx == nil {
		fmt.Printf("%sIndexExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- IndexExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if idx.X != nil {
		fmt.Printf("%s-- X\n", indent1)
		dumpExpr(idx.X, indentNum + 2)
	}

	if idx.Index != nil {
		fmt.Printf("%s-- Index\n", indent1)
		dumpExpr(idx.Index, indentNum + 2)
	}
}

func dumpInterfaceType(ityp *ast.InterfaceType, indentNum int) {
	indent := buildIndent(indentNum)

	if ityp == nil {
		fmt.Printf("%sInterfaceType -> <nil>\n", indent)
		return
	}

	var istr string
	if ityp.Incomplete {
		istr = " Incomplete"
	}

	fmt.Printf("%sInterfaceType ->%s\n", indent, istr)
	dumpFieldList(ityp.Methods, indentNum + 1)
}

func dumpKeyValueExpr(kv *ast.KeyValueExpr, indentNum int) {
	indent := buildIndent(indentNum)

	if kv == nil {
		fmt.Printf("%sKeyValueExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- KeyValueExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if kv.Key != nil {
		fmt.Printf("%s-- Key\n", indent1)
		dumpExpr(kv.Key, indentNum + 2)
	}

	if kv.Value != nil {
		fmt.Printf("%s-- Value\n", indent1)
		dumpExpr(kv.Value, indentNum + 2)
	}
}

func dumpLabeledStmt(lex *ast.LabeledStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if lex == nil {
		fmt.Printf("%sLabeledExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- LabeledExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if lex.Label != nil {
		fmt.Printf("%s-- Label -> %v\n", indent1, lex.Label)
	}
	if lex.Stmt != nil {
		fmt.Printf("%s-- Stmt\n", indent1)
		dumpStmt(lex.Stmt, indentNum + 2)
	}
}

func dumpParenExpr(prn *ast.ParenExpr, indentNum int) {
	indent := buildIndent(indentNum)

	if prn == nil {
		fmt.Printf("%sParenExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- ParenExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if prn.X != nil {
		fmt.Printf("%s-- X\n", indent1)
		dumpExpr(prn.X, indentNum + 2)
	}
}

func dumpRangeStmt(rstmt *ast.RangeStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if rstmt == nil {
		fmt.Printf("%sRangeStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- RangeStmt -> %s\n", indent, rstmt.Tok.String())

	indent1 := buildIndent(indentNum + 1)

	if rstmt.Key != nil {
		fmt.Printf("%s-- Key\n", indent1)
		dumpExpr(rstmt.Key, indentNum + 2)
	}
	if rstmt.Value != nil {
		fmt.Printf("%s-- Value\n", indent1)
		dumpExpr(rstmt.Value, indentNum + 2)
	}
	if rstmt.X != nil {
		fmt.Printf("%s-- X\n", indent1)
		dumpExpr(rstmt.X, indentNum + 2)
	}
	if rstmt.Body != nil {
		fmt.Printf("%s-- Body\n", indent1)
		dumpBlockStmt(rstmt.Body, indentNum + 2)
	}
}

func dumpReturnStmt(rtn *ast.ReturnStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if rtn == nil {
		fmt.Printf("%sReturnStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sReturnStmt ->\n", indent)
	for _, expr := range rtn.Results {
		dumpExpr(expr, indentNum + 1)
	}
}

func dumpScope(scope *ast.Scope, indentNum int) {
	indent := buildIndent(indentNum)

	if scope == nil {
		fmt.Printf("%s<nil>\n", indent)
		return
	}

	if scope.Objects == nil {
		fmt.Printf("%s<no Objects>\n", indent)
	} else {
		for k, v := range scope.Objects {
			fmt.Printf("%s%s: <%T> %v\n", indent, k, v, v)
		}
	}

	dumpScope(scope.Outer, indentNum + 1)
}

func dumpSendStmt(ss *ast.SendStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if ss == nil {
		fmt.Printf("%sSendStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sSendStmt\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if ss.Chan != nil {
		fmt.Printf("%s-- Chan\n", indent1)
		dumpExpr(ss.Chan, indentNum + 2)
	}
	if ss.Value != nil {
		fmt.Printf("%s-- Value\n", indent1)
		dumpExpr(ss.Value, indentNum + 2)
	}
}

func dumpSliceExpr(slc *ast.SliceExpr, indentNum int) {
	indent := buildIndent(indentNum)

	if slc == nil {
		fmt.Printf("%sSliceExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- SliceExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if slc.X != nil {
		fmt.Printf("%s-- X\n", indent1)
		dumpExpr(slc.X, indentNum + 2)
	}

	if slc.Low != nil {
		fmt.Printf("%s-- Low\n", indent1)
		dumpExpr(slc.Low, indentNum + 2)
	}

	if slc.High != nil {
		fmt.Printf("%s-- High\n", indent1)
		dumpExpr(slc.High, indentNum + 2)
	}

	if slc.Max != nil {
		fmt.Printf("%s-- Max\n", indent1)
		dumpExpr(slc.Max, indentNum + 2)
	}
}

func dumpSpec(spec ast.Spec, indentNum int) {
	if spec == nil {
		fmt.Printf("%sSpec -> <nil>\n", buildIndent(indentNum))
		return
	}

	switch s := spec.(type) {
	case *ast.ImportSpec:
		dumpImportSpec(s, indentNum)
	case *ast.TypeSpec:
		dumpTypeSpec(s, indentNum)
	case *ast.ValueSpec:
		dumpValueSpec(s, indentNum)
	default:
		fmt.Printf("%s??Spec -> %T\n", buildIndent(indentNum), spec)
	}
}

func dumpStmt(stmt ast.Stmt, indentNum int) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		dumpAssignStmt(s, indentNum)
	case *ast.BlockStmt:
		dumpBlockStmt(s, indentNum)
	case *ast.BranchStmt:
		dumpBranchStmt(s, indentNum)
	case *ast.CaseClause:
		dumpCaseClause(s, indentNum)
	// case *ast.CommClause:
	case *ast.DeclStmt:
		fmt.Printf("%s-- DeclStmt\n", buildIndent(indentNum))
		dumpDecl(s.Decl, indentNum + 1)
	case *ast.DeferStmt:
		dumpDeferStmt(s, indentNum)
	// case *ast.EmptyStmt:
	case *ast.ExprStmt:
		fmt.Printf("%s-- ExprStmt\n", buildIndent(indentNum))
		dumpExpr(s.X, indentNum + 1)
	case *ast.ForStmt:
		dumpForStmt(s, indentNum)
	case *ast.GoStmt:
		dumpGoStmt(s, indentNum)
	case *ast.IfStmt:
		dumpIfStmt(s, indentNum)
	case *ast.IncDecStmt:
		dumpIncDecStmt(s, indentNum)
	case *ast.LabeledStmt:
		dumpLabeledStmt(s, indentNum)
	case *ast.RangeStmt:
		dumpRangeStmt(s, indentNum)
	case *ast.ReturnStmt:
		dumpReturnStmt(s, indentNum)
	// case *ast.SelectStmt:
	case *ast.SendStmt:
		dumpSendStmt(s, indentNum)
	case *ast.SwitchStmt:
		dumpSwitchStmt(s, indentNum)
	case *ast.TypeSwitchStmt:
		dumpTypeSwitchStmt(s, indentNum)
	default:
		fmt.Printf("%s??Stmt -> %T\n", buildIndent(indentNum), stmt)
	}
}

func dumpStructType(styp *ast.StructType, indentNum int) {
	indent := buildIndent(indentNum)

	if styp == nil {
		fmt.Printf("%sStructType -> <nil>\n", indent)
		return
	}

	var istr string
	if styp.Incomplete {
		istr = " Incomplete"
	}

	fmt.Printf("%sStructType ->%s\n", indent, istr)
	dumpFieldList(styp.Fields, indentNum + 1)
}

func dumpSwitchStmt(ss *ast.SwitchStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if ss == nil {
		fmt.Printf("%sSwitchStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sSwitchStmt\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if ss.Init != nil {
		fmt.Printf("%s-- Init\n", indent1)
		dumpStmt(ss.Init, indentNum + 2)
	}
	if ss.Tag != nil {
		fmt.Printf("%s-- Tag\n", indent1)
		dumpExpr(ss.Tag, indentNum + 2)
	}
	if ss.Body != nil {
		fmt.Printf("%s-- Body\n", indent1)
		dumpBlockStmt(ss.Body, indentNum + 2)
	}
}

func dumpTypeAssertExpr(tae *ast.TypeAssertExpr, indentNum int) {
	indent := buildIndent(indentNum)

	if tae == nil {
		fmt.Printf("%sTypeAssertExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- TypeAssertExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if tae.X != nil {
		fmt.Printf("%s-- X\n", indent1)
		dumpExpr(tae.X, indentNum + 2)
	}

	if tae.Type != nil {
		fmt.Printf("%s-- Type\n", indent1)
		dumpExpr(tae.Type, indentNum + 2)
	}
}

func dumpTypeSpec(spec *ast.TypeSpec, indentNum int) {
	indent := buildIndent(indentNum)

	if spec == nil {
		fmt.Printf("%sTypeSpec -> <nil>\n", indent)
		return
	}

	var name string
	if spec.Name == nil {
		name = "<nil>"
	} else {
		name = spec.Name.Name
	}

	fmt.Printf("%sTypeSpec -> name %s\n", indent, name)

	if spec.Type != nil {
		fmt.Printf("%s-- Type\n", indent)
		dumpExpr(spec.Type, indentNum + 1)
	}
}

func dumpTypeSwitchStmt(tss *ast.TypeSwitchStmt, indentNum int) {
	indent := buildIndent(indentNum)

	if tss == nil {
		fmt.Printf("%sTypeSwitchStmt -> <nil>\n", indent)
		return
	}

	fmt.Printf("%sTypeSwitchStmt\n", indent)

	indent1 := buildIndent(indentNum + 1)

	if tss.Init != nil {
		fmt.Printf("%s-- Init\n", indent1)
		dumpStmt(tss.Init, indentNum + 2)
	}
	if tss.Assign != nil {
		fmt.Printf("%s-- Assign\n", indent1)
		dumpStmt(tss.Assign, indentNum + 2)
	}
	if tss.Body != nil {
		fmt.Printf("%s-- Body\n", indent1)
		dumpBlockStmt(tss.Body, indentNum + 2)
	}
}

func dumpUnaryExpr(unary *ast.UnaryExpr, indentNum int) {
	indent := buildIndent(indentNum)

	if unary == nil {
		fmt.Printf("%sUnaryExpr -> <nil>\n", indent)
		return
	}

	fmt.Printf("%s-- UnaryExpr\n", indent)

	indent1 := buildIndent(indentNum + 1)

	fmt.Printf("%sOp -> \"%s\"\n", indent1, unary.Op.String())

	if unary.X != nil {
		fmt.Printf("%s-- X\n", indent1)
		dumpExpr(unary.X, indentNum + 2)
	}
}

func dumpValueSpec(spec *ast.ValueSpec, indentNum int) {
	if spec == nil {
		fmt.Printf("%sValueSpec -> <nil>\n", buildIndent(indentNum))
		return
	}

	var names string
	if spec.Names == nil || len(spec.Names) == 0 {
		names = "<anonymous>"
	} else {
		for _, name := range spec.Names {
			if names != "" {
				names += ", "
			}

			if name == nil {
				names += "<nil>"
			} else {
				names += name.Name
			}
		}
	}

	indent := buildIndent(indentNum)

	fmt.Printf("%sValueSpec -> names %s\n", indent, names)

	if spec.Type != nil {
		fmt.Printf("%s-- Type\n", indent)
		dumpExpr(spec.Type, indentNum + 1)
	}

	if spec.Values != nil && len(spec.Values) > 0 {
		fmt.Printf("%s-- Values\n", indent)
		for _, val := range spec.Values {
			dumpExpr(val, indentNum + 1)
		}
	}
}
