package grammar

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	//"runtime/debug"  // debug.PrintStack()
	"strings"
)

const indentStep = "    "

func isNil(j JObject) bool {
	return j == nil || !reflect.ValueOf(j).Elem().IsValid()
}

type JObject interface {
}

type JUnimplemented struct {
	TypeStr string
}

func NewJUnimplemented(typestr string) *JUnimplemented {
	return &JUnimplemented{TypeStr: typestr}
}

type JAnnotation struct {
	name *JTypeName
	elem []JObject
	has_parens bool
}

func NewJAnnotation(name *JTypeName, elem []JObject,
	has_parens bool) *JAnnotation {
	if name == nil {
		ReportError("Annotation name cannot be nil")
	}

	return &JAnnotation{name: name, elem: elem, has_parens: has_parens}
}

type JArrayAlloc struct {
	Typename *JTypeName
	Dimexprs []JObject
	Dims int
	Init []*JVariableInit
}

func NewJArrayAlloc(typename *JTypeName, dimexprs []JObject,
	dims int) *JArrayAlloc {
	if typename == nil {
		ReportError("ArrayExpr name cannot be nil")
	}

	return &JArrayAlloc{Typename: typename, Dimexprs: dimexprs, Dims: dims}
}

func (j *JArrayAlloc) SetInitializers(init []*JVariableInit) {
	j.Init = init
}

type JArrayReference struct {
	Name *JTypeName
	Obj JObject
	Expr JObject
}

func NewJArrayReference(name *JTypeName, obj JObject, expr JObject) *JArrayReference {
	if name == nil && obj == nil {
		ReportError("NewArrayExpr name or obj must be specified")
	} else if name != nil && obj != nil {
		ReportError("NewArrayExpr name and obj cannot both be specified")
	} else if expr == nil {
		ReportError("NewArrayExpr expr cannot be nil")
	}

	return &JArrayReference{Name: name, Obj: obj, Expr: expr}
}

type JAssignmentExpr struct {
	Left JObject
	Op string
	Right JObject
}

func NewJAssignmentExpr(left JObject, op string, right JObject) *JAssignmentExpr {
	if left == nil {
		ReportError("JAssignmentExpr left expression cannot be nil")
	} else if right == nil {
		ReportError("JAssignmentExpr right expression cannot be nil")
	}

	return &JAssignmentExpr{Left: left, Op: op, Right: right}
}

type JBinaryExpr struct {
	Obj1 JObject
	Op string
	Obj2 JObject
}

func NewJBinaryExpr(obj1 JObject, op string, obj2 JObject) *JBinaryExpr {
	if obj1 == nil {
		ReportError("BinaryExpr object1 cannot be nil")
	} else if op == "" {
		ReportError("BinaryExpr op cannot be empty")
	} else if obj2 == nil {
		ReportError("BinaryExpr object2 cannot be nil")
	}

	return &JBinaryExpr{Obj1: obj1, Op: op, Obj2: obj2}
}

type JBlock struct {
	static bool
	List []JObject
}

func NewJBlock(list []JObject) *JBlock {
	return &JBlock{List: list}
}

func (j *JBlock) SetStatic() { j.static = true }

type JCastExpr struct {
	Reftype *JReferenceType
	Target JObject
}

func NewJCastExpr(reftype *JReferenceType, target JObject) *JCastExpr {
	if reftype == nil {
		ReportError("CastExpr reference type cannot be nil")
	} else if target == nil {
		ReportError("CastExpr target cannot be nil")
	}

	return &JCastExpr{Reftype: reftype, Target: target}
}

type JCatch struct {
	Modifiers *JModifiers
	TypeList []*JTypeName
	Name string
	Block *JBlock
}

func NewJCatch(modifiers *JModifiers, typelist []*JTypeName, name string,
	block *JBlock) *JCatch {
	if typelist == nil || len(typelist) == 0 {
		panic("List of exception types cannot be nil/empty")
	}

	return &JCatch{Modifiers: modifiers, TypeList: typelist, Name: name,
		Block: block}
}

type JClassAllocationExpr struct {
	Name *JTypeName
	TypeArgs []*JTypeArgument
	Arglist []JObject
	Body []JObject
}

func NewJClassAllocationExpr(name *JTypeName, obj_args []JObject,
	arglist []JObject) *JClassAllocationExpr {
	if name == nil {
		ReportError("Class name cannot be nil")
	}

	var type_args []*JTypeArgument
	if obj_args != nil && len(obj_args) > 0 {
		type_args = make([]*JTypeArgument, len(obj_args))
		for i, o := range obj_args {
			if arg, ok := o.(*JTypeArgument); ok {
				type_args[i] = arg
			} else {
				ReportCastError("JTypeArgument", o)
			}
		}
	}

	return &JClassAllocationExpr{Name: name, TypeArgs: type_args,
		Arglist: arglist}
}

func (j *JClassAllocationExpr) SetBody(body []JObject) {
	j.Body = body
}

type JClassBody struct {
	List []JObject
}

func NewJClassBody(list []JObject) *JClassBody {
	if list == nil || len(list) == 0 {
		panic("JClassBody list cannot be empty/nil")
	}

	return &JClassBody{List: list}
}

type JClassDecl struct {
	modifiers *JModifiers
	Name string
	type_params []JObject
	Extends *JReferenceType
	Interfaces []*JTypeName
	Body []JObject
}

func NewJClassDecl(modifiers *JModifiers, name string, type_params []JObject,
	extends *JReferenceType, interfaces []*JTypeName, body []JObject) *JClassDecl {
	if name == "" {
		ReportError("JClassDecl name cannot be empty")
	}

	return &JClassDecl{modifiers: modifiers, Name: name,
		type_params: type_params, Extends: extends, Interfaces: interfaces,
		Body: body}
}

type JConditionalExpr struct {
	CondExpr JObject
	IfExpr JObject
	ElseExpr JObject
}

func NewJConditionalExpr(condexpr JObject, ifexpr JObject, elseexpr JObject) *JConditionalExpr {
	return &JConditionalExpr{CondExpr: condexpr, IfExpr: ifexpr,
		ElseExpr: elseexpr}
}

type JConstantDecl struct {
	modifiers *JModifiers
	TypeSpec *JReferenceType
	Name string
	Dims int
	Init *JVariableInit
}

func NewJConstantDecl(name string, dims int,
	init *JVariableInit) *JConstantDecl {
	return &JConstantDecl{Name: name, Dims: dims, Init: init}
}

func (j *JConstantDecl) HasName() bool {
	return j.Name != ""
}

func (j *JConstantDecl) SetModifiers(modifiers *JModifiers) {
       j.modifiers = modifiers
}

func (j *JConstantDecl) SetName(name string) {
	if name == "" {
		panic("Constant name cannot be empty")
	} else if j.Name != "" {
		panic(fmt.Sprintf("Cannot overwrite constant name \"%s\" with \"%s\"",
			j.Name, name))
	}

	j.Name = name
}

func (j *JConstantDecl) SetType(typespec *JReferenceType) {
	j.TypeSpec = typespec
}

type JElementValuePair struct {
	name string
	value JObject
}

func NewJElementValuePair(name string, value JObject) *JElementValuePair {
	if name == "" {
		ReportError("ReferenceType name cannot be empty")
	}

	return &JElementValuePair{name: name, value: value}
}

type JEmpty struct {
}

func NewJEmpty() *JEmpty {
	return &JEmpty{}
}

type JEnumBody struct {
	constants []*JEnumConstant
	bodydecl []JObject
}

func NewJEnumBody(clist []JObject, bodydecl []JObject) *JEnumBody {
	var constants []*JEnumConstant
	if clist != nil && len(clist) > 0 {
		constants = make([]*JEnumConstant, len(clist))
		for i, v := range clist {
			if jcon, ok := v.(*JEnumConstant); !ok {
				ReportCastError("JEnumConstant", v)
			} else {
				constants[i] = jcon
			}
		}
	}

	return &JEnumBody{constants: constants, bodydecl: bodydecl}
}

type JEnumConstant struct {
	Annotations []*JAnnotation
	Name string
	ArgList []JObject
	Body []JObject
}

func NewJEnumConstant(alist []JObject, name string, arglist []JObject,
	body []JObject) *JEnumConstant {
	var annotations []*JAnnotation
	if alist != nil && len(alist) > 0 {
		annotations := make([]*JAnnotation, len(alist))
		for i, v := range alist {
			if jann, ok := v.(*JAnnotation); !ok {
				ReportCastError("JAnnotation", v)
			} else {
				annotations[i] = jann
			}
		}
	}

	return &JEnumConstant{Annotations: annotations, Name: name,
		ArgList: arglist, Body: body}
}

type JEnumDecl struct {
	modifiers *JModifiers
	Name string
	Interfaces []*JTypeName
	Constants []*JEnumConstant
	BodyDecl []JObject
}

func NewJEnumDecl(modifiers *JModifiers, name string, interfaces []*JTypeName,
	body *JEnumBody) *JEnumDecl {
	if name == "" {
		ReportError("JEnumDecl name cannot be empty")
	}

	var constants []*JEnumConstant
	var bodydecl []JObject
	if body != nil {
		constants = body.constants
		bodydecl = body.bodydecl
	}

	return &JEnumDecl{modifiers: modifiers, Name: name, Interfaces: interfaces,
		Constants: constants, BodyDecl: bodydecl}
}

type JForColon struct {
	VarDecl *JVariableDecl
	Expr JObject
	Body JObject
}

func NewJForColon(modifiers *JModifiers, typespec *JReferenceType,
	name string, dims int, expr JObject) *JForColon {
	if typespec == nil {
		panic("JForColon type specifier cannot be nil")
	} else if name == "" {
		panic("JForColon variable name cannot be empty")
	} else if expr == nil {
		panic("JForColon expr cannot be nil")
	}

	vardecl := NewJVariableDecl(name, dims, nil)
	vardecl.SetModifiers(modifiers)
	vardecl.SetType(typespec)

	return &JForColon{VarDecl: vardecl, Expr: expr}
}

func (j *JForColon) SetBody(body JObject) {
	j.Body = body
}

type JForExpr struct {
	Init []JObject
	Expr JObject
	Incr []JObject
	Body JObject
}

func NewJForExpr(init []JObject, expr JObject, incr []JObject) *JForExpr {
	return &JForExpr{Init: init, Expr: expr, Incr: incr}
}

func (j *JForExpr) SetBody(body JObject) {
	j.Body = body
}

type JForVar struct {
	VarDecl *JVariableDecl
	Decl JObject
	Expr JObject
	Incr []JObject
	Body JObject
}

func NewJForVar(modifiers *JModifiers, typespec *JReferenceType,
	name string, dims int) *JForVar {
	if typespec == nil {
		panic("JForVar type specifier cannot be nil")
	} else if name == "" {
		panic("JForVar variable name cannot be empty")
	}

	vardecl := NewJVariableDecl(name, dims, nil)
	if modifiers != nil {
		vardecl.SetModifiers(modifiers)
	}
	vardecl.SetType(typespec)

	return &JForVar{VarDecl: vardecl}
}

func (j *JForVar) SetBody(body JObject) {
	j.Body = body
}

func (j *JForVar) SetDecl(decl JObject) {
	j.Decl = decl
}

func (j *JForVar) SetExpr(expr JObject) {
	j.Expr = expr
}

func (j *JForVar) SetIncr(incr []JObject) {
	j.Incr = incr
}

func (j *JForVar) SetInit(init *JVariableInit) {
	j.VarDecl.Init = init
}

type JFormalParameter struct {
	Modifiers *JModifiers
	TypeSpec *JReferenceType
	DotDotDot bool
	Name string
	Dims int
}

func NewJFormalParameter(typespec *JReferenceType, dotdotdot bool,
	name string, dims int) *JFormalParameter {
	if name == "" {
		panic("FormalParameter name cannot be empty")
	}

	return &JFormalParameter{TypeSpec: typespec, DotDotDot: dotdotdot,
		Name: name, Dims: dims}
}

func (j *JFormalParameter) SetModifiers(modifiers *JModifiers) {
       j.Modifiers = modifiers
}

type JIfElseStmt struct {
	Cond JObject
	IfBlock JObject
	ElseBlock JObject
}

func NewJIfElseStmt(cond JObject, ifblock JObject, elseblock JObject) *JIfElseStmt {
	if cond == nil {
		panic("'if' condition cannot be nil")
	} else if ifblock == nil {
		panic("'if' block cannot be nil")
	}

	return &JIfElseStmt{Cond: cond, IfBlock: ifblock, ElseBlock: elseblock}
}

type JImportStmt struct {
	Name *JTypeName
	is_wild bool
	is_static bool
}

func NewJImportStmt(name *JTypeName, is_wild bool, is_static bool) *JImportStmt {
	return &JImportStmt{Name: name, is_wild: is_wild, is_static: is_static}
}

type JInstanceOf struct {
	Obj JObject
	TypeSpec *JReferenceType
}

func NewJInstanceOf(obj JObject, typespec *JReferenceType) *JInstanceOf {
	if obj == nil {
		ReportError("Instanceof object cannot be nil")
	} else if typespec == nil {
		ReportError("TwoObjects type specifier cannot be nil")
	}

	return &JInstanceOf{Obj: obj, TypeSpec: typespec}
}

type JInterfaceDecl struct {
	modifiers *JModifiers
	Name *JTypeName
	type_params []JObject
	extends []*JTypeName
	Body []JObject
}

func NewJInterfaceDecl(modifiers *JModifiers, name *JTypeName,
	type_params []JObject, extends []*JTypeName,
	body []JObject) *JInterfaceDecl {
	if name == nil {
		ReportError("JInterfaceDecl name cannot be nil")
	}

	return &JInterfaceDecl{modifiers: modifiers, Name: name,
		type_params: type_params, extends: extends, Body: body}
}

type JInterfaceMethodDecl struct {
	modifiers *JModifiers
	type_params []JObject
	TypeSpec *JReferenceType
	Name string
	FormalParams []*JFormalParameter
	dims int
	throws []*JTypeName
}

func NewJInterfaceMethodDecl(formal_params []*JFormalParameter, dims int,
	throws []*JTypeName) *JInterfaceMethodDecl {
	return &JInterfaceMethodDecl{FormalParams: formal_params, dims: dims,
		throws: throws}
}

func (j *JInterfaceMethodDecl) SetModifiers(modifiers *JModifiers) {
       j.modifiers = modifiers
}

func (j *JInterfaceMethodDecl) SetName(name string) {
	if name == "" {
		panic("Interface name cannot be empty")
	} else if j.Name != "" {
		panic(fmt.Sprintf("Cannot overwrite interface name \"%s\" with \"%s\"",
			j.Name, name))
	}

	j.Name = name
}

func (j *JInterfaceMethodDecl) SetType(typespec *JReferenceType) {
	j.TypeSpec = typespec
}

func (j *JInterfaceMethodDecl) SetTypeParameters(type_params []JObject) {
	j.type_params = type_params
}

type JJumpToLabel struct {
	IsContinue bool
	Label string
}

func NewJJumpToLabel(token int, label string) *JJumpToLabel {
	var is_continue bool
	if token == CONTINUE {
		is_continue = true
	} else if token != BREAK {
		panic(fmt.Sprintf("JJumpToLabel token #%d is not BREAK(#%d)" +
			" or CONTINUE(#%d)", token, BREAK, CONTINUE))
	}

	return &JJumpToLabel{IsContinue: is_continue, Label: label}
}

type JKeyword struct {
	Token int
	Name string
}

func NewJKeyword(token int, name string) *JKeyword {
	if name == "" {
		ReportError("Keyword name cannot be empty")
	}

	return &JKeyword{Token: token, Name: name}
}

type JLabeledStatement struct {
	Label string
	Stmt JObject
}

func NewJLabeledStatement(label string, stmt JObject) *JLabeledStatement {
	if label == "" {
		ReportError("Label cannot be empty")
	} else if stmt == nil {
		ReportError("Label object cannot be nil")
	}

	return &JLabeledStatement{Label: label, Stmt: stmt}
}

type JLiteral struct {
	Text string
}

func NewJLiteral(text string) *JLiteral {
	if text == "" {
		ReportError("Literal text cannot be empty")
	}

	return &JLiteral{Text: text}
}

type JLocalVariableDecl struct {
	Modifiers *JModifiers
	TypeSpec *JReferenceType
	Vars []*JVariableDecl
}

func NewJLocalVariableDecl(modifiers *JModifiers, typespec *JReferenceType,
	vars []*JVariableDecl) *JLocalVariableDecl {
	return &JLocalVariableDecl{Modifiers: modifiers, TypeSpec: typespec,
		Vars: vars}
}

type JMethodAccess struct {
	NameObj JObject
	NameKey *JKeyword
	NameType *JTypeName
	Method string
	ArgList []JObject
}

func NewJMethodAccessComplex(obj JObject, name string,
	arglist []JObject) *JMethodAccess {
	if obj == nil {
		panic("Method object cannot be nil")
	} else if name == "" {
		panic("Method name cannot be empty")
	}

	return &JMethodAccess{NameObj: obj, Method: name, ArgList: arglist}
}

func NewJMethodAccessKeyword(token int, name string, arglist []JObject) *JMethodAccess {
	return &JMethodAccess{NameKey: NewJKeyword(token, name), ArgList: arglist}
}

func NewJMethodAccessName(nametyp *JTypeName,
	arglist []JObject) *JMethodAccess {
	if nametyp == nil {
		panic("Method name cannot be nil")
	}

	name := nametyp.LastType()
	if name == "" {
		panic("Method name cannot be empty")
	}

	var class *JTypeName
	if nametyp.IsDotted() {
		class = nametyp.NotLast()
	}

	return &JMethodAccess{NameType: class, Method: name, ArgList: arglist}
}

type JMethodDecl struct {
	Modifiers *JModifiers
	type_params []JObject
	TypeSpec *JReferenceType
	Name string
	FormalParams []*JFormalParameter
	dims int
	throws []*JTypeName
	Block *JBlock
}

func NewJMethodDecl(formal_params []*JFormalParameter, dims int,
	throws []*JTypeName, block *JBlock) *JMethodDecl {
	return &JMethodDecl{FormalParams: formal_params, dims: dims,
		throws: throws, Block: block}
}

func (j *JMethodDecl) SetModifiers(modifiers *JModifiers) {
       j.Modifiers = modifiers
}

func (j *JMethodDecl) SetName(name string) {
	if name == "" {
		panic("Method name cannot be empty")
	}

	j.Name = name
}

func (j *JMethodDecl) SetType(typespec *JReferenceType) {
	j.TypeSpec = typespec
}

func (j *JMethodDecl) SetTypeParameters(type_params []JObject) {
	j.type_params = type_params
}

const (
	ModPublic = 0x1
	modProtected = 0x2
	ModPrivate = 0x4
	modAbstract = 0x8
	ModFinal = 0x10
	ModStatic = 0x20
	modTransient = 0x40
	modVolatile = 0x80
	modNative = 0x100
	modSynchronized = 0x200
	modMax = modSynchronized
)

type JModifiers struct {
	annotations []*JAnnotation
	mod_bits int
}

func NewJModifiers(name string, annotation *JAnnotation) *JModifiers {
	j := &JModifiers{}
	if name != "" {
		j.AddModifier(name)
	} else if annotation != nil {
		j.AddAnnotation(annotation)
	}
	return j
}

func (j *JModifiers) AddAnnotation(annotation *JAnnotation) *JModifiers {
	if j.annotations == nil {
		j.annotations = make([]*JAnnotation, 1)
		j.annotations[0] = annotation
	} else {
		j.annotations = append(j.annotations, annotation)
	}
	return j
}

func (j *JModifiers) AddModifier(name string) *JModifiers {
	switch name {
	case "public": j.mod_bits |= ModPublic
	case "protected": j.mod_bits |= modProtected
	case "private": j.mod_bits |= ModPrivate
	case "abstract": j.mod_bits |= modAbstract
	case "final": j.mod_bits |= ModFinal
	case "static": j.mod_bits |= ModStatic
	case "transient": j.mod_bits |= modTransient
	case "volatile": j.mod_bits |= modVolatile
	case "native": j.mod_bits |= modNative
	case "synchronized": j.mod_bits |= modSynchronized
	default: ReportError(fmt.Sprintf("Unknown modifier \"%s\"", name))
	}

	return j
}

func (j *JModifiers) HasAnnotation(name string) bool {
	if j.annotations != nil && len(j.annotations) > 0 {
		for _, ann := range j.annotations {
			if ann.name.String() == name {
				return true
			}
		}
	}

	return false
}

func (j *JModifiers) IsSet(modbit int) bool {
	return (j.mod_bits & modbit) == modbit
}

func (j *JModifiers) String() string {
	b := &bytes.Buffer{}
	j.writeModifiers(b)
	return b.String()
}

func (j *JModifiers) writeModifiers(out io.Writer) {
	for m := 0x1; m <= modMax; m <<= 1 {
		if j.mod_bits & m == m {
			switch m {
			case ModPublic: io.WriteString(out, "public ")
			case modProtected: io.WriteString(out, "protected ")
			case ModPrivate: io.WriteString(out, "private ")
			case modAbstract: io.WriteString(out, "abstract ")
			case ModFinal: io.WriteString(out, "final ")
			case ModStatic: io.WriteString(out, "static ")
			case modTransient: io.WriteString(out, "transient ")
			case modVolatile: io.WriteString(out, "volatile ")
			case modNative: io.WriteString(out, "native ")
			case modSynchronized: io.WriteString(out, "synchronized ")
			}
		}
	}
}

type JNameDotObject struct {
	Name *JTypeName
	Obj JObject
}

func NewJNameDotObject(name *JTypeName, obj JObject) *JNameDotObject {
	if name == nil {
		ReportError("NameDotObject name cannot be nil")
	} else if obj == nil {
		ReportError("NameDotObject object cannot be nil")
	}

	return &JNameDotObject{Name: name, Obj: obj}
}

type JObjectDotName struct {
	Obj JObject
	Name *JTypeName
}

func NewJObjectDotName(obj JObject, name *JTypeName) *JObjectDotName {
	if obj == nil {
		ReportError("ObjectDotName object cannot be nil")
	} else if name == nil {
		ReportError("ObjectDotName name cannot be nil")
	}

	return &JObjectDotName{Obj: obj, Name: name}
}

type JPackageStmt struct {
	Name *JTypeName
}

func NewJPackageStmt(name *JTypeName) *JPackageStmt {
	return &JPackageStmt{Name: name}
}

type JParens struct {
	Expr JObject
}

func NewJParens(expr JObject) *JParens {
	if expr == nil {
		ReportError("JParens expression cannot be nil")
	}

	return &JParens{Expr: expr}
}

type JProgramFile struct {
	Pkg *JPackageStmt
	Imports []JObject
	TypeDecls []JObject
}

func NewJProgramFile(pobj JObject, imports []JObject,
	type_decls []JObject) *JProgramFile {
	j := &JProgramFile{}

	if pobj != nil {
		if pkg, ok := pobj.(*JPackageStmt); !ok {
			ReportCastError("JPackageStmt", pobj)
		} else {
			j.Pkg = pkg
		}
	}

	j.Imports = imports
	j.TypeDecls = type_decls

	return j
}

type JReferenceType struct {
	Name *JTypeName
	TypeArgs []*JTypeArgument
	Dims int
}

func NewJReferenceType(name *JTypeName, obj_args []JObject,
	dims int) *JReferenceType {
	if name == nil {
		ReportError("ReferenceType name cannot be nil")
	}

	var type_args []*JTypeArgument
	if obj_args != nil && len(obj_args) > 0 {
		type_args = make([]*JTypeArgument, len(obj_args))
		for i, o := range obj_args {
			if arg, ok := o.(*JTypeArgument); ok {
				type_args[i] = arg
			} else {
				ReportCastError("JTypeArgument", o)
			}
		}
	}

	return &JReferenceType{Name: name, TypeArgs: type_args, Dims: dims}
}

type JSimpleStatement struct {
	Keyword *JKeyword
	Object JObject
}

func NewJSimpleStatement(keyword *JKeyword, object JObject) *JSimpleStatement {
	return &JSimpleStatement{Keyword: keyword, Object: object}
}

type JSwitch struct {
	Expr JObject
	Groups []*JSwitchGroup
}

func NewJSwitch(expr JObject, grouplist []JObject) *JSwitch {
	if expr == nil {
		panic("No expression for JSwitch")
	}

	var groups []*JSwitchGroup
	if grouplist != nil && len(grouplist) > 0 {
		groups = make([]*JSwitchGroup, len(grouplist))
		for i, o := range grouplist {
			if grp, ok := o.(*JSwitchGroup); !ok {
				ReportCastError("JSwitch", o)
			} else {
				groups[i] = grp
			}
		}
	}

	return &JSwitch{Expr: expr, Groups: groups}
}

type JSwitchGroup struct {
	Labels []*JSwitchLabel
	Stmts []JObject
}

func NewJSwitchGroup(labellist []JObject, stmtlist []JObject) *JSwitchGroup {
	if labellist == nil || len(labellist) == 0 {
		panic("No labels on JSwitchGroup")
	}

	var labels []*JSwitchLabel
	if labellist != nil && len(labellist) > 0 {
		labels = make([]*JSwitchLabel, len(labellist))
		for i, o := range labellist {
			if lbl, ok := o.(*JSwitchLabel); !ok {
				ReportCastError("JSwitchLabel", o)
			} else {
				labels[i] = lbl
			}
		}
	}

	return &JSwitchGroup{Labels: labels, Stmts: stmtlist}
}

type JSwitchLabel struct {
	Name string
	Expr JObject
	IsDefault bool
}

func NewJSwitchLabel(name string, expr JObject, is_default bool) *JSwitchLabel {
	if !is_default {
		if name == "" && expr == nil {
			panic("Must set something for JSwitchLabel")
		} else if name != "" && expr != nil {
			panic("Cannot set both name and expr for JSwitchLabel")
		}
	} else if name != "" || expr != nil {
		panic("Cannot set name or expr for 'default' JSwitchLabel")
	}

	return &JSwitchLabel{Name: name, Expr: expr, IsDefault: is_default}
}

type JSynchronized struct {
	Expr JObject
	Block *JBlock
}

func NewJSynchronized(expr JObject, block *JBlock) *JSynchronized {
	if block == nil {
		ReportError("JSynchronized block cannot be nil")
	}

	return &JSynchronized{Expr: expr, Block: block}
}

type JTry struct {
	Block *JBlock
	Catches []*JCatch
	Finally *JBlock
}

func NewJTry(block *JBlock, clist []JObject, finally *JBlock) *JTry {
	if block == nil {
		ReportError("JTry block cannot be nil")
	}

	var catches []*JCatch
	if clist != nil && len(clist) > 0 {
		catches = make([]*JCatch, len(clist))
		for i, c := range clist {
			if catch, ok := c.(*JCatch); !ok {
				ReportCastError(fmt.Sprintf("JCatch#%d", i), c)
			} else {
				catches[i] = catch
			}
		}
	}

	return &JTry{Block: block, Catches: catches, Finally: finally}
}

const (
	TS_NONE = iota
	TS_EXTENDS
	TS_SUPER
	TS_PLAIN
)

type JTypeArgument struct {
	TypeSpec *JReferenceType
	Ts_type int
}

func NewJTypeArgument(typespec *JReferenceType, ts_type int) *JTypeArgument {
	return &JTypeArgument{TypeSpec: typespec, Ts_type: ts_type}
}

type JTypeName struct {
	is_primitive bool
	names []string
}

func NewJTypeName(name string, is_primitive bool) *JTypeName {
	if name == "" {
		panic("Cannot build type name from empty string")
	}

	names := make([]string, 1)
	names[0] = name

	return &JTypeName{names: names, is_primitive: is_primitive}
}

func (qn *JTypeName) Add(name string) {
	if qn.is_primitive {
		panic(fmt.Sprintf("Cannot add \"%s\" to primitive type \"%s\"",
			name, qn.names[0]))
	} else if name == "" {
		panic(fmt.Sprintf("Cannot add empty string to \"%s\"", qn.String()))
	}

	qn.names = append(qn.names, name)
}

func (qn *JTypeName) FirstType() string {
	return qn.names[0]
}

func (qn *JTypeName) IsDotted() bool {
	return !qn.is_primitive && len(qn.names) > 1
}

func (qn *JTypeName) IsPrimitive() bool {
	return qn.is_primitive
}

func (qn *JTypeName) LastType() string {
	return qn.names[len(qn.names)-1]
}

func (qn *JTypeName) NotFirst() *JTypeName {
	if len(qn.names) == 1 {
		panic(fmt.Sprintf("Only one name found in %v", qn))
	} else if qn.is_primitive {
		panic(fmt.Sprintf("Primitive type %v has only one name", qn))
	}

	return &JTypeName{names: qn.names[1:], is_primitive: false}
}

func (qn *JTypeName) NotLast() *JTypeName {
	if len(qn.names) == 1 {
		panic(fmt.Sprintf("Only one name found in %v", qn))
	} else if qn.is_primitive {
		panic(fmt.Sprintf("Primitive type %v has only one name", qn))
	}

	return &JTypeName{names: qn.names[0:len(qn.names)-1], is_primitive: false}
}

func (qn *JTypeName) PackageString() string {
	if len(qn.names) == 1 {
		return qn.names[0]
	} else if qn.is_primitive {
		panic(fmt.Sprintf("Primitive type %v is not a package", qn))
	}

	return strings.Join(qn.names[:len(qn.names)-1], ".")
}

func (qn *JTypeName) String() string {
	if len(qn.names) == 1 {
		return qn.names[0]
	}

	return strings.Join(qn.names, ".")
}

type JTypeParameter struct {
	name string
	bounds []JObject
}

func NewJTypeParameter(name string, bounds []JObject) *JTypeParameter {
	return &JTypeParameter{name: name, bounds: bounds}
}

type JUnaryExpr struct {
	Op string
	Obj JObject
	is_prefix bool
}

func NewJUnaryExpr(op string, obj JObject, is_prefix bool) *JUnaryExpr {
	if op == "" {
		ReportError("UnaryExpr op cannot be empty")
	} else if obj == nil {
		ReportError("UnaryExpr object cannot be nil")
	}

	return &JUnaryExpr{Op: op, Obj: obj, is_prefix: is_prefix}
}

type JVariableDecl struct {
	Modifiers *JModifiers
	TypeSpec *JReferenceType
	Name string
	Dims int
	Init *JVariableInit
}

func NewJVariableDecl(name string, dims int,
	init *JVariableInit) *JVariableDecl {
	return &JVariableDecl{Name: name, Dims: dims, Init: init}
}

func (j *JVariableDecl) SetInit(init *JVariableInit) {
       j.Init = init
}

func (j *JVariableDecl) SetModifiers(modifiers *JModifiers) {
       j.Modifiers = modifiers
}

func (j *JVariableDecl) SetType(typespec *JReferenceType) {
	j.TypeSpec = typespec
}

type JVariableInit struct {
	Expr JObject
	ArrayList []*JVariableInit
}

func NewJVariableInit(expr JObject, arraylist []*JVariableInit) *JVariableInit {
	if expr == nil && arraylist == nil {
		panic("Either arraylist or expr must be non-nil")
	} else if expr != nil && arraylist != nil && len(arraylist) > 0 {
		panic("Cannot set both arraylist and expr")
	}

	return &JVariableInit{Expr: expr, ArrayList: arraylist}
}

type JWhile struct {
	Expr JObject
	Stmt JObject
	IsDoWhile bool
}

func NewJWhile(expr JObject, stmt JObject, is_do_while bool) *JWhile {
	return &JWhile{Expr: expr, Stmt: stmt, IsDoWhile: is_do_while}
}
