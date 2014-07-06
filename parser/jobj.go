package parser

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
	typestr string
}

func NewJUnimplemented(typestr string) *JUnimplemented {
	return &JUnimplemented{typestr: typestr}
}

type JAnnotation struct {
	name *JTypeName
	elem []JObject
	has_parens bool
}

func NewJAnnotation(name *JTypeName, elem []JObject,
	has_parens bool) *JAnnotation {
	if name == nil {
		reportError("Annotation name cannot be nil")
	}

	return &JAnnotation{name: name, elem: elem, has_parens: has_parens}
}

type JArrayAlloc struct {
	typename *JTypeName
	dimexprs []JObject
	dims int
	init []*JVariableInit
}

func NewJArrayAlloc(typename *JTypeName, dimexprs []JObject,
	dims int) *JArrayAlloc {
	if typename == nil {
		reportError("ArrayExpr name cannot be nil")
	}

	return &JArrayAlloc{typename: typename, dimexprs: dimexprs, dims: dims}
}

func (j *JArrayAlloc) SetInitializers(init []*JVariableInit) {
	j.init = init
}

type JArrayReference struct {
	name *JTypeName
	obj JObject
	expr JObject
}

func NewJArrayReference(name *JTypeName, obj JObject, expr JObject) *JArrayReference {
	if name == nil && obj == nil {
		reportError("NewArrayExpr name or obj must be specified")
	} else if name != nil && obj != nil {
		reportError("NewArrayExpr name and obj cannot both be specified")
	} else if expr == nil {
		reportError("NewArrayExpr expr cannot be nil")
	}

	return &JArrayReference{name: name, obj: obj, expr: expr}
}

type JAssignmentExpr struct {
	left JObject
	op string
	right JObject
}

func NewJAssignmentExpr(left JObject, op string, right JObject) *JAssignmentExpr {
	if left == nil {
		reportError("JAssignmentExpr left expression cannot be nil")
	} else if right == nil {
		reportError("JAssignmentExpr right expression cannot be nil")
	}

	return &JAssignmentExpr{left: left, op: op, right: right}
}

type JBinaryExpr struct {
	obj1 JObject
	op string
	obj2 JObject
}

func NewJBinaryExpr(obj1 JObject, op string, obj2 JObject) *JBinaryExpr {
	if obj1 == nil {
		reportError("BinaryExpr object1 cannot be nil")
	} else if op == "" {
		reportError("BinaryExpr op cannot be empty")
	} else if obj2 == nil {
		reportError("BinaryExpr object2 cannot be nil")
	}

	return &JBinaryExpr{obj1: obj1, op: op, obj2: obj2}
}

type JBlock struct {
	static bool
	list []JObject
}

func NewJBlock(list []JObject) *JBlock {
	return &JBlock{list: list}
}

func (j *JBlock) SetStatic() { j.static = true }

type JCastExpr struct {
	reftype *JReferenceType
	target JObject
}

func NewJCastExpr(reftype *JReferenceType, target JObject) *JCastExpr {
	if reftype == nil {
		reportError("CastExpr reference type cannot be nil")
	} else if target == nil {
		reportError("CastExpr target cannot be nil")
	}

	return &JCastExpr{reftype: reftype, target: target}
}

type JCatch struct {
	modifiers *JModifiers
	typelist []*JTypeName
	name string
	block *JBlock
}

func NewJCatch(modifiers *JModifiers, typelist []*JTypeName, name string,
	block *JBlock) *JCatch {
	if typelist == nil || len(typelist) == 0 {
		panic("List of exception types cannot be nil/empty")
	}

	return &JCatch{modifiers: modifiers, typelist: typelist, name: name,
		block: block}
}

type JClassAllocationExpr struct {
	name *JTypeName
	type_args []*JTypeArgument
	arglist []JObject
	body []JObject
}

func NewJClassAllocationExpr(name *JTypeName, obj_args []JObject,
	arglist []JObject) *JClassAllocationExpr {
	if name == nil {
		reportError("Class name cannot be nil")
	}

	var type_args []*JTypeArgument
	if obj_args != nil && len(obj_args) > 0 {
		type_args = make([]*JTypeArgument, len(obj_args))
		for i, o := range obj_args {
			if arg, ok := o.(*JTypeArgument); ok {
				type_args[i] = arg
			} else {
				reportCastError("JTypeArgument", o)
			}
		}
	}

	return &JClassAllocationExpr{name: name, type_args: type_args,
		arglist: arglist}
}

func (j *JClassAllocationExpr) SetBody(body []JObject) {
	j.body = body
}

type JClassBody struct {
	list []JObject
}

func NewJClassBody(list []JObject) *JClassBody {
	if list == nil || len(list) == 0 {
		panic("JClassBody list cannot be empty/nil")
	}

	return &JClassBody{list: list}
}

type JClassDecl struct {
	modifiers *JModifiers
	name string
	type_params []JObject
	extends *JReferenceType
	interfaces []*JTypeName
	body []JObject
}

func NewJClassDecl(modifiers *JModifiers, name string, type_params []JObject,
	extends *JReferenceType, interfaces []*JTypeName, body []JObject) *JClassDecl {
	if name == "" {
		reportError("JClassDecl name cannot be empty")
	}

	return &JClassDecl{modifiers: modifiers, name: name,
		type_params: type_params, extends: extends, interfaces: interfaces,
		body: body}
}

type JConditionalExpr struct {
	condexpr JObject
	ifexpr JObject
	elseexpr JObject
}

func NewJConditionalExpr(condexpr JObject, ifexpr JObject, elseexpr JObject) *JConditionalExpr {
	return &JConditionalExpr{condexpr: condexpr, ifexpr: ifexpr,
		elseexpr: elseexpr}
}

type JConstantDecl struct {
	modifiers *JModifiers
	typespec *JReferenceType
	name string
	dims int
	init *JVariableInit
}

func NewJConstantDecl(name string, dims int,
	init *JVariableInit) *JConstantDecl {
	return &JConstantDecl{name: name, dims: dims, init: init}
}

func (j *JConstantDecl) SetModifiers(modifiers *JModifiers) {
       j.modifiers = modifiers
}

func (j *JConstantDecl) SetName(name string) {
	if name == "" {
		panic("Constant name cannot be empty")
	} else if j.name != "" {
		panic(fmt.Sprintf("Cannot overwrite constant name \"%s\" with \"%s\"",
			j.name, name))
	}

	j.name = name
}

func (j *JConstantDecl) SetType(typespec *JReferenceType) {
	j.typespec = typespec
}

type JElementValuePair struct {
	name string
	value JObject
}

func NewJElementValuePair(name string, value JObject) *JElementValuePair {
	if name == "" {
		reportError("ReferenceType name cannot be empty")
	}

	return &JElementValuePair{name: name, value: value}
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
				reportCastError("JEnumConstant", v)
			} else {
				constants[i] = jcon
			}
		}
	}

	return &JEnumBody{constants: constants, bodydecl: bodydecl}
}

type JEnumConstant struct {
	annotations []*JAnnotation
	name string
	arglist []JObject
	body []JObject
}

func NewJEnumConstant(alist []JObject, name string, arglist []JObject,
	body []JObject) *JEnumConstant {
	var annotations []*JAnnotation
	if alist != nil && len(alist) > 0 {
		annotations := make([]*JAnnotation, len(alist))
		for i, v := range alist {
			if jann, ok := v.(*JAnnotation); !ok {
				reportCastError("JAnnotation", v)
			} else {
				annotations[i] = jann
			}
		}
	}

	return &JEnumConstant{annotations: annotations, name: name,
		arglist: arglist, body: body}
}

type JEnumDecl struct {
	modifiers *JModifiers
	name string
	interfaces []*JTypeName
	constants []*JEnumConstant
	bodydecl []JObject
}

func NewJEnumDecl(modifiers *JModifiers, name string, interfaces []*JTypeName,
	body *JEnumBody) *JEnumDecl {
	if name == "" {
		reportError("JEnumDecl name cannot be empty")
	}

	var constants []*JEnumConstant
	var bodydecl []JObject
	if body != nil {
		constants = body.constants
		bodydecl = body.bodydecl
	}

	return &JEnumDecl{modifiers: modifiers, name: name, interfaces: interfaces,
		constants: constants, bodydecl: bodydecl}
}

type JForColon struct {
	vardecl *JVariableDecl
	expr JObject
	body JObject
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

	return &JForColon{vardecl: vardecl, expr: expr}
}

func (j *JForColon) SetBody(body JObject) {
	j.body = body
}

type JForExpr struct {
	init []JObject
	expr JObject
	incr []JObject
	body JObject
}

func NewJForExpr(init []JObject, expr JObject, incr []JObject) *JForExpr {
	return &JForExpr{init: init, expr: expr, incr: incr}
}

func (j *JForExpr) SetBody(body JObject) {
	j.body = body
}

type JForVar struct {
	vardecl *JVariableDecl
	decl JObject
	expr JObject
	incr []JObject
	body JObject
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

	return &JForVar{vardecl: vardecl}
}

func (j *JForVar) SetBody(body JObject) {
	j.body = body
}

func (j *JForVar) SetDecl(decl JObject) {
	j.decl = decl
}

func (j *JForVar) SetExpr(expr JObject) {
	j.expr = expr
}

func (j *JForVar) SetIncr(incr []JObject) {
	j.incr = incr
}

func (j *JForVar) SetInit(init *JVariableInit) {
	j.vardecl.init = init
}

type JFormalParameter struct {
	modifiers *JModifiers
	typespec *JReferenceType
	dotdotdot bool
	name string
	dims int
}

func NewJFormalParameter(typespec *JReferenceType, dotdotdot bool,
	name string, dims int) *JFormalParameter {
	if name == "" {
		panic("FormalParameter name cannot be empty")
	}

	return &JFormalParameter{typespec: typespec, dotdotdot: dotdotdot,
		name: name, dims: dims}
}

func (j *JFormalParameter) SetModifiers(modifiers *JModifiers) {
       j.modifiers = modifiers
}

type JIfElseStmt struct {
	cond JObject
	ifblock JObject
	elseblock JObject
}

func NewJIfElseStmt(cond JObject, ifblock JObject, elseblock JObject) *JIfElseStmt {
	if cond == nil {
		panic("'if' condition cannot be nil")
	} else if ifblock == nil {
		panic("'if' block cannot be nil")
	}

	return &JIfElseStmt{cond: cond, ifblock: ifblock, elseblock: elseblock}
}

type JImportStmt struct {
	name *JTypeName
	is_wild bool
	is_static bool
}

func NewJImportStmt(name *JTypeName, is_wild bool, is_static bool) *JImportStmt {
	return &JImportStmt{name: name, is_wild: is_wild, is_static: is_static}
}

type JInstanceOf struct {
	obj JObject
	typespec *JReferenceType
}

func NewJInstanceOf(obj JObject, typespec *JReferenceType) *JInstanceOf {
	if obj == nil {
		reportError("Instanceof object cannot be nil")
	} else if typespec == nil {
		reportError("TwoObjects type specifier cannot be nil")
	}

	return &JInstanceOf{obj: obj, typespec: typespec}
}

type JInterfaceDecl struct {
	modifiers *JModifiers
	name *JTypeName
	type_params []JObject
	extends []*JTypeName
	body []JObject
}

func NewJInterfaceDecl(modifiers *JModifiers, name *JTypeName,
	type_params []JObject, extends []*JTypeName,
	body []JObject) *JInterfaceDecl {
	if name == nil {
		reportError("JInterfaceDecl name cannot be nil")
	}

	return &JInterfaceDecl{modifiers: modifiers, name: name,
		type_params: type_params, extends: extends, body: body}
}

type JInterfaceMethodDecl struct {
	modifiers *JModifiers
	type_params []JObject
	typespec *JReferenceType
	name string
	formal_params []*JFormalParameter
	dims int
	throws []*JTypeName
}

func NewJInterfaceMethodDecl(formal_params []*JFormalParameter, dims int,
	throws []*JTypeName) *JInterfaceMethodDecl {
	return &JInterfaceMethodDecl{formal_params: formal_params, dims: dims,
		throws: throws}
}

func (j *JInterfaceMethodDecl) SetModifiers(modifiers *JModifiers) {
       j.modifiers = modifiers
}

func (j *JInterfaceMethodDecl) SetName(name string) {
	if name == "" {
		panic("Interface name cannot be empty")
	} else if j.name != "" {
		panic(fmt.Sprintf("Cannot overwrite interface name \"%s\" with \"%s\"",
			j.name, name))
	}

	j.name = name
}

func (j *JInterfaceMethodDecl) SetType(typespec *JReferenceType) {
	j.typespec = typespec
}

func (j *JInterfaceMethodDecl) SetTypeParameters(type_params []JObject) {
	j.type_params = type_params
}

type JLabeledStatement struct {
	label string
	stmt JObject
}

func NewJLabeledStatement(label string, stmt JObject) *JLabeledStatement {
	if label == "" {
		reportError("Label cannot be empty")
	} else if stmt == nil {
		reportError("Label object cannot be nil")
	}

	return &JLabeledStatement{label: label, stmt: stmt}
}

type JLocalVariableDecl struct {
	modifiers *JModifiers
	typespec *JReferenceType
	vars []*JVariableDecl
}

func NewJLocalVariableDecl(modifiers *JModifiers, typespec *JReferenceType,
	vars []*JVariableDecl) *JLocalVariableDecl {
	return &JLocalVariableDecl{modifiers: modifiers, typespec: typespec,
		vars: vars}
}

type JMethodAccess struct {
	nameobj JObject
	namekey *GoKeyword
	nametyp *JTypeName
	method string
	arglist []JObject
}

func NewJMethodAccessComplex(obj JObject, name string,
	arglist []JObject) *JMethodAccess {
	if obj == nil {
		panic("Method object cannot be nil")
	} else if name == "" {
		panic("Method name cannot be empty")
	}

	return &JMethodAccess{nameobj: obj, method: name, arglist: arglist}
}

func NewJMethodAccessKeyword(token int, name string, arglist []JObject) *JMethodAccess {
	return &JMethodAccess{namekey: NewGoKeyword(token, name), arglist: arglist}
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

	return &JMethodAccess{nametyp: class, method: name, arglist: arglist}
}

type JMethodDecl struct {
	modifiers *JModifiers
	type_params []JObject
	typespec *JReferenceType
	name string
	formal_params []*JFormalParameter
	dims int
	throws []*JTypeName
	block *JBlock
}

func NewJMethodDecl(formal_params []*JFormalParameter, dims int,
	throws []*JTypeName, block *JBlock) *JMethodDecl {
	return &JMethodDecl{formal_params: formal_params, dims: dims,
		throws: throws, block: block}
}

func (j *JMethodDecl) SetModifiers(modifiers *JModifiers) {
       j.modifiers = modifiers
}

func (j *JMethodDecl) SetName(name string) {
	if name == "" {
		panic("Method name cannot be empty")
	}

	j.name = name
}

func (j *JMethodDecl) SetType(typespec *JReferenceType) {
	j.typespec = typespec
}

func (j *JMethodDecl) SetTypeParameters(type_params []JObject) {
	j.type_params = type_params
}

const (
	modPublic = 0x1
	modProtected = 0x2
	modPrivate = 0x4
	modAbstract = 0x8
	modFinal = 0x10
	modStatic = 0x20
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
	case "public": j.mod_bits |= modPublic
	case "protected": j.mod_bits |= modProtected
	case "private": j.mod_bits |= modPrivate
	case "abstract": j.mod_bits |= modAbstract
	case "final": j.mod_bits |= modFinal
	case "static": j.mod_bits |= modStatic
	case "transient": j.mod_bits |= modTransient
	case "volatile": j.mod_bits |= modVolatile
	case "native": j.mod_bits |= modNative
	case "synchronized": j.mod_bits |= modSynchronized
	default: reportError(fmt.Sprintf("Unknown modifier \"%s\"", name))
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
			case modPublic: io.WriteString(out, "public ")
			case modProtected: io.WriteString(out, "protected ")
			case modPrivate: io.WriteString(out, "private ")
			case modAbstract: io.WriteString(out, "abstract ")
			case modFinal: io.WriteString(out, "final ")
			case modStatic: io.WriteString(out, "static ")
			case modTransient: io.WriteString(out, "transient ")
			case modVolatile: io.WriteString(out, "volatile ")
			case modNative: io.WriteString(out, "native ")
			case modSynchronized: io.WriteString(out, "synchronized ")
			}
		}
	}
}

type JNameDotObject struct {
	name *JTypeName
	obj JObject
}

func NewJNameDotObject(name *JTypeName, obj JObject) *JNameDotObject {
	if name == nil {
		reportError("NameDotObject name cannot be nil")
	} else if obj == nil {
		reportError("NameDotObject object cannot be nil")
	}

	return &JNameDotObject{name: name, obj: obj}
}

type JObjectDotName struct {
	obj JObject
	name *JTypeName
}

func NewJObjectDotName(obj JObject, name *JTypeName) *JObjectDotName {
	if obj == nil {
		reportError("ObjectDotName object cannot be nil")
	} else if name == nil {
		reportError("ObjectDotName name cannot be nil")
	}

	return &JObjectDotName{obj: obj, name: name}
}

type JPackageStmt struct {
	name *JTypeName
}

func NewJPackageStmt(name *JTypeName) *JPackageStmt {
	return &JPackageStmt{name: name}
}

type JParens struct {
	expr JObject
}

func NewJParens(expr JObject) *JParens {
	if expr == nil {
		reportError("JParens expression cannot be nil")
	}

	return &JParens{expr: expr}
}

type JProgramFile struct {
	pkg *JPackageStmt
	imports []JObject
	type_decls []JObject
}

func NewJProgramFile(pobj JObject, imports []JObject,
	type_decls []JObject) *JProgramFile {
	j := &JProgramFile{}

	if pobj != nil {
		if pkg, ok := pobj.(*JPackageStmt); !ok {
			reportCastError("JPackageStmt", pobj)
		} else {
			j.pkg = pkg
		}
	}

	j.imports = imports
	j.type_decls = type_decls

	return j
}

type JReferenceType struct {
	name *JTypeName
	type_args []*JTypeArgument
	dims int
}

func NewJReferenceType(name *JTypeName, obj_args []JObject,
	dims int) *JReferenceType {
	if name == nil {
		reportError("ReferenceType name cannot be nil")
	}

	var type_args []*JTypeArgument
	if obj_args != nil && len(obj_args) > 0 {
		type_args = make([]*JTypeArgument, len(obj_args))
		for i, o := range obj_args {
			if arg, ok := o.(*JTypeArgument); ok {
				type_args[i] = arg
			} else {
				reportCastError("JTypeArgument", o)
			}
		}
	}

	return &JReferenceType{name: name, type_args: type_args, dims: dims}
}

type JSimpleStatement struct {
	keyword *GoKeyword
	object JObject
}

func NewJSimpleStatement(keyword *GoKeyword, object JObject) *JSimpleStatement {
	return &JSimpleStatement{keyword: keyword, object: object}
}

type JSwitch struct {
	expr JObject
	groups []*JSwitchGroup
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
				reportCastError("JSwitch", o)
			} else {
				groups[i] = grp
			}
		}
	}

	return &JSwitch{expr: expr, groups: groups}
}

type JSwitchGroup struct {
	labels []*JSwitchLabel
	stmts []JObject
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
				reportCastError("JSwitchLabel", o)
			} else {
				labels[i] = lbl
			}
		}
	}

	return &JSwitchGroup{labels: labels, stmts: stmtlist}
}

type JSwitchLabel struct {
	name string
	expr JObject
	is_default bool
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

	return &JSwitchLabel{name: name, expr: expr, is_default: is_default}
}

type JSynchronized struct {
	expr JObject
	block *JBlock
}

func NewJSynchronized(expr JObject, block *JBlock) *JSynchronized {
	if block == nil {
		reportError("JSynchronized block cannot be nil")
	}

	return &JSynchronized{expr: expr, block: block}
}

type JTry struct {
	block *JBlock
	catches []*JCatch
	finally *JBlock
}

func NewJTry(block *JBlock, clist []JObject, finally *JBlock) *JTry {
	if block == nil {
		reportError("JTry block cannot be nil")
	}

	var catches []*JCatch
	if clist != nil && len(clist) > 0 {
		catches = make([]*JCatch, len(clist))
		for i, c := range clist {
			if catch, ok := c.(*JCatch); !ok {
				reportCastError(fmt.Sprintf("JCatch#%d", i), c)
			} else {
				catches[i] = catch
			}
		}
	}

	return &JTry{block: block, catches: catches, finally: finally}
}

const (
	TS_NONE = iota
	TS_EXTENDS
	TS_SUPER
	TS_PLAIN
)

type JTypeArgument struct {
	typespec *JReferenceType
	ts_type int
}

func NewJTypeArgument(typespec *JReferenceType, ts_type int) *JTypeArgument {
	return &JTypeArgument{typespec: typespec, ts_type: ts_type}
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
	op string
	obj JObject
	is_prefix bool
}

func NewJUnaryExpr(op string, obj JObject, is_prefix bool) *JUnaryExpr {
	if op == "" {
		reportError("UnaryExpr op cannot be empty")
	} else if obj == nil {
		reportError("UnaryExpr object cannot be nil")
	}

	return &JUnaryExpr{op: op, obj: obj, is_prefix: is_prefix}
}

type JVariableDecl struct {
	modifiers *JModifiers
	typespec *JReferenceType
	name string
	dims int
	init *JVariableInit
}

func NewJVariableDecl(name string, dims int,
	init *JVariableInit) *JVariableDecl {
	return &JVariableDecl{name: name, dims: dims, init: init}
}

func (j *JVariableDecl) SetInit(init *JVariableInit) {
       j.init = init
}

func (j *JVariableDecl) SetModifiers(modifiers *JModifiers) {
       j.modifiers = modifiers
}

func (j *JVariableDecl) SetType(typespec *JReferenceType) {
	j.typespec = typespec
}

type JVariableInit struct {
	expr JObject
	arraylist []*JVariableInit
}

func NewJVariableInit(expr JObject, arraylist []*JVariableInit) *JVariableInit {
	if expr == nil && arraylist == nil {
		panic("Either arraylist or expr must be non-nil")
	} else if expr != nil && arraylist != nil && len(arraylist) > 0 {
		panic("Cannot set both arraylist and expr")
	}

	return &JVariableInit{expr: expr, arraylist: arraylist}
}

type JWhile struct {
	expr JObject
	stmt JObject
	is_do_while bool
}

func NewJWhile(expr JObject, stmt JObject, is_do_while bool) *JWhile {
	return &JWhile{expr: expr, stmt: stmt, is_do_while: is_do_while}
}
