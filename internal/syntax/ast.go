package syntax

import "fmt"

// Kind identifies an AST node.
type Kind int

const (
	KindObjectFile Kind = iota
	KindObject
	KindType
	KindID
	KindName
	KindPropertyList
	KindProperty
	KindLiteral
	KindCodeValue
	KindBlock
	KindCodeSection
	KindCall
	KindIdentifier
	KindBinary
	KindAssign
	KindParen
	KindExprStmt
	KindError
	KindVariable
	KindProcedure
)

func (k Kind) String() string {
	switch k {
	case KindObjectFile:
		return "ObjectFile"
	case KindObject:
		return "Object"
	case KindType:
		return "Type"
	case KindID:
		return "ID"
	case KindName:
		return "Name"
	case KindPropertyList:
		return "PropertyList"
	case KindProperty:
		return "Property"
	case KindLiteral:
		return "Literal"
	case KindCodeValue:
		return "CodeValue"
	case KindBlock:
		return "Block"
	case KindCodeSection:
		return "CodeSection"
	case KindCall:
		return "Call"
	case KindIdentifier:
		return "Identifier"
	case KindBinary:
		return "Binary"
	case KindAssign:
		return "Assign"
	case KindParen:
		return "Paren"
	case KindExprStmt:
		return "ExprStmt"
	case KindError:
		return "Error"
	case KindVariable:
		return "Variable"
	case KindProcedure:
		return "Procedure"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

// Severity is the severity of a parse diagnostic.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

// Diagnostic is a parse error. Rule findings are a separate list.
type Diagnostic struct {
	ID       string
	Message  string
	Severity Severity
	Span     SourceSpan
}

// Node is one AST node. Children are ordered as in the file.
type Node interface {
	Kind() Kind
	Span() SourceSpan
	FullSpan() SourceSpan
	FirstToken() int
	LastToken() int
	Parent() Node
	SetParent(Node)
	Children() []Node

	tokenRange() (int, int)
	explicitSpan() bool
	setSpans(span, full SourceSpan)
}

type nodeBase struct {
	kind     Kind
	span     SourceSpan
	full     SourceSpan
	first    int
	last     int
	parent   Node
	explicit bool
}

func (n *nodeBase) Kind() Kind             { return n.kind }
func (n *nodeBase) Span() SourceSpan       { return n.span }
func (n *nodeBase) FullSpan() SourceSpan   { return n.full }
func (n *nodeBase) FirstToken() int        { return n.first }
func (n *nodeBase) LastToken() int         { return n.last }
func (n *nodeBase) Parent() Node           { return n.parent }
func (n *nodeBase) SetParent(p Node)       { n.parent = p }
func (n *nodeBase) tokenRange() (int, int) { return n.first, n.last }
func (n *nodeBase) explicitSpan() bool     { return n.explicit }
func (n *nodeBase) setSpans(span, full SourceSpan) {
	n.span = span
	n.full = full
}

// SyntaxTree is the parse result: one object file.
type SyntaxTree = ObjectFile

// ObjectFile is the root. Path is the path handed to Parse.
type ObjectFile struct {
	nodeBase
	Path        string
	Tokens      []Token
	Diagnostics []Diagnostic
	Objects     []*Object
}

func (f *ObjectFile) Children() []Node {
	out := make([]Node, len(f.Objects))
	for i, obj := range f.Objects {
		out[i] = obj
	}
	return out
}

// OpaqueSection is a section the first slice does not deeply parse.
type OpaqueSection struct {
	Name string
	Span SourceSpan
}

// Object is one OBJECT. Type, ID, and Name are separate children.
type Object struct {
	nodeBase
	Type             *HeaderPart
	ID               *HeaderPart
	Name             *HeaderPart
	ObjectProperties *PropertyList
	Properties       *PropertyList
	Members          []Node
	OpaqueSections   []OpaqueSection
}

func (o *Object) Children() []Node {
	var out []Node
	if o.Type != nil {
		out = append(out, o.Type)
	}
	if o.ID != nil {
		out = append(out, o.ID)
	}
	if o.Name != nil {
		out = append(out, o.Name)
	}
	if o.ObjectProperties != nil {
		out = append(out, o.ObjectProperties)
	}
	if o.Properties != nil {
		out = append(out, o.Properties)
	}
	out = append(out, o.Members...)
	return out
}

// HeaderPart is the Type, ID, or Name child of an object header.
type HeaderPart struct {
	nodeBase
	Text string
}

func (h *HeaderPart) Children() []Node { return nil }

// PropertyList is OBJECT-PROPERTIES or PROPERTIES.
type PropertyList struct {
	nodeBase
	Properties []*Property
}

func (l *PropertyList) Children() []Node {
	out := make([]Node, len(l.Properties))
	for i, p := range l.Properties {
		out[i] = p
	}
	return out
}

// Property is one Name=Value pair. Name is the raw text before '='.
type Property struct {
	nodeBase
	Name  string
	Value Node
}

func (p *Property) Children() []Node {
	if p.Value == nil {
		return nil
	}
	return []Node{p.Value}
}

// Literal is a raw property value or a C/AL literal.
// For a string, Text is the decoded contents and Span includes the quotes.
type Literal struct {
	nodeBase
	Text string
}

func (l *Literal) Children() []Node { return nil }

// Identifier is a name written in C/AL.
type Identifier struct {
	nodeBase
	Text string
}

func (i *Identifier) Children() []Node { return nil }

// CodeValue is a trigger or property value that holds C/AL.
type CodeValue struct {
	nodeBase
	Locals []*Variable
	Body   *Block
}

func (c *CodeValue) Children() []Node {
	var out []Node
	for _, loc := range c.Locals {
		out = append(out, loc)
	}
	if c.Body != nil {
		out = append(out, c.Body)
	}
	return out
}

// Block is BEGIN … END, or BEGIN … END. for the object trigger.
type Block struct {
	nodeBase
	Statements []Node
}

func (b *Block) Children() []Node { return b.Statements }

// CodeSection is the CODE section of a table or codeunit.
type CodeSection struct {
	nodeBase
	Globals       []*Variable
	Procedures    []*Procedure
	ObjectTrigger *CodeValue
}

func (s *CodeSection) Children() []Node {
	var out []Node
	for _, g := range s.Globals {
		out = append(out, g)
	}
	for _, p := range s.Procedures {
		out = append(out, p)
	}
	if s.ObjectTrigger != nil {
		out = append(out, s.ObjectTrigger)
	}
	return out
}

// CallExpr is a call. The '(' and ')' tokens belong to the call.
// A statement call includes the terminating semicolon in its token range.
type CallExpr struct {
	nodeBase
	Callee Node
	Args   []Node
	commas []int
	file   *ObjectFile
}

func (c *CallExpr) Children() []Node {
	out := make([]Node, 0, 1+len(c.Args))
	if c.Callee != nil {
		out = append(out, c.Callee)
	}
	out = append(out, c.Args...)
	return out
}

// CommaTok returns the comma token between Args[afterArg] and Args[afterArg+1].
// ok is false when that comma is missing.
func (c *CallExpr) CommaTok(afterArg int) (Token, bool) {
	if c == nil || c.file == nil || afterArg < 0 || afterArg >= len(c.commas) {
		return Token{}, false
	}
	idx := c.commas[afterArg]
	if idx < 0 || idx >= len(c.file.Tokens) {
		return Token{}, false
	}
	return c.file.Tokens[idx], true
}

// BinaryExpr is a binary operator with left and right operands.
// Op is the index of the operator token.
type BinaryExpr struct {
	nodeBase
	Left  Node
	Op    int
	Right Node
}

func (b *BinaryExpr) Children() []Node { return []Node{b.Left, b.Right} }

// AssignStmt is an assignment. Assign is the index of the ':=' token.
type AssignStmt struct {
	nodeBase
	Left   Node
	Assign int
	Right  Node
}

func (a *AssignStmt) Children() []Node { return []Node{a.Left, a.Right} }

// ParenExpr is a parenthesized expression. The '(' ')' pair is part of this node.
type ParenExpr struct {
	nodeBase
	Expr Node
}

func (p *ParenExpr) Children() []Node {
	if p.Expr == nil {
		return nil
	}
	return []Node{p.Expr}
}

// ExprStmt is an expression used as a statement, including a trailing semicolon
// when the expression itself is not a call.
type ExprStmt struct {
	nodeBase
	Expr Node
}

func (e *ExprStmt) Children() []Node {
	if e.Expr == nil {
		return nil
	}
	return []Node{e.Expr}
}

// ErrorNode covers text the parser skipped.
type ErrorNode struct {
	nodeBase
	Message string
}

func (e *ErrorNode) Children() []Node { return nil }

// Variable is a global or local. SymbolID is the @ id; Name does not include it.
type Variable struct {
	nodeBase
	Name     string
	SymbolID int
}

func (v *Variable) Children() []Node { return nil }

// Procedure is a PROCEDURE. The first tests do not produce one.
type Procedure struct {
	nodeBase
	Name string
}

func (p *Procedure) Children() []Node { return nil }

var (
	_ Node = (*ObjectFile)(nil)
	_ Node = (*Object)(nil)
	_ Node = (*HeaderPart)(nil)
	_ Node = (*PropertyList)(nil)
	_ Node = (*Property)(nil)
	_ Node = (*Literal)(nil)
	_ Node = (*Identifier)(nil)
	_ Node = (*CodeValue)(nil)
	_ Node = (*Block)(nil)
	_ Node = (*CodeSection)(nil)
	_ Node = (*CallExpr)(nil)
	_ Node = (*BinaryExpr)(nil)
	_ Node = (*AssignStmt)(nil)
	_ Node = (*ParenExpr)(nil)
	_ Node = (*ExprStmt)(nil)
	_ Node = (*ErrorNode)(nil)
	_ Node = (*Variable)(nil)
	_ Node = (*Procedure)(nil)
)
