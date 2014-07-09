package parser

import (
	"fmt"
	"go/token"
	"log"
	//"os"

	"github.com/dglo/java2go/grammar"
)

func findMethod(owner GoMethodOwner, class GoMethodOwner, name string,
	arglist *GoMethodArguments, verbose bool) GoMethod {
	var mthd GoMethod
	if class != nil && !class.IsNil() {
		mthd = class.FindMethod(name, arglist)
	}

	if mthd == nil && owner != nil && !owner.IsNil() {
		mthd = owner.FindMethod(name, arglist)
	}

	if mthd == nil {
		mthd = NewGoMethodReference(class, name, arglist, verbose)
		if owner != nil && !owner.IsNil() {
			owner.AddMethod(mthd)
		} else if class != nil && !class.IsNil() {
			class.AddMethod(mthd)
		} else {
			panic("Both class and owner are nil")
		}
	}

	return mthd
}

func analyzeAllocationExpr(gs *GoState, owner GoMethodOwner,
	alloc *grammar.JClassAllocationExpr) *GoClassAlloc {
	if alloc.Name.IsPrimitive() {
		panic(fmt.Sprintf("Class allocation should not use primitive \"%s\"",
			alloc.Name.String()))
	}

	var type_args []*TypeData
	if alloc.TypeArgs != nil && len(alloc.TypeArgs) > 0 {
		type_args = make([]*TypeData, len(alloc.TypeArgs))
		for i, arg := range alloc.TypeArgs {
			if arg.Ts_type != grammar.TS_NONE {
				log.Printf("//ERR// Ignoring allexpr ts#%d type %v\n",
					i, arg.Ts_type)
			}

			if arg.TypeSpec == nil {
				type_args[i] = genericObject
			} else {
				type_args[i] = gs.Program().createTypeData(arg.TypeSpec.Name,
					arg.TypeSpec.TypeArgs, 0)
			}
		}
	}

	var args []GoExpr
	if alloc.Arglist != nil && len(alloc.Arglist) > 0 {
		args = make([]GoExpr, len(alloc.Arglist))
		for i, arg := range alloc.Arglist {
			args[i] = analyzeExpr(gs, owner, arg)
		}
	}

	alloc_name := alloc.Name.String()

	var cref GoClass

	var body []GoStatement
	for _, b := range alloc.Body {
		switch j := b.(type) {
		case *grammar.JClassBody:
			gs2 := NewGoState(gs)

			c2 := NewGoClassDefinition(gs.Program(), owner, alloc_name)
			analyzeClassBody(gs2, c2, j)

			var anon_name string
			for i := 0; ; i++ {
				anon_name = fmt.Sprintf("Anonymous_%s_%d", alloc_name, i)
				if gs.findClass(owner, anon_name) == nil {
					break
				}
			}
			c2.name = anon_name

			gs.addClass(c2)
			cref = c2
		case *grammar.JBlock:
			stmt := analyzeBlock(gs, owner, j)
			if stmt != nil {
				body = append(body, stmt)
			}
		case *grammar.JEmpty:
			; // do nothing
		default:
			grammar.ReportCastError("JClassAllocationExpr.body", b)
		}
	}

	if cref == nil {
		cref = gs.findClass(owner, alloc_name)
		if cref == nil {
			cref = &GoClassReference{name: alloc_name, parent: owner}
			gs.addClass(cref)
		}
	}

	if type_args != nil && len(type_args) > 0 {
		if gs.Program().verbose {
			log.Printf("//ERR// Not handling %v type_args\n", alloc_name)
		} else {
			log.Printf("//ERR// Not handling clsalloc type_args\n")
		}

		type_args = nil
	}

	if body != nil && len(body) > 0 {
		if gs.Program().verbose {
			log.Printf(
				"//ERR// Ignoring %v class alloc body (%d stmts)\n",
				cref.Name(), len(body))
		} else {
			log.Printf("//ERR// Not handling clsalloc body\n")
		}
	}

	arglist := &GoMethodArguments{args}
	mthd := findMethod(owner, cref, "New" + cref.Name(), arglist,
		gs.Program().verbose)

	return &GoClassAlloc{class: cref, method: mthd,  type_args: type_args,
		args: args, body: body}
}

func analyzeArrayAlloc(gs *GoState, owner GoMethodOwner, aa *grammar.JArrayAlloc,
	govar GoVar) GoArrayExpr {
	td := gs.Program().createTypeData(aa.Typename, nil, aa.Dims)

	if aa.Dimexprs != nil && len(aa.Dimexprs) > 0 && aa.Init != nil &&
		len(aa.Init) > 0 {
		panic(fmt.Sprintf("JArrayAlloc has both dimexprs#%d and init#%d",
			len(aa.Dimexprs), len(aa.Init)))
	}

	if aa.Dimexprs != nil && len(aa.Dimexprs) > 0 {
		args := make([]GoExpr, len(aa.Dimexprs))
		for i, arg := range aa.Dimexprs {
			args[i] = analyzeExpr(gs, owner, arg)
		}

		return &GoArrayAlloc{typedata: td, args: args}
	}

	var elements []GoExpr
	if aa.Init != nil && len(aa.Init) > 0 {
		elements = make([]GoExpr, len(aa.Init))
		for i, elem := range aa.Init {
			elements[i] = analyzeVariableInit(gs, owner, elem, govar)
		}
	}

	return &GoArrayInit{typedata: td, elems: elements}
}

func analyzeAssignExpr(gs *GoState, owner GoMethodOwner,
	expr *grammar.JAssignmentExpr) *GoAssign {
	var op token.Token
	switch expr.Op {
	case "=":
		op = token.ASSIGN
	case "+=":
		op = token.ADD_ASSIGN
	case "-=":
		op = token.SUB_ASSIGN
	case "*=":
		op = token.MUL_ASSIGN
	case "/=":
		op = token.QUO_ASSIGN
	case "%=":
		op = token.REM_ASSIGN
	case "&=":
		op = token.AND_ASSIGN
	case "|=":
		op = token.OR_ASSIGN
	case "^=":
		op = token.XOR_ASSIGN
	case "<<=":
		op = token.SHL_ASSIGN
	case ">>=":
		op = token.SHR_ASSIGN
	case "&^=":
		op = token.AND_NOT_ASSIGN
	default:
		panic(fmt.Sprintf("Unknown assignment operator '%s'", expr.Op))
	}

	var lhs GoVar
	switch v := expr.Left.(type) {
	case *grammar.JArrayReference:
		lhs = analyzeArrayReference(gs, owner, v)
	case *grammar.JConditionalExpr:
		log.Printf("//ERR// Faking asgnexpr lhs condexpr\n")
		lhs = NewFakeVar("<<condexpr>>", nil, 0)
	case *grammar.JNameDotObject:
		log.Printf("//ERR// Faking asgnexpr lhs NDO\n")
		lhs = NewFakeVar(v.Name.String(), nil, 0)
	case *grammar.JObjectDotName:
		lhs = analyzeObjectDotName(gs, owner, v)
	case *grammar.JReferenceType:
		lhs = analyzeReferenceType(gs, v)
	default:
		panic(fmt.Sprintf("//ERR// Unknown asgnexpr LHS %T\n", expr.Left))
	}

	rhs := make([]GoExpr, 1)
	rhs[0] = analyzeExpr(gs, owner, expr.Right)

	return &GoAssign{govar: lhs, tok: op, rhs: rhs}
}

func analyzeBinaryExpr(gs *GoState, owner GoMethodOwner,
	bexpr *grammar.JBinaryExpr) *GoBinaryExpr {
	var op token.Token
	switch bexpr.Op {
	case "+":
		op = token.ADD
	case "-":
		op = token.SUB
	case "*":
		op = token.MUL
	case "/":
		op = token.QUO
	case "%":
		op = token.REM
	case "&":
		op = token.AND
	case "|":
		op = token.OR
	case "^":
		op = token.XOR
	case "<<":
		op = token.SHL
	case ">>":
		op = token.SHR
	case "&&":
		op = token.LAND
	case "||":
		op = token.LOR
	case "==":
		op = token.EQL
	case "!=":
		op = token.NEQ
	case "<":
		op = token.LSS
	case ">":
		op = token.GTR
	case "<=":
		op = token.LEQ
	case ">=":
		op = token.GEQ
	case ">>>":
		op = token.SHR
	default:
		panic(fmt.Sprintf("Unknown binary operator \"%s\"", bexpr.Op))
	}

	return &GoBinaryExpr{x: analyzeExpr(gs, owner, bexpr.Obj1), op: op,
		y: analyzeExpr(gs, owner, bexpr.Obj2)}
}

func analyzeBlock(gs *GoState, owner GoMethodOwner, blk *grammar.JBlock) *GoBlock {
	stmts := make([]GoStatement, 0)

	gs2 := NewGoState(gs)

	for _, bobj := range blk.List {
		switch b := bobj.(type) {
		case *grammar.JBlock:
			blk := analyzeBlock(gs2, owner, b)
			if blk != nil {
				stmts = append(stmts, blk)
			}
		case *grammar.JClassDecl:
			gs2.addClassDecl(owner, b)
		case *grammar.JEnumDecl:
			gs2.Program().addEnum(b)
		case *grammar.JEmpty:
			continue
		case *grammar.JForColon:
			stmt := analyzeForColon(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *grammar.JForExpr:
			stmt := analyzeForExpr(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *grammar.JForVar:
			stmt := analyzeForVar(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *grammar.JIfElseStmt:
			stmt := analyzeIfElseStmt(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *grammar.JJumpToLabel:
			if b != nil {
				stmts = append(stmts, NewGoJumpToLabel(b.Label, b.IsContinue))
			}
		case *grammar.JLabeledStatement:
			stmt := analyzeLabelledStmt(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *grammar.JLocalVariableDecl:
			lvlist := analyzeLocalVariableDeclaration(gs2, owner, b)
			if lvlist != nil && len(lvlist) > 0 {
				stmts = append(stmts, lvlist...)
			}
		case *grammar.JMethodDecl:
			if cls, ok := owner.(*GoClassDefinition); ok {
				cls.AddNewMethod(gs2, b)
			} else {
				log.Printf("//ERR// Cannot add new method to %T\n", owner)
			}
		case *grammar.JSimpleStatement:
			ss := analyzeSimpleStatement(gs2, owner, b)
			if ss != nil {
				stmts = append(stmts, ss)
			}
		case *grammar.JSynchronized:
			ss := analyzeSynchronized(gs2, owner, b)
			if ss != nil {
				stmts = append(stmts, ss)
			}
		case *grammar.JSwitch:
			swtch := analyzeSwitch(gs2, owner, b)
			if swtch != nil {
				stmts = append(stmts, swtch)
			}
		case *grammar.JTry:
			try := analyzeTry(gs2, owner, b)
			if try != nil {
				stmts = append(stmts, try)
			}
		case *grammar.JUnimplemented:
			log.Printf("//ERR// Ignoring unimplemented block %s\n", b.TypeStr)
		case *grammar.JWhile:
			while := analyzeWhile(gs2, owner, b)
			if while != nil {
				stmts = append(stmts, while)
			}
		case nil:
			log.Printf("//ERR// Ignoring nil block entry\n")
		default:
			grammar.ReportCastError("Block", bobj)
		}
	}

	return &GoBlock{stmts: stmts}
}

func analyzeBranchStmt(gs *GoState, owner GoMethodOwner, tok token.Token,
	obj grammar.JObject) *GoBranchStmt {
	var lexpr GoExpr
	if obj != nil {
		lexpr = analyzeExpr(gs, owner, obj)
	}

	var label string
	if lexpr != nil {
		if gs.Program().verbose {
			log.Printf("//ERR// cannot convert %s label %v<%T>\n", tok,
				lexpr, lexpr)
		} else {
			log.Printf("//ERR// cannot convert branch label %T\n", lexpr)
		}
	}

	return &GoBranchStmt{tok: tok, label: label}
}

func analyzeCastExpr(gs *GoState, owner GoMethodOwner,
	cex *grammar.JCastExpr) *GoCastType {
	td := gs.Program().createTypeData(cex.Reftype.Name,
		cex.Reftype.TypeArgs, cex.Reftype.Dims)
	return &GoCastType{target: analyzeExpr(gs, owner, cex.Target), casttype: td}
}

func analyzeClassBody(gs *GoState, cls *GoClassDefinition, body *grammar.JClassBody) {
	for _, bobj := range body.List {
		switch b := bobj.(type) {
		case *grammar.JClassDecl:
			gs.addClassDecl(cls, b)
		case *grammar.JEnumDecl:
			gs.Program().addEnum(b)
		case *grammar.JInterfaceDecl:
			log.Printf("//ERR// Ignoring class body %T\n", b)
/*
			idecl := analyzeInterfaceDeclaration(gs, b)
			if idecl != nil {
				decls = append(decls, idecl)
			}
*/
		case *grammar.JMethodDecl:
			cls.AddNewMethod(gs, b)
		case *grammar.JUnimplemented:
			log.Printf("//ERR// Ignoring unimplemented body %s\n", b.TypeStr)
		case *grammar.JVariableDecl:
			govar := gs.addVariableDecl(b, true)

			if cls.vars == nil {
				cls.vars = make([]*GoVarInit, 0)
			}

			var goinit *GoVarInit
			if b.Init == nil {
				goinit = &GoVarInit{govar: govar}
			} else {
				goinit = analyzeVariableInit(gs, cls, b.Init, govar)
			}
			cls.vars = append(cls.vars, goinit)
		default:
			grammar.ReportCastError("Body.ClassDecl", bobj)
		}
	}
}

func analyzeConstant(gs *GoState, owner GoMethodOwner, jcon *grammar.JConstantDecl) {
	ctype := gs.Program().createTypeData(jcon.TypeSpec.Name,
		jcon.TypeSpec.TypeArgs, jcon.Dims)
	init := analyzeVariableInit(gs, owner, jcon.Init, nil)
	owner.AddConstant(&GoConstant{name: jcon.Name, typedata: ctype, init: init})
}

func analyzeExpr(gs *GoState, owner GoMethodOwner, obj grammar.JObject) GoExpr {
	switch e := obj.(type) {
	case *grammar.JArrayAlloc:
		return analyzeArrayAlloc(gs, owner, e, nil)
	case *grammar.JAssignmentExpr:
		return analyzeAssignExpr(gs, owner, e)
	case *grammar.JBinaryExpr:
		return analyzeBinaryExpr(gs, owner, e)
	case *grammar.JCastExpr:
		return analyzeCastExpr(gs, owner, e)
	case *grammar.JClassAllocationExpr:
		return analyzeAllocationExpr(gs, owner, e)
	case *grammar.JConditionalExpr:
		if gs.Program().verbose {
			log.Printf("//ERR// Not converting condexpr %T (%T|?|%T|:|%T) to Expr\n",
				e, e.CondExpr, e.IfExpr, e.ElseExpr)
		} else {
			log.Printf("//ERR// Not converting condexpr to Expr\n")
		}
		return &GoUnimplemented{fname: "expr", text: fmt.Sprintf("%T", e)}
	case *grammar.JInstanceOf:
		return &GoInstanceOf{expr: analyzeExpr(gs, owner, e.Obj),
			vartype: analyzeReferenceType(gs, e.TypeSpec)}
	case *grammar.JKeyword:
		return NewGoKeyword(e.Token, e.Name)
	case *grammar.JLiteral:
		return NewGoLiteral(e.Text)
	case *grammar.JMethodAccess:
		return analyzeMethodAccess(gs, owner, e)
	case *grammar.JNameDotObject:
		return analyzeNameDotObject(gs, e)
	case *grammar.JArrayReference:
		return analyzeArrayReference(gs, owner, e)
	case *grammar.JObjectDotName:
		return analyzeObjectDotName(gs, owner, e)
	case *grammar.JParens:
		return analyzeExpr(gs, owner, e.Expr)
	case *GoPrimitiveType:
		return e
	case *grammar.JReferenceType:
		return analyzeReferenceType(gs, e)
	case *grammar.JUnaryExpr:
		return analyzeUnaryExpr(gs, owner, e)
	case *grammar.JUnimplemented:
		log.Printf("//ERR// Not analyzing unimplemented expr %s\n", e.TypeStr)
		return &GoUnimplemented{fname: "expr",
			text: fmt.Sprintf("%s", e.TypeStr)}
	case *grammar.JVariableInit:
		return analyzeVariableInit(gs, owner, e, nil)
	}

	panic(fmt.Sprintf("Unknown expression type %T", obj))
}

func analyzeForColon(gs *GoState, owner GoMethodOwner,
	jfc *grammar.JForColon) *GoForColon {
	gs2 := NewGoState(gs)

	govar := gs2.addVariableDecl(jfc.VarDecl, false)
	if govar == nil {
		panic("ForColon addVariable returned nil")
	}

	var expr GoExpr
	if jfc.Expr != nil {
		expr = analyzeExpr(gs2, owner, jfc.Expr)
	}

	var body *GoBlock
	if jfc.Body != nil {
		body = makeBlock(gs2, owner, jfc.Body)
	}

	if body == nil {
		log.Printf("//ERR// adding empty forcolon body\n")
		list := make([]GoStatement, 0)
		body = &GoBlock{stmts: list}
	}

	return &GoForColon{govar: govar, expr: expr, body: body}
}

func analyzeForExpr(gs *GoState, owner GoMethodOwner,
	jfor *grammar.JForExpr) *GoForExpr {
	gs2 := NewGoState(gs)

	fe := &GoForExpr{}

	if jfor.Init != nil && len(jfor.Init) > 0 {
		fe.init = make([]GoExpr, len(jfor.Init))
		for i, v := range jfor.Init {
			fe.init[i] = analyzeExpr(gs2, owner, v)
		}
	}

	if jfor.Expr != nil {
		fe.cond = analyzeExpr(gs2, owner, jfor.Expr)
	}

	if jfor.Incr != nil && len(jfor.Incr) > 0 {
		fe.incr = make([]GoExpr, len(jfor.Incr))
		for i, v := range jfor.Incr {
			fe.incr[i] = analyzeExpr(gs2, owner, v)
		}
	}

	if jfor.Body != nil {
		fe.block = makeBlock(gs2, owner, jfor.Body)
	}

	return fe
}

func analyzeForVar(gs *GoState, owner GoMethodOwner, jfv *grammar.JForVar) *GoForVar {
	gs2 := NewGoState(gs)

	forvar := &GoForVar{govar: gs2.addVariableDecl(jfv.VarDecl, false)}

	if jfv.VarDecl.Init != nil {
		forvar.init = analyzeVariableInit(gs2, owner, jfv.VarDecl.Init,
			forvar.govar)
	}

	if jfv.Decl != nil {
		log.Printf("//ERR// ignoring forvar decl %T\n", jfv.Decl)
	}

	//multiple initialization; a consolidated bool expression with && and ||; multiple 'incrementation'
	// for i, j, s := 0, 5, "a"; i < 3 && j < 100 && s != "aaaaa"; i, j, s = i+1, j+1, s + "a"  {

	if jfv.Expr != nil {
		forvar.cond = analyzeExpr(gs2, owner, jfv.Expr)
	}

	if jfv.Incr != nil && len(jfv.Incr) > 0 {
		forvar.incr = make([]GoStatement, 0)
		for _, v := range jfv.Incr {
			stmts := analyzeStmt(gs2, owner, v)
			if stmts != nil && len(stmts) > 0 {
				forvar.incr = append(forvar.incr, stmts...)
			}
		}
	}

	if jfv.Body != nil {
		forvar.block = makeBlock(gs2, owner, jfv.Body)
	}

	return forvar
}

func analyzeIfElseStmt(gs *GoState, owner GoMethodOwner,
	ifelse *grammar.JIfElseStmt) *GoIfElse {

	stmts := analyzeStmt(gs, owner, ifelse.IfBlock)
	if stmts == nil || len(stmts) == 0 {
		panic("IfStmt body cannot be nil")
	}

	var ifblk GoStatement
	if len(stmts) == 1 {
		ifblk = stmts[0]
	} else {
		ifblk = &GoBlock{stmts: stmts}
	}

	ifstmt := &GoIfElse{cond: analyzeExpr(gs, owner, ifelse.Cond),
		ifblk: ifblk}

	if ifelse.ElseBlock != nil {
		stmts = analyzeStmt(gs, owner, ifelse.ElseBlock)
		if stmts != nil && len(stmts) > 0 {
			if len(stmts) == 1 {
				ifstmt.elseblk = stmts[0]
			} else {
				ifstmt.elseblk = &GoBlock{stmts: stmts}
			}
		}
	}

	return ifstmt
}

func analyzeLabelledStmt(gs *GoState, owner GoMethodOwner,
	ls *grammar.JLabeledStatement) *GoLabeledStmt {
	stmts := analyzeStmt(gs, owner, ls.Stmt)
	if len(stmts) != 1 {
		panic("Found label assigned to multiple statements")
	}

	return &GoLabeledStmt{label: ls.Label, stmt: stmts[0]}
}

func analyzeLocalVariableDeclaration(gs *GoState, owner GoMethodOwner,
	vdec *grammar.JLocalVariableDecl) []GoStatement {
	if vdec.Vars == nil {
		return nil
	}

	stmts := make([]GoStatement, 0)
	for i, lvar := range vdec.Vars {
		lv := analyzeLocalVariableInternal(gs, owner, vdec.Modifiers,
			vdec.TypeSpec, lvar, i)
		stmts = append(stmts, lv)
	}

	return stmts
}

func analyzeLocalVariableInternal(gs *GoState, owner GoMethodOwner,
	defaultModifiers *grammar.JModifiers, defaultTypeSpec *grammar.JReferenceType,
	vardec *grammar.JVariableDecl, idx int) GoStatement {

	var typespec *grammar.JReferenceType
	if vardec.TypeSpec != nil {
		typespec = vardec.TypeSpec
	} else {
		typespec = defaultTypeSpec
	}

	var mods *grammar.JModifiers
	if vardec.Modifiers != nil {
		mods = vardec.Modifiers
	} else {
		mods = defaultModifiers
	}

	govar := gs.addVariable(vardec.Name, mods, vardec.Dims, typespec, false)
	if govar == nil {
		panic("addVariable returned nil")
	}

	if vardec.Init == nil {
		return NewGoLocalVarNoInit(govar)
	}

	if vardec.Init.ArrayList != nil {
		init := analyzeVariableInit(gs, owner, vardec.Init, govar)
		return NewGoLocalVarInit(govar, init)
	}

	cex, ok := vardec.Init.Expr.(*grammar.JCastExpr)
	if !ok {
		init := analyzeExpr(gs, owner, vardec.Init.Expr)
		return NewGoLocalVarInit(govar, init)
	}

	return NewGoLocalVarCast(govar, analyzeCastExpr(gs, owner, cex))
}

func analyzeMethodAccess(gs *GoState, owner GoMethodOwner,
	mth *grammar.JMethodAccess) GoExpr {
	arglist := NewGoMethodArguments(gs, owner, mth.ArgList)

	if mth.NameKey != nil {
		is_super := mth.NameKey.Token == grammar.SUPER
		if !is_super && mth.NameKey.Token != grammar.THIS {
			panic(fmt.Sprintf("Keyword %v is neither THIS nor SUPER",
				mth.NameKey))
		}

		return &GoMethodAccessKeyword{is_super: is_super, args: arglist}
	}

	if mth.Method == "" {
		panic("JMethodAccess name is nil")
	}

	var expr GoExpr
	if mth.NameObj != nil {
		expr = analyzeExpr(gs, owner, mth.NameObj)

		var class GoMethodOwner
		if macc, ok := expr.(*GoMethodAccess); ok {
			class = macc.method.Class()
		} else {
			tmpcls := gs.Class()
			if tmpcls != nil {
				class = tmpcls
			} else {
				class = nilMethodOwner
			}
		}

		mthd := findMethod(owner, class, mth.Method, arglist,
			gs.Program().verbose)

		return &GoMethodAccessExpr{expr: expr, method: mthd, args: arglist}
	}

	var class GoMethodOwner
	var govar GoVar
	if mth.NameType == nil {
		class = nilMethodOwner
	} else {
		govar = gs.findVariable(mth.NameType)
		if govar != nil {
			class = owner
		} else {
			class = gs.findClass(owner, mth.NameType.LastType())
			if class == nil {
				class = &GoFakeClass{name: mth.NameType.String()}
			}
		}
	}

	mthd := findMethod(owner, class, mth.Method, arglist, gs.Program().verbose)

	if govar != nil {
		return &GoMethodAccessVar{govar: govar, method: mthd, args: arglist}
	}

	return &GoMethodAccess{method: mthd, args: arglist}
}

func analyzeNameDotObject(gs *GoState, ndo *grammar.JNameDotObject) GoExpr {
	switch o := ndo.Obj.(type) {
	case *grammar.JKeyword:
		log.Printf("//ERR// Not converting ndoobj %T (kwd %s)\n", ndo.Obj, o.Name)
	default:
		log.Printf("//ERR// Not converting ndoobj %T (%T) to Expr\n", ndo, ndo.Obj)
	}

	return &GoUnimplemented{fname: "ndo", text: fmt.Sprintf("%T", ndo.Obj)}
}

func analyzeArrayReference(gs *GoState, owner GoMethodOwner,
	nae *grammar.JArrayReference) *GoArrayReference {
	var govar GoVar
	var obj GoExpr
	if nae.Name != nil {
		govar = gs.findOrFakeVariable(nae.Name, "arrayref")
	} else {
		obj = analyzeExpr(gs, owner, nae.Obj)
	}

	return &GoArrayReference{govar: govar, obj: obj,
		index: analyzeExpr(gs, owner, nae.Expr)}
}

func analyzeObjectDotName(gs *GoState, owner GoMethodOwner,
	odn *grammar.JObjectDotName) GoVar {
	switch o := odn.Obj.(type) {
	case *grammar.JKeyword:
		if o.Token == grammar.THIS && gs.Receiver() != "" {
			rvar := gs.findVariable(grammar.NewJTypeName(gs.Receiver(), false))
			if rvar == nil {
				rvar = gs.addVariable(gs.Receiver(), nil, 0, nil, false)
			}

			ref := gs.findVariable(odn.Name)
			if ref == nil {
				ref = gs.addVariable(gs.Receiver(), nil, 0, nil, false)
			}

			return NewGoSelector(rvar, ref)
		} else if o.Token == grammar.SUPER && gs.Receiver() != "" {
			log.Printf("//ERR// Not converting odnobj super\n")
			return NewFakeVar("<<super>>", nil, 0)
		} else {
			log.Printf("//ERR// Not converting odnobj %T (kwd %s)\n",
				odn.Obj, o.Name)
			return NewFakeVar(fmt.Sprintf("<<%v>>", o.Name), nil, 0)
		}
	default:
		return NewObjectDotName(odn, analyzeExpr(gs, owner, odn.Obj), gs)
	}
}

func analyzeReferenceType(gs *GoState, ref *grammar.JReferenceType) GoVar {
	if ref.TypeArgs != nil && len(ref.TypeArgs) > 0 {
		fmt.Sprintf("//ERR// Not handling reftype type_args in %v\n", ref.Name)
	}

	govar := gs.findVariable(ref.Name)
	if govar != nil {
		return govar
	}

	return NewFakeVar(ref.Name.String(), ref.TypeArgs, ref.Dims)
}

func analyzeSimpleStatement(gs *GoState, owner GoMethodOwner,
	jstmt *grammar.JSimpleStatement) GoStatement {

	if jstmt.Keyword != nil {
		switch jstmt.Keyword.Token {
		case grammar.BREAK:
			return analyzeBranchStmt(gs, owner, token.BREAK, jstmt.Object)
		case grammar.CONTINUE:
			return analyzeBranchStmt(gs, owner, token.CONTINUE, jstmt.Object)
		case grammar.RETURN:
			var expr GoExpr
			if jstmt.Object != nil {
				expr = analyzeExpr(gs, owner, jstmt.Object)
			}

			return &GoReturn{expr: expr}
		case grammar.THROW:
			return &GoThrow{expr: analyzeExpr(gs, owner, jstmt.Object)}
		default:
			return &GoUnimplemented{fname: "simpstmt",
				text: jstmt.Keyword.Name}
		}
	}

	switch expr := jstmt.Object.(type) {
	case *grammar.JAssignmentExpr:
		return analyzeAssignExpr(gs, owner, expr)
	case *grammar.JClassAllocationExpr:
		return &GoExprStmt{x: analyzeAllocationExpr(gs, owner, expr)}
	case *grammar.JMethodAccess:
		return &GoExprStmt{x: analyzeMethodAccess(gs, owner, expr)}
	case *grammar.JUnaryExpr:
		return analyzeUnaryExpr(gs, owner, expr)
	default:
		log.Printf("//ERR// -------- not analyzing simpstmt %T\n", jstmt.Object)
		return &GoUnimplemented{fname: "simpstmt",
			text: fmt.Sprintf("%T", jstmt.Object)}
	}
}

func analyzeStmt(gs *GoState, owner GoMethodOwner,
	jstmt grammar.JObject) []GoStatement {
	switch stmt := jstmt.(type) {
	case *grammar.JAssignmentExpr:
		return []GoStatement{ analyzeAssignExpr(gs, owner, stmt), }
	case *grammar.JBlock:
		return []GoStatement{ analyzeBlock(gs, owner, stmt), }
	case *grammar.JForColon:
		return []GoStatement{ analyzeForColon(gs, owner, stmt), }
	case *grammar.JForVar:
		return []GoStatement{ analyzeForVar(gs, owner, stmt), }
	case *grammar.JIfElseStmt:
		return []GoStatement{ analyzeIfElseStmt(gs, owner, stmt), }
	case *grammar.JJumpToLabel:
		return []GoStatement{ NewGoJumpToLabel(stmt.Label, stmt.IsContinue), }
	case *grammar.JLocalVariableDecl:
		return analyzeLocalVariableDeclaration(gs, owner, stmt)
	case *grammar.JSimpleStatement:
		return []GoStatement{ analyzeSimpleStatement(gs, owner, stmt), }
	case *grammar.JTry:
		return []GoStatement{ analyzeTry(gs, owner, stmt), }
	case *grammar.JUnaryExpr:
		return []GoStatement{ analyzeUnaryExpr(gs, owner, stmt), }
	case *grammar.JUnimplemented:
		log.Printf("//ERR// Not analyzing unimplemented stmt %s\n", stmt.TypeStr)
		return []GoStatement{ &GoUnimplemented{fname: "stmt",
			text: stmt.TypeStr}, }
	case *grammar.JWhile:
		return []GoStatement{ analyzeWhile(gs, owner, stmt), }
	default:
		log.Printf("//ERR// Not analyzing stmt %T\n", stmt)
		return []GoStatement{ &GoUnimplemented{fname: "stmt",
			text: fmt.Sprintf("%T", stmt)}, }
	}
}

func analyzeSwitch(gs *GoState, owner GoMethodOwner,
	jsw *grammar.JSwitch) *GoSwitch {
	if jsw.Groups == nil || len(jsw.Groups) == 0 {
		// ignore empty switch statements
		return nil
	}

	gsw := &GoSwitch{expr: analyzeExpr(gs, owner, jsw.Expr),
		cases: make([]*GoSwitchCase, len(jsw.Groups))}
	for i, c := range jsw.Groups {
		gsw.cases[i] = analyzeSwitchCase(gs, owner, c)
	}

	return gsw
}

func analyzeSwitchCase(gs *GoState, owner GoMethodOwner,
	jsg *grammar.JSwitchGroup) *GoSwitchCase {
	if jsg.Labels == nil || len(jsg.Labels) == 0 {
		panic("No labels for switch case")
	}

	labels := make([]*GoSwitchLabel, len(jsg.Labels))
	for i, l := range jsg.Labels {
		labels[i] = analyzeSwitchLabel(gs, owner, l)
	}

	stmts := make([]GoStatement, 0)
	for _, s := range jsg.Stmts {
		gs2 := NewGoState(gs)

		st := analyzeStmt(gs2, owner, s)
		if st != nil && len(st) > 0 {
			stmts = append(stmts, st...)
		}
	}

	// if last statement is 'break', delete it (it's implicit in Go)
	// if it's not 'break', add a fallthrough
	need_fall := true
	l := len(stmts)
	if l > 0 {
		switch br := stmts[l - 1].(type) {
		case *GoBranchStmt:
			if br.tok == token.BREAK {
				stmts = stmts[0:l-1]
				need_fall = false
			}
		case *GoJumpToLabel:
			need_fall = false
		}
	}

	if need_fall {
		stmts = append(stmts, &GoBranchStmt{tok: token.FALLTHROUGH})
	}

	return &GoSwitchCase{labels: labels, stmts: stmts}
}

func analyzeSwitchLabel(gs *GoState, owner GoMethodOwner,
	jsl *grammar.JSwitchLabel) *GoSwitchLabel {
	if jsl.IsDefault {
		return &GoSwitchLabel{is_default: true}
	}

	var expr GoExpr
	if jsl.Name != "" {
		expr = &GoLiteral{text: jsl.Name}
	} else if  jsl.Expr != nil {
		expr = analyzeExpr(gs, owner, jsl.Expr)
	} else {
		panic("Empty switch label")
	}

	return &GoSwitchLabel{expr: expr}
}

func analyzeSynchronized(gs *GoState, owner GoMethodOwner,
	sync *grammar.JSynchronized) *GoSynchronized {
	return &GoSynchronized{expr: analyzeExpr(gs, owner, sync.Expr),
		block: analyzeBlock(gs, owner, sync.Block)}
}

func analyzeTry(gs *GoState, owner GoMethodOwner, try *grammar.JTry) *GoTry {
	gt := &GoTry{block: analyzeBlock(NewGoState(gs), owner, try.Block)}

	if try.Catches != nil && len(try.Catches) > 0 {
		gs2 := NewGoState(gs)

		gt.catches = make([]*GoTryCatch, len(try.Catches))
		for i, c := range try.Catches {
			var exc *grammar.JTypeName
			if len(c.TypeList) == 1 {
				exc = c.TypeList[0]
			} else {
				exc = grammar.NewJTypeName("Exception", false)
			}

			govar := gs.addVariable(c.Name, c.Modifiers, 0,
				grammar.NewJReferenceType(exc, nil, 0), false)

			if len(c.TypeList) != 1 {
				log.Printf("//ERR// Ignoring catch#%d long typelist (len=%d)\n",
					i, len(c.TypeList))
			}

			gt.catches[i] = &GoTryCatch{govar: govar,
				block: analyzeBlock(gs2, owner, c.Block)}
		}
	}

	if try.Finally != nil {
		gt.finally = analyzeBlock(NewGoState(gs), owner, try.Finally)
	}

	return gt
}

func analyzeUnaryExpr(gs *GoState, owner GoMethodOwner, uexpr *grammar.JUnaryExpr) *GoUnaryExpr {
	var op token.Token
	switch uexpr.Op {
	case "++":
		op = token.INC
	case "--":
		op = token.DEC
	case "!":
		op = token.NOT
	case "~":
		op = token.XOR
	case "+":
		op = token.ADD
	case "-":
		op = token.SUB
	default:
		panic(fmt.Sprintf("Unknown unary operator \"%s\"", uexpr.Op))
	}

	return &GoUnaryExpr{op: op, x: analyzeExpr(gs, owner, uexpr.Obj)}
}

func analyzeVariableInit(gs *GoState, owner GoMethodOwner,
	init *grammar.JVariableInit, govar GoVar) *GoVarInit {
	if init.Expr != nil {
		var expr GoExpr
		switch v := init.Expr.(type) {
		case *grammar.JVariableInit:
			expr = analyzeVariableInit(gs, owner, v, govar)
		case *grammar.JArrayAlloc:
			expr = analyzeArrayAlloc(gs, owner, v, govar)
		default:
			expr = analyzeExpr(gs, owner, init.Expr)
		}

		return &GoVarInit{govar: govar, expr: expr}
	}

	if init.ArrayList == nil {
		panic("No variable Initialization")
	}

	elements := make([]GoExpr, len(init.ArrayList))
	for i, elem := range init.ArrayList {
		elements[i] = analyzeVariableInit(gs, owner, elem, govar)
	}

	return &GoVarInit{govar: govar, elements: elements}
}

func analyzeWhile(gs *GoState, owner GoMethodOwner, while *grammar.JWhile) *GoWhile {
	gw := &GoWhile{expr: analyzeExpr(gs, owner, while.Expr),
		is_do_while: while.IsDoWhile}

	stmts := analyzeStmt(gs, owner, while.Stmt)
	if stmts != nil && len(stmts) > 0 {
		if len(stmts) == 1 {
			gw.stmt = stmts[0]
		} else {
			gw.stmt = &GoBlock{stmts: stmts}
		}
	}

	return gw
}

func makeBlock(gs *GoState, owner GoMethodOwner, block grammar.JObject) *GoBlock {
	stmts := analyzeStmt(gs, owner, block)
	if stmts == nil || len(stmts) == 0 {
		return nil
	}

	if len(stmts) == 1 {
		if blk, ok := stmts[0].(*GoBlock); ok {
			return blk
		}
	}

	return &GoBlock{stmts: stmts}
}
