package parser

import (
	"fmt"
	"go/token"
	"log"
	//"os"
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
	alloc *JClassAllocationExpr) *GoClassAlloc {
	if alloc.name.IsPrimitive() {
		panic(fmt.Sprintf("Class allocation should not use primitive \"%s\"",
			alloc.name.String()))
	}

	var type_args []*TypeData
	if alloc.type_args != nil && len(alloc.type_args) > 0 {
		type_args = make([]*TypeData, len(alloc.type_args))
		for i, arg := range alloc.type_args {
			if arg.ts_type != TS_NONE {
				log.Printf("//ERR// Ignoring allexpr ts#%d type %v\n",
					i, arg.ts_type)
			}

			if arg.typespec == nil {
				type_args[i] = genericObject
			} else {
				type_args[i] = gs.Program().createTypeData(arg.typespec.name,
					arg.typespec.type_args, 0)
			}
		}
	}

	var args []GoExpr
	if alloc.arglist != nil && len(alloc.arglist) > 0 {
		args = make([]GoExpr, len(alloc.arglist))
		for i, arg := range alloc.arglist {
			args[i] = analyzeExpr(gs, owner, arg)
		}
	}

	alloc_name := alloc.name.String()

	var cref GoClass

	var body []GoStatement
	for _, b := range alloc.body {
		switch j := b.(type) {
		case *JClassBody:
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
		case *JBlock:
			stmt := analyzeBlock(gs, owner, j)
			if stmt != nil {
				body = append(body, stmt)
			}
		case *GoEmpty:
			; // do nothing
		default:
			reportCastError("JClassAllocationExpr.body", b)
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

func analyzeArrayAlloc(gs *GoState, owner GoMethodOwner, aa *JArrayAlloc,
	govar GoVar) GoArrayExpr {
	td := gs.Program().createTypeData(aa.typename, nil, aa.dims)

	if aa.dimexprs != nil && len(aa.dimexprs) > 0 && aa.init != nil &&
		len(aa.init) > 0 {
		panic(fmt.Sprintf("JArrayAlloc has both dimexprs#%d and init#%d",
			len(aa.dimexprs), len(aa.init)))
	}

	if aa.dimexprs != nil && len(aa.dimexprs) > 0 {
		args := make([]GoExpr, len(aa.dimexprs))
		for i, arg := range aa.dimexprs {
			args[i] = analyzeExpr(gs, owner, arg)
		}

		return &GoArrayAlloc{typedata: td, args: args}
	}

	var elements []GoExpr
	if aa.init != nil && len(aa.init) > 0 {
		elements = make([]GoExpr, len(aa.init))
		for i, elem := range aa.init {
			elements[i] = analyzeVariableInit(gs, owner, elem, govar)
		}
	}

	return &GoArrayInit{typedata: td, elems: elements}
}

func analyzeAssignExpr(gs *GoState, owner GoMethodOwner,
	expr *JAssignmentExpr) *GoAssign {
	var op token.Token
	switch expr.op {
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
		panic(fmt.Sprintf("Unknown assignment operator '%s'", expr.op))
	}

	var lhs GoVar
	switch v := expr.left.(type) {
	case *JArrayReference:
		lhs = analyzeArrayReference(gs, owner, v)
	case *JConditionalExpr:
		log.Printf("//ERR// Faking asgnexpr lhs condexpr\n")
		lhs = NewFakeVar("<<condexpr>>", nil, 0)
	case *JNameDotObject:
		log.Printf("//ERR// Faking asgnexpr lhs NDO\n")
		lhs = NewFakeVar(v.name.String(), nil, 0)
	case *JObjectDotName:
		lhs = analyzeObjectDotName(gs, owner, v)
	case *JReferenceType:
		lhs = analyzeReferenceType(gs, v)
	default:
		panic(fmt.Sprintf("//ERR// Unknown asgnexpr LHS %T\n", expr.left))
	}

	rhs := make([]GoExpr, 1)
	rhs[0] = analyzeExpr(gs, owner, expr.right)

	return &GoAssign{govar: lhs, tok: op, rhs: rhs}
}

func analyzeBinaryExpr(gs *GoState, owner GoMethodOwner,
	bexpr *JBinaryExpr) *GoBinaryExpr {
	var op token.Token
	switch bexpr.op {
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
		panic(fmt.Sprintf("Unknown binary operator \"%s\"", bexpr.op))
	}

	return &GoBinaryExpr{x: analyzeExpr(gs, owner, bexpr.obj1), op: op,
		y: analyzeExpr(gs, owner, bexpr.obj2)}
}

func analyzeBlock(gs *GoState, owner GoMethodOwner, blk *JBlock) *GoBlock {
	stmts := make([]GoStatement, 0)

	gs2 := NewGoState(gs)

	for _, bobj := range blk.list {
		switch b := bobj.(type) {
		case *JBlock:
			blk := analyzeBlock(gs2, owner, b)
			if blk != nil {
				stmts = append(stmts, blk)
			}
		case *JClassDecl:
			gs2.addClassDecl(owner, b)
		case *JEnumDecl:
			gs2.Program().addEnum(b)
		case *GoEmpty:
			continue
		case *JForColon:
			stmt := analyzeForColon(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *JForExpr:
			stmt := analyzeForExpr(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *JForVar:
			stmt := analyzeForVar(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *JIfElseStmt:
			stmt := analyzeIfElseStmt(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *GoJumpToLabel:
			if b != nil {
				stmts = append(stmts, b)
			}
		case *JLabeledStatement:
			stmt := analyzeLabelledStmt(gs2, owner, b)
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		case *JLocalVariableDecl:
			lvlist := analyzeLocalVariableDeclaration(gs2, owner, b)
			if lvlist != nil && len(lvlist) > 0 {
				stmts = append(stmts, lvlist...)
			}
		case *JMethodDecl:
			if cls, ok := owner.(*GoClassDefinition); ok {
				cls.AddNewMethod(gs2, b)
			} else {
				log.Printf("//ERR// Cannot add new method to %T\n", owner)
			}
		case *JSimpleStatement:
			ss := analyzeSimpleStatement(gs2, owner, b)
			if ss != nil {
				stmts = append(stmts, ss)
			}
		case *JSynchronized:
			ss := analyzeSynchronized(gs2, owner, b)
			if ss != nil {
				stmts = append(stmts, ss)
			}
		case *JSwitch:
			swtch := analyzeSwitch(gs2, owner, b)
			if swtch != nil {
				stmts = append(stmts, swtch)
			}
		case *JTry:
			try := analyzeTry(gs2, owner, b)
			if try != nil {
				stmts = append(stmts, try)
			}
		case *JUnimplemented:
			log.Printf("//ERR// Ignoring unimplemented block %s\n", b.typestr)
		case *JWhile:
			while := analyzeWhile(gs2, owner, b)
			if while != nil {
				stmts = append(stmts, while)
			}
		case nil:
			log.Printf("//ERR// Ignoring nil block entry\n")
		default:
			reportCastError("Block", bobj)
		}
	}

	return &GoBlock{stmts: stmts}
}

func analyzeBranchStmt(gs *GoState, owner GoMethodOwner, tok token.Token,
	obj JObject) *GoBranchStmt {
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
	cex *JCastExpr) *GoCastType {
	td := gs.Program().createTypeData(cex.reftype.name,
		cex.reftype.type_args, cex.reftype.dims)
	return &GoCastType{target: analyzeExpr(gs, owner, cex.target), casttype: td}
}

func analyzeClassBody(gs *GoState, cls *GoClassDefinition, body *JClassBody) {
	for _, bobj := range body.list {
		switch b := bobj.(type) {
		case *JClassDecl:
			gs.addClassDecl(cls, b)
		case *JEnumDecl:
			gs.Program().addEnum(b)
		case *JInterfaceDecl:
			log.Printf("//ERR// Ignoring class body %T\n", b)
/*
			idecl := analyzeInterfaceDeclaration(gs, b)
			if idecl != nil {
				decls = append(decls, idecl)
			}
*/
		case *JMethodDecl:
			cls.AddNewMethod(gs, b)
		case *JUnimplemented:
			log.Printf("//ERR// Ignoring unimplemented body %s\n", b.typestr)
		case *JVariableDecl:
			govar := gs.addVariableDecl(b, true)

			if cls.vars == nil {
				cls.vars = make([]*GoVarInit, 0)
			}

			var goinit *GoVarInit
			if b.init == nil {
				goinit = &GoVarInit{govar: govar}
			} else {
				goinit = analyzeVariableInit(gs, cls, b.init, govar)
			}
			cls.vars = append(cls.vars, goinit)
		default:
			reportCastError("Body.ClassDecl", bobj)
		}
	}
}

func analyzeConstant(gs *GoState, owner GoMethodOwner, jcon *JConstantDecl) {
	ctype := gs.Program().createTypeData(jcon.typespec.name,
		jcon.typespec.type_args, jcon.dims)
	init := analyzeVariableInit(gs, owner, jcon.init, nil)
	owner.AddConstant(&GoConstant{name: jcon.name, typedata: ctype, init: init})
}

func analyzeExpr(gs *GoState, owner GoMethodOwner, obj JObject) GoExpr {
	switch e := obj.(type) {
	case *JArrayAlloc:
		return analyzeArrayAlloc(gs, owner, e, nil)
	case *JAssignmentExpr:
		return analyzeAssignExpr(gs, owner, e)
	case *JBinaryExpr:
		return analyzeBinaryExpr(gs, owner, e)
	case *JCastExpr:
		return analyzeCastExpr(gs, owner, e)
	case *JClassAllocationExpr:
		return analyzeAllocationExpr(gs, owner, e)
	case *JConditionalExpr:
		if gs.Program().verbose {
			log.Printf("//ERR// Not converting condexpr %T (%T|?|%T|:|%T) to Expr\n",
				e, e.condexpr, e.ifexpr, e.elseexpr)
		} else {
			log.Printf("//ERR// Not converting condexpr to Expr\n")
		}
		return &GoUnimplemented{fname: "expr", text: fmt.Sprintf("%T", e)}
	case *JInstanceOf:
		return &GoInstanceOf{expr: analyzeExpr(gs, owner, e.obj),
			vartype: analyzeReferenceType(gs, e.typespec)}
	case *GoKeyword:
		return e
	case *GoLiteral:
		return e
	case *JMethodAccess:
		return analyzeMethodAccess(gs, owner, e)
	case *JNameDotObject:
		return analyzeNameDotObject(gs, e)
	case *JArrayReference:
		return analyzeArrayReference(gs, owner, e)
	case *JObjectDotName:
		return analyzeObjectDotName(gs, owner, e)
	case *JParens:
		return analyzeExpr(gs, owner, e.expr)
	case *GoPrimitiveType:
		return e
	case *JReferenceType:
		return analyzeReferenceType(gs, e)
	case *JUnaryExpr:
		return analyzeUnaryExpr(gs, owner, e)
	case *JUnimplemented:
		log.Printf("//ERR// Not analyzing unimplemented expr %s\n", e.typestr)
		return &GoUnimplemented{fname: "expr",
			text: fmt.Sprintf("%s", e.typestr)}
	case *JVariableInit:
		return analyzeVariableInit(gs, owner, e, nil)
	}

	panic(fmt.Sprintf("Unknown expression type %T", obj))
}

func analyzeForColon(gs *GoState, owner GoMethodOwner,
	jfc *JForColon) *GoForColon {
	gs2 := NewGoState(gs)

	govar := gs2.addVariableDecl(jfc.vardecl, false)
	if govar == nil {
		panic("ForColon addVariable returned nil")
	}

	var expr GoExpr
	if jfc.expr != nil {
		expr = analyzeExpr(gs2, owner, jfc.expr)
	}

	var body *GoBlock
	if jfc.body != nil {
		body = makeBlock(gs2, owner, jfc.body)
	}

	if body == nil {
		log.Printf("//ERR// adding empty forcolon body\n")
		list := make([]GoStatement, 0)
		body = &GoBlock{stmts: list}
	}

	return &GoForColon{govar: govar, expr: expr, body: body}
}

func analyzeForExpr(gs *GoState, owner GoMethodOwner,
	jfor *JForExpr) *GoForExpr {
	gs2 := NewGoState(gs)

	fe := &GoForExpr{}

	if jfor.init != nil && len(jfor.init) > 0 {
		fe.init = make([]GoExpr, len(jfor.init))
		for i, v := range jfor.init {
			fe.init[i] = analyzeExpr(gs2, owner, v)
		}
	}

	if jfor.expr != nil {
		fe.cond = analyzeExpr(gs2, owner, jfor.expr)
	}

	if jfor.incr != nil && len(jfor.incr) > 0 {
		fe.incr = make([]GoExpr, len(jfor.incr))
		for i, v := range jfor.incr {
			fe.incr[i] = analyzeExpr(gs2, owner, v)
		}
	}

	if jfor.body != nil {
		fe.block = makeBlock(gs2, owner, jfor.body)
	}

	return fe
}

func analyzeForVar(gs *GoState, owner GoMethodOwner, jfv *JForVar) *GoForVar {
	gs2 := NewGoState(gs)

	forvar := &GoForVar{govar: gs2.addVariableDecl(jfv.vardecl, false)}

	if jfv.vardecl.init != nil {
		forvar.init = analyzeVariableInit(gs2, owner, jfv.vardecl.init,
			forvar.govar)
	}

	if jfv.decl != nil {
		log.Printf("//ERR// ignoring forvar decl %T\n", jfv.decl)
	}

	//multiple initialization; a consolidated bool expression with && and ||; multiple 'incrementation'
	// for i, j, s := 0, 5, "a"; i < 3 && j < 100 && s != "aaaaa"; i, j, s = i+1, j+1, s + "a"  {

	if jfv.expr != nil {
		forvar.cond = analyzeExpr(gs2, owner, jfv.expr)
	}

	if jfv.incr != nil && len(jfv.incr) > 0 {
		forvar.incr = make([]GoStatement, 0)
		for _, v := range jfv.incr {
			stmts := analyzeStmt(gs2, owner, v)
			if stmts != nil && len(stmts) > 0 {
				forvar.incr = append(forvar.incr, stmts...)
			}
		}
	}

	if jfv.body != nil {
		forvar.block = makeBlock(gs2, owner, jfv.body)
	}

	return forvar
}

func analyzeIfElseStmt(gs *GoState, owner GoMethodOwner,
	ifelse *JIfElseStmt) *GoIfElse {

	stmts := analyzeStmt(gs, owner, ifelse.ifblock)
	if stmts == nil || len(stmts) == 0 {
		panic("IfStmt body cannot be nil")
	}

	var ifblk GoStatement
	if len(stmts) == 1 {
		ifblk = stmts[0]
	} else {
		ifblk = &GoBlock{stmts: stmts}
	}

	ifstmt := &GoIfElse{cond: analyzeExpr(gs, owner, ifelse.cond),
		ifblk: ifblk}

	if ifelse.elseblock != nil {
		stmts = analyzeStmt(gs, owner, ifelse.elseblock)
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
	ls *JLabeledStatement) *GoLabeledStmt {
	stmts := analyzeStmt(gs, owner, ls.stmt)
	if len(stmts) != 1 {
		panic("Found label assigned to multiple statements")
	}

	return &GoLabeledStmt{label: ls.label, stmt: stmts[0]}
}

func analyzeLocalVariableDeclaration(gs *GoState, owner GoMethodOwner,
	vdec *JLocalVariableDecl) []GoStatement {
	if vdec.vars == nil {
		return nil
	}

	stmts := make([]GoStatement, 0)
	for i, lvar := range vdec.vars {
		lv := analyzeLocalVariableInternal(gs, owner, vdec.modifiers,
			vdec.typespec, lvar, i)
		stmts = append(stmts, lv)
	}

	return stmts
}

func analyzeLocalVariableInternal(gs *GoState, owner GoMethodOwner,
	defaultModifiers *JModifiers, defaultTypespec *JReferenceType,
	vardec *JVariableDecl, idx int) GoStatement {

	var typespec *JReferenceType
	if vardec.typespec != nil {
		typespec = vardec.typespec
	} else {
		typespec = defaultTypespec
	}

	var mods *JModifiers
	if vardec.modifiers != nil {
		mods = vardec.modifiers
	} else {
		mods = defaultModifiers
	}

	govar := gs.addVariable(vardec.name, mods, vardec.dims, typespec, false)
	if govar == nil {
		panic("addVariable returned nil")
	}

	if vardec.init == nil {
		return NewGoLocalVarNoInit(govar)
	}

	if vardec.init.arraylist != nil {
		init := analyzeVariableInit(gs, owner, vardec.init, govar)
		return NewGoLocalVarInit(govar, init)
	}

	cex, ok := vardec.init.expr.(*JCastExpr)
	if !ok {
		init := analyzeExpr(gs, owner, vardec.init.expr)
		return NewGoLocalVarInit(govar, init)
	}

	return NewGoLocalVarCast(govar, analyzeCastExpr(gs, owner, cex))
}

func analyzeMethodAccess(gs *GoState, owner GoMethodOwner,
	mth *JMethodAccess) GoExpr {
	arglist := NewGoMethodArguments(gs, owner, mth.arglist)

	if mth.namekey != nil {
		is_super := mth.namekey.token == SUPER
		if !is_super && mth.namekey.token != THIS {
			panic(fmt.Sprintf("Keyword %v is neither THIS nor SUPER",
				mth.namekey))
		}

		return &GoMethodAccessKeyword{is_super: is_super, args: arglist}
	}

	if mth.method == "" {
		panic("JMethodAccess name is nil")
	}

	var expr GoExpr
	if mth.nameobj != nil {
		expr = analyzeExpr(gs, owner, mth.nameobj)

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

		mthd := findMethod(owner, class, mth.method, arglist,
			gs.Program().verbose)

		return &GoMethodAccessExpr{expr: expr, method: mthd, args: arglist}
	}

	var class GoMethodOwner
	var govar GoVar
	if mth.nametyp == nil {
		class = nilMethodOwner
	} else {
		govar = gs.findVariable(mth.nametyp)
		if govar != nil {
			class = owner
		} else {
			class = gs.findClass(owner, mth.nametyp.LastType())
			if class == nil {
				class = &GoFakeClass{name: mth.nametyp.String()}
			}
		}
	}

	mthd := findMethod(owner, class, mth.method, arglist, gs.Program().verbose)

	if govar != nil {
		return &GoMethodAccessVar{govar: govar, method: mthd, args: arglist}
	}

	return &GoMethodAccess{method: mthd, args: arglist}
}

func analyzeNameDotObject(gs *GoState, ndo *JNameDotObject) GoExpr {
	switch o := ndo.obj.(type) {
	case *GoKeyword:
		log.Printf("//ERR// Not converting ndoobj %T (kwd %s)\n", ndo.obj, o.name)
	default:
		log.Printf("//ERR// Not converting ndoobj %T (%T) to Expr\n", ndo, ndo.obj)
	}

	return &GoUnimplemented{fname: "ndo", text: fmt.Sprintf("%T", ndo.obj)}
}

func analyzeArrayReference(gs *GoState, owner GoMethodOwner,
	nae *JArrayReference) *GoArrayReference {
	var govar GoVar
	var obj GoExpr
	if nae.name != nil {
		govar = gs.findOrFakeVariable(nae.name, "arrayref")
	} else {
		obj = analyzeExpr(gs, owner, nae.obj)
	}

	return &GoArrayReference{govar: govar, obj: obj,
		index: analyzeExpr(gs, owner, nae.expr)}
}

func analyzeObjectDotName(gs *GoState, owner GoMethodOwner,
	odn *JObjectDotName) GoVar {
	switch o := odn.obj.(type) {
	case *GoKeyword:
		if o.token == THIS && gs.Receiver() != "" {
			rvar := gs.findVariable(NewJTypeName(gs.Receiver(), false))
			if rvar == nil {
				rvar = gs.addVariable(gs.Receiver(), nil, 0, nil, false)
			}

			ref := gs.findVariable(odn.name)
			if ref == nil {
				ref = gs.addVariable(gs.Receiver(), nil, 0, nil, false)
			}

			return NewGoSelector(rvar, ref)
		} else if o.token == SUPER && gs.Receiver() != "" {
			log.Printf("//ERR// Not converting odnobj super\n")
			return NewFakeVar("<<super>>", nil, 0)
		} else {
			log.Printf("//ERR// Not converting odnobj %T (kwd %s)\n",
				odn.obj, o.name)
			return NewFakeVar(fmt.Sprintf("<<%v>>", o.name), nil, 0)
		}
	default:
		return NewObjectDotName(odn, analyzeExpr(gs, owner, odn.obj), gs)
	}
}

func analyzeReferenceType(gs *GoState, ref *JReferenceType) GoVar {
	if ref.type_args != nil && len(ref.type_args) > 0 {
		fmt.Sprintf("//ERR// Not handling reftype type_args in %v\n", ref.name)
	}

	govar := gs.findVariable(ref.name)
	if govar != nil {
		return govar
	}

	return NewFakeVar(ref.name.String(), ref.type_args, ref.dims)
}

func analyzeSimpleStatement(gs *GoState, owner GoMethodOwner,
	jstmt *JSimpleStatement) GoStatement {

	if jstmt.keyword != nil {
		switch jstmt.keyword.token {
		case BREAK:
			return analyzeBranchStmt(gs, owner, token.BREAK, jstmt.object)
		case CONTINUE:
			return analyzeBranchStmt(gs, owner, token.CONTINUE, jstmt.object)
		case RETURN:
			var expr GoExpr
			if jstmt.object != nil {
				expr = analyzeExpr(gs, owner, jstmt.object)
			}

			return &GoReturn{expr: expr}
		case THROW:
			return &GoThrow{expr: analyzeExpr(gs, owner, jstmt.object)}
		default:
			return &GoUnimplemented{fname: "simpstmt",
				text: jstmt.keyword.name}
		}
	}

	switch expr := jstmt.object.(type) {
	case *JAssignmentExpr:
		return analyzeAssignExpr(gs, owner, expr)
	case *JClassAllocationExpr:
		return &GoExprStmt{x: analyzeAllocationExpr(gs, owner, expr)}
	case *JMethodAccess:
		return &GoExprStmt{x: analyzeMethodAccess(gs, owner, expr)}
	case *JUnaryExpr:
		return analyzeUnaryExpr(gs, owner, expr)
	default:
		log.Printf("//ERR// -------- not analyzing simpstmt %T\n", jstmt.object)
		return &GoUnimplemented{fname: "simpstmt",
			text: fmt.Sprintf("%T", jstmt.object)}
	}
}

func analyzeStmt(gs *GoState, owner GoMethodOwner,
	jstmt JObject) []GoStatement {
	switch stmt := jstmt.(type) {
	case *JAssignmentExpr:
		return []GoStatement{ analyzeAssignExpr(gs, owner, stmt), }
	case *JBlock:
		return []GoStatement{ analyzeBlock(gs, owner, stmt), }
	case *JForColon:
		return []GoStatement{ analyzeForColon(gs, owner, stmt), }
	case *JForVar:
		return []GoStatement{ analyzeForVar(gs, owner, stmt), }
	case *JIfElseStmt:
		return []GoStatement{ analyzeIfElseStmt(gs, owner, stmt), }
	case *GoJumpToLabel:
		return []GoStatement{ stmt, }
	case *JLocalVariableDecl:
		return analyzeLocalVariableDeclaration(gs, owner, stmt)
	case *JSimpleStatement:
		return []GoStatement{ analyzeSimpleStatement(gs, owner, stmt), }
	case *JTry:
		return []GoStatement{ analyzeTry(gs, owner, stmt), }
	case *JUnaryExpr:
		return []GoStatement{ analyzeUnaryExpr(gs, owner, stmt), }
	case *JUnimplemented:
		log.Printf("//ERR// Not analyzing unimplemented stmt %s\n", stmt.typestr)
		return []GoStatement{ &GoUnimplemented{fname: "stmt",
			text: stmt.typestr}, }
	case *JWhile:
		return []GoStatement{ analyzeWhile(gs, owner, stmt), }
	default:
		log.Printf("//ERR// Not analyzing stmt %T\n", stmt)
		return []GoStatement{ &GoUnimplemented{fname: "stmt",
			text: fmt.Sprintf("%T", stmt)}, }
	}
}

func analyzeSwitch(gs *GoState, owner GoMethodOwner,
	jsw *JSwitch) *GoSwitch {
	if jsw.groups == nil || len(jsw.groups) == 0 {
		// ignore empty switch statements
		return nil
	}

	gsw := &GoSwitch{expr: analyzeExpr(gs, owner, jsw.expr),
		cases: make([]*GoSwitchCase, len(jsw.groups))}
	for i, c := range jsw.groups {
		gsw.cases[i] = analyzeSwitchCase(gs, owner, c)
	}

	return gsw
}

func analyzeSwitchCase(gs *GoState, owner GoMethodOwner,
	jsg *JSwitchGroup) *GoSwitchCase {
	if jsg.labels == nil || len(jsg.labels) == 0 {
		panic("No labels for switch case")
	}

	labels := make([]*GoSwitchLabel, len(jsg.labels))
	for i, l := range jsg.labels {
		labels[i] = analyzeSwitchLabel(gs, owner, l)
	}

	stmts := make([]GoStatement, 0)
	for _, s := range jsg.stmts {
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
	jsl *JSwitchLabel) *GoSwitchLabel {
	if jsl.is_default {
		return &GoSwitchLabel{is_default: true}
	}

	var expr GoExpr
	if jsl.name != "" {
		expr = &GoLiteral{text: jsl.name}
	} else if  jsl.expr != nil {
		expr = analyzeExpr(gs, owner, jsl.expr)
	} else {
		panic("Empty switch label")
	}

	return &GoSwitchLabel{expr: expr}
}

func analyzeSynchronized(gs *GoState, owner GoMethodOwner,
	sync *JSynchronized) *GoSynchronized {
	return &GoSynchronized{expr: analyzeExpr(gs, owner, sync.expr),
		block: analyzeBlock(gs, owner, sync.block)}
}

func analyzeTry(gs *GoState, owner GoMethodOwner, try *JTry) *GoTry {
	gt := &GoTry{block: analyzeBlock(NewGoState(gs), owner, try.block)}

	if try.catches != nil && len(try.catches) > 0 {
		gs2 := NewGoState(gs)

		gt.catches = make([]*GoTryCatch, len(try.catches))
		for i, c := range try.catches {
			var exc *JTypeName
			if len(c.typelist) == 1 {
				exc = c.typelist[0]
			} else {
				exc = NewJTypeName("Exception", false)
			}

			govar := gs.addVariable(c.name, c.modifiers, 0,
				NewJReferenceType(exc, nil, 0), false)

			if len(c.typelist) != 1 {
				log.Printf("//ERR// Ignoring catch#%d long typelist (len=%d)\n",
					i, len(c.typelist))
			}

			gt.catches[i] = &GoTryCatch{govar: govar,
				block: analyzeBlock(gs2, owner, c.block)}
		}
	}

	if try.finally != nil {
		gt.finally = analyzeBlock(NewGoState(gs), owner, try.finally)
	}

	return gt
}

func analyzeUnaryExpr(gs *GoState, owner GoMethodOwner, uexpr *JUnaryExpr) *GoUnaryExpr {
	var op token.Token
	switch uexpr.op {
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
		panic(fmt.Sprintf("Unknown unary operator \"%s\"", uexpr.op))
	}

	return &GoUnaryExpr{op: op, x: analyzeExpr(gs, owner, uexpr.obj)}
}

func analyzeVariableInit(gs *GoState, owner GoMethodOwner,
	init *JVariableInit, govar GoVar) *GoVarInit {
	if init.expr != nil {
		var expr GoExpr
		switch v := init.expr.(type) {
		case *JVariableInit:
			expr = analyzeVariableInit(gs, owner, v, govar)
		case *JArrayAlloc:
			expr = analyzeArrayAlloc(gs, owner, v, govar)
		default:
			expr = analyzeExpr(gs, owner, init.expr)
		}

		return &GoVarInit{govar: govar, expr: expr}
	}

	if init.arraylist == nil {
		panic("No variable Initialization")
	}

	elements := make([]GoExpr, len(init.arraylist))
	for i, elem := range init.arraylist {
		elements[i] = analyzeVariableInit(gs, owner, elem, govar)
	}

	return &GoVarInit{govar: govar, elements: elements}
}

func analyzeWhile(gs *GoState, owner GoMethodOwner, while *JWhile) *GoWhile {
	gw := &GoWhile{expr: analyzeExpr(gs, owner, while.expr),
		is_do_while: while.is_do_while}

	stmts := analyzeStmt(gs, owner, while.stmt)
	if stmts != nil && len(stmts) > 0 {
		if len(stmts) == 1 {
			gw.stmt = stmts[0]
		} else {
			gw.stmt = &GoBlock{stmts: stmts}
		}
	}

	return gw
}

func makeBlock(gs *GoState, owner GoMethodOwner, block JObject) *GoBlock {
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
