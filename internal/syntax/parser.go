package syntax

import "strings"

type parser struct {
	lex     *lexer
	file    *ObjectFile
	peeked  Token
	hasPeek bool
}

func (p *parser) peek() Token {
	if !p.hasPeek {
		p.peeked = p.lex.read()
		p.hasPeek = true
	}
	return p.peeked
}

func (p *parser) advance() Token {
	tok := p.peek()
	p.hasPeek = false
	return tok
}

func (p *parser) report(msg string) {
	tok := p.peek()
	span := tok.Span
	if tok.Kind == TokEOF && len(p.lex.tokens) > 0 {
		span = p.lex.tokens[len(p.lex.tokens)-1].Span
	}
	p.file.Diagnostics = append(p.file.Diagnostics, Diagnostic{
		ID:       "PARSE0001",
		Message:  msg,
		Severity: SeverityError,
		Span:     span,
	})
}

func (p *parser) expectKeyword(word string) Token {
	tok := p.peek()
	if strings.EqualFold(tok.Text, word) && (tok.Kind == TokKeyword || tok.Kind == TokIdentifier) {
		return p.advance()
	}
	p.report("expected " + word)
	if tok.Kind != TokEOF {
		return p.advance()
	}
	return tok
}

func (p *parser) expectSymbol(text string) Token {
	tok := p.peek()
	if tok.Text == text {
		return p.advance()
	}
	p.report("expected " + text)
	if tok.Kind != TokEOF {
		return p.advance()
	}
	return tok
}

func (p *parser) parseFile() {
	for p.peek().Kind != TokEOF {
		if strings.EqualFold(p.peek().Text, "OBJECT") {
			p.file.Objects = append(p.file.Objects, p.parseObject())
			continue
		}
		p.report("expected OBJECT")
		p.advance()
	}
}

func (p *parser) parseObject() *Object {
	objTok := p.expectKeyword("OBJECT")
	obj := &Object{}
	obj.kind = KindObject
	obj.first = objTok.Index
	obj.OpaqueSections = []OpaqueSection{}

	typeTok := p.advance()
	obj.Type = p.header(KindType, typeTok, typeTok.Text)

	idTok := p.peek()
	if idTok.Kind == TokNumber {
		idTok = p.advance()
	} else {
		p.report("expected object id")
		idTok = p.advance()
	}
	obj.ID = p.header(KindID, idTok, idTok.Text)

	obj.Name = p.parseName(objTok.Span.Start.Line)
	p.expectSymbol("{")
	p.parseObjectBody(obj)
	return obj
}

func (p *parser) header(kind Kind, tok Token, text string) *HeaderPart {
	h := &HeaderPart{Text: text}
	h.kind = kind
	if tok.Index >= 0 {
		h.first = tok.Index
		h.last = tok.Index
	} else {
		h.explicit = true
		h.first = -1
		h.last = -1
	}
	return h
}

func (p *parser) parseName(line int) *HeaderPart {
	var toks []Token
	for {
		t := p.peek()
		if t.Kind == TokEOF || t.Text == "{" || t.Span.Start.Line != line {
			break
		}
		toks = append(toks, p.advance())
	}
	h := &HeaderPart{}
	h.kind = KindName
	if len(toks) == 0 {
		h.explicit = true
		h.first = -1
		h.last = -1
		p.report("expected object name")
		return h
	}
	h.first = toks[0].Index
	h.last = toks[len(toks)-1].Index
	raw := p.lex.slice(toks[0].Span.Start.Offset, toks[len(toks)-1].Span.End.Offset)
	h.Text = strings.TrimSpace(raw)
	return h
}

func (p *parser) parseObjectBody(obj *Object) {
	for {
		t := p.peek()
		if t.Kind == TokEOF {
			p.report("unclosed object")
			obj.last = obj.first
			return
		}
		if t.Text == "}" {
			obj.last = p.advance().Index
			return
		}
		switch strings.ToUpper(t.Text) {
		case "OBJECT-PROPERTIES":
			obj.ObjectProperties = p.parsePropertySection()
		case "PROPERTIES":
			obj.Properties = p.parsePropertySection()
		case "CODE":
			obj.Members = append(obj.Members, p.parseCodeSection())
		default:
			p.report("unknown section " + t.Text)
			before := t.Index
			p.advance()
			if p.peek().Index == before {
				return
			}
		}
	}
}

func (p *parser) parsePropertySection() *PropertyList {
	kw := p.advance()
	p.expectSymbol("{")
	list := &PropertyList{}
	list.kind = KindPropertyList
	list.first = kw.Index
	for {
		t := p.peek()
		if t.Kind == TokEOF || t.Text == "}" {
			break
		}
		before := t.Index
		list.Properties = append(list.Properties, p.parseProperty())
		if p.peek().Kind != TokEOF && p.peek().Index == before {
			p.advance()
		}
	}
	list.last = p.expectSymbol("}").Index
	return list
}

func (p *parser) parseProperty() *Property {
	var nameToks []Token
	for {
		t := p.peek()
		if t.Kind == TokEOF || t.Text == "=" || t.Text == "}" || t.Text == ";" {
			break
		}
		nameToks = append(nameToks, p.advance())
	}
	eq := p.expectSymbol("=")
	prop := &Property{}
	prop.kind = KindProperty
	if len(nameToks) > 0 {
		prop.first = nameToks[0].Index
		prop.Name = p.lex.slice(nameToks[0].Span.Start.Offset, nameToks[len(nameToks)-1].Span.End.Offset)
	} else {
		prop.first = eq.Index
	}
	next := p.peek()
	if strings.EqualFold(next.Text, "BEGIN") || strings.EqualFold(next.Text, "VAR") {
		prop.Value = p.parseCodeValue(false)
		prop.last = p.expectSymbol(";").Index
		return prop
	}
	prop.Value = p.parseRawLiteral(eq)
	prop.last = p.expectSymbol(";").Index
	return prop
}

func (p *parser) parseRawLiteral(eq Token) *Literal {
	var toks []Token
	for {
		t := p.peek()
		if t.Kind == TokEOF || t.Text == ";" || t.Text == "}" {
			break
		}
		toks = append(toks, p.advance())
	}
	lit := &Literal{}
	lit.kind = KindLiteral
	if len(toks) == 0 {
		end := eq.Span.End
		if semi := p.peek(); semi.Kind != TokEOF {
			end = semi.Span.Start
		}
		lit.explicit = true
		lit.span = SourceSpan{Start: eq.Span.End, End: end}
		lit.first = -1
		lit.last = -1
		lit.Text = ""
		return lit
	}
	lit.first = toks[0].Index
	lit.last = toks[len(toks)-1].Index
	lit.Text = p.lex.slice(toks[0].Span.Start.Offset, toks[len(toks)-1].Span.End.Offset)
	return lit
}

func (p *parser) parseCodeValue(objectTrigger bool) *CodeValue {
	block := p.parseBlock(objectTrigger)
	cv := &CodeValue{Body: block}
	cv.kind = KindCodeValue
	cv.first = block.first
	cv.last = block.last
	return cv
}

func (p *parser) parseBlock(objectTrigger bool) *Block {
	begin := p.expectKeyword("BEGIN")
	p.lex.setMode(modeCode)
	block := &Block{}
	block.kind = KindBlock
	block.first = begin.Index
	for {
		t := p.peek()
		if t.Kind == TokEOF || strings.EqualFold(t.Text, "END") || t.Text == "}" {
			break
		}
		before := t.Index
		block.Statements = append(block.Statements, p.parseStmt())
		if p.peek().Kind != TokEOF && p.peek().Index == before {
			p.advance()
		}
	}
	end := p.expectKeyword("END")
	block.last = end.Index
	if objectTrigger && p.peek().Text == "." {
		block.last = p.advance().Index
	}
	p.lex.setMode(modeStructure)
	return block
}

func (p *parser) parseCodeSection() *CodeSection {
	kw := p.expectKeyword("CODE")
	p.expectSymbol("{")
	sec := &CodeSection{}
	sec.kind = KindCodeSection
	sec.first = kw.Index
	if strings.EqualFold(p.peek().Text, "BEGIN") {
		block := p.parseBlock(true)
		cv := &CodeValue{Body: block}
		cv.kind = KindCodeValue
		cv.first = block.first
		cv.last = block.last
		sec.ObjectTrigger = cv
	}
	sec.last = p.expectSymbol("}").Index
	return sec
}

func (p *parser) parseStmt() Node {
	expr := p.parseExpr()
	if p.peek().Text == ":=" {
		op := p.advance()
		right := p.parseExpr()
		as := &AssignStmt{Left: expr, Assign: op.Index, Right: right}
		as.kind = KindAssign
		as.first = firstTok(expr)
		as.last = lastTok(right)
		if p.peek().Text == ";" {
			as.last = p.advance().Index
		}
		return as
	}
	if p.peek().Text == ";" {
		semi := p.advance()
		if call, ok := expr.(*CallExpr); ok {
			call.last = semi.Index
			return call
		}
		stmt := &ExprStmt{Expr: expr}
		stmt.kind = KindExprStmt
		stmt.first = firstTok(expr)
		stmt.last = semi.Index
		return stmt
	}
	return expr
}

func (p *parser) parseExpr() Node { return p.parseBinary(1) }

func (p *parser) parseBinary(minPrec int) Node {
	left := p.parseUnary()
	for {
		prec, ok := binaryPrec(p.peek())
		if !ok || prec < minPrec {
			break
		}
		op := p.advance()
		right := p.parseBinary(prec + 1)
		bin := &BinaryExpr{Left: left, Op: op.Index, Right: right}
		bin.kind = KindBinary
		bin.first = firstTok(left)
		bin.last = lastTok(right)
		left = bin
	}
	return left
}

func binaryPrec(tok Token) (int, bool) {
	switch strings.ToUpper(tok.Text) {
	case "OR", "XOR":
		return 1, true
	case "AND":
		return 2, true
	case "=", "<>", "<", ">", "<=", ">=":
		return 3, true
	case "+", "-":
		return 4, true
	case "*", "/", "DIV", "MOD":
		return 5, true
	default:
		return 0, false
	}
}

func (p *parser) parseUnary() Node {
	tok := p.peek()
	var opKind UnaryOp
	switch {
	case tok.Text == "-":
		opKind = UnaryMinus
	case tok.Text == "+":
		opKind = UnaryPlus
	case strings.EqualFold(tok.Text, "NOT") && (tok.Kind == TokKeyword || tok.Kind == TokIdentifier):
		opKind = UnaryNot
	default:
		return p.parsePostfix()
	}
	op := p.advance()
	operand := p.parseUnary()
	u := &UnaryExpr{OpKind: opKind, Op: op.Index, X: operand}
	u.kind = KindUnary
	u.first = op.Index
	u.last = lastTok(operand)
	return u
}

func (p *parser) parsePostfix() Node {
	left := p.parsePrimary()
	for {
		switch p.peek().Text {
		case "(":
			left = p.parseCall(left)
		case "[":
			left = p.parseIndex(left)
		case "::":
			left = p.parseOptionAccess(left)
		default:
			return left
		}
	}
}

func (p *parser) parsePrimary() Node {
	tok := p.peek()
	if tok.Kind == TokEOF {
		p.report("expected expression")
		e := &ErrorNode{Message: "expected expression"}
		e.kind = KindError
		e.explicit = true
		e.first = -1
		e.last = -1
		return e
	}
	if tok.Text == "(" {
		open := p.advance()
		inner := p.parseExpr()
		close := p.expectSymbol(")")
		paren := &ParenExpr{Expr: inner}
		paren.kind = KindParen
		paren.first = open.Index
		paren.last = close.Index
		return paren
	}
	if tok.Kind == TokString || tok.Kind == TokNumber || isBool(tok) {
		t := p.advance()
		lit := &Literal{}
		lit.kind = KindLiteral
		lit.first = t.Index
		lit.last = t.Index
		if t.Kind == TokString {
			lit.Text = unquoteCAL(t.Text)
		} else {
			lit.Text = t.Text
		}
		return lit
	}
	if tok.Kind == TokIdentifier || tok.Kind == TokKeyword {
		t := p.advance()
		id := &Identifier{Text: t.Text}
		id.kind = KindIdentifier
		id.first = t.Index
		id.last = t.Index
		return id
	}
	p.report("expected expression")
	e := &ErrorNode{Message: "expected expression"}
	e.kind = KindError
	e.first = tok.Index
	e.last = tok.Index
	p.advance()
	return e
}

func (p *parser) parseCall(callee Node) *CallExpr {
	p.expectSymbol("(")
	call := &CallExpr{Callee: callee, file: p.file}
	call.kind = KindCall
	call.first = firstTok(callee)
	if p.peek().Text != ")" && p.peek().Kind != TokEOF {
		for {
			call.Args = append(call.Args, p.parseExpr())
			if p.peek().Text != "," {
				break
			}
			call.commas = append(call.commas, p.advance().Index)
		}
	}
	close := p.expectSymbol(")")
	call.last = close.Index
	return call
}

func (p *parser) parseIndex(x Node) *IndexExpr {
	lbrack := p.expectSymbol("[")
	idx := &IndexExpr{X: x, Lbrack: lbrack.Index, file: p.file}
	idx.kind = KindIndex
	idx.first = firstTok(x)
	if p.peek().Text != "]" && p.peek().Kind != TokEOF {
		for {
			idx.Args = append(idx.Args, p.parseExpr())
			if p.peek().Text != "," {
				break
			}
			idx.commas = append(idx.commas, p.advance().Index)
		}
	}
	rbrack := p.expectSymbol("]")
	idx.Rbrack = rbrack.Index
	idx.last = rbrack.Index
	return idx
}

func (p *parser) parseOptionAccess(x Node) *OptionAccessExpr {
	cc := p.expectSymbol("::")
	sel := p.parsePrimary()
	oa := &OptionAccessExpr{X: x, ColonColon: cc.Index, Sel: sel}
	oa.kind = KindOptionAccess
	oa.first = firstTok(x)
	oa.last = lastTok(sel)
	return oa
}

func isBool(tok Token) bool {
	return strings.EqualFold(tok.Text, "TRUE") || strings.EqualFold(tok.Text, "FALSE")
}

func unquoteCAL(raw string) string {
	if len(raw) < 2 || raw[0] != '\'' || raw[len(raw)-1] != '\'' {
		return raw
	}
	return strings.ReplaceAll(raw[1:len(raw)-1], "''", "'")
}

func firstTok(n Node) int {
	if n == nil {
		return -1
	}
	f, _ := n.tokenRange()
	return f
}

func lastTok(n Node) int {
	if n == nil {
		return -1
	}
	_, l := n.tokenRange()
	return l
}
