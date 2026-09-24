package syntax

import (
	"strings"
	"unicode"
)

type lexMode int

const (
	modeStructure lexMode = iota
	modeDeclaration
	modeCode
	modeRaw
)

type lexer struct {
	src     string
	runes   []rune
	byteOff []int
	pos     int
	line    int
	col     int
	mode    lexMode
	tokens  []Token
	done    bool
}

type mark struct {
	pos  int
	line int
	col  int
}

func newLexer(src string) *lexer {
	runes, byteOff := indexSource(src)
	return &lexer{
		src:     src,
		runes:   runes,
		byteOff: byteOff,
		line:    1,
		col:     1,
		mode:    modeStructure,
	}
}

func indexSource(s string) ([]rune, []int) {
	runes := make([]rune, 0, len(s))
	byteOff := make([]int, 0, len(s)+1)
	for i, r := range s {
		byteOff = append(byteOff, i)
		runes = append(runes, r)
	}
	byteOff = append(byteOff, len(s))
	return runes, byteOff
}

func (l *lexer) setMode(m lexMode) { l.mode = m }

func (l *lexer) markAt() mark { return mark{l.pos, l.line, l.col} }

func (l *lexer) restore(m mark) { l.pos, l.line, l.col = m.pos, m.line, m.col }

func (l *lexer) slice(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if end > len(l.byteOff)-1 {
		end = len(l.byteOff) - 1
	}
	return l.src[l.byteOff[start]:l.byteOff[end]]
}

func (l *lexer) spanFrom(start mark) SourceSpan {
	return SourceSpan{
		Start: SourcePosition{Offset: start.pos, Line: start.line, Column: start.col},
		End:   SourcePosition{Offset: l.pos, Line: l.line, Column: l.col},
	}
}

func (l *lexer) bump() rune {
	r := l.runes[l.pos]
	l.pos++
	if r == '\n' {
		l.line++
		l.col = 1
		return r
	}
	l.col++
	return r
}

func (l *lexer) eof() bool { return l.pos >= len(l.runes) }

// finish attaches trailing trivia of the last token. It is safe to call twice.
func (l *lexer) finish() {
	if !l.done {
		l.read()
	}
}

func (l *lexer) read() Token {
	if l.done {
		return Token{Kind: TokEOF, Index: -1}
	}
	trivia := l.scanTrivia()
	if l.eof() {
		if n := len(l.tokens); n > 0 {
			l.tokens[n-1].TrailingTrivia = trivia
		}
		l.done = true
		return Token{Kind: TokEOF, Index: -1}
	}
	lead := trivia
	if n := len(l.tokens); n > 0 {
		trail, rest := splitTrivia(trivia)
		l.tokens[n-1].TrailingTrivia = trail
		lead = rest
	}
	tok := l.scanToken()
	tok.LeadingTrivia = lead
	tok.Index = len(l.tokens)
	l.tokens = append(l.tokens, tok)
	return tok
}

// splitTrivia assigns trivia through the first line break to the previous token.
// Trivia with no line break belongs to the next token.
func splitTrivia(ts []Trivia) (trail, lead []Trivia) {
	for i, t := range ts {
		if t.Kind == TriviaEndOfLine {
			trail = append([]Trivia(nil), ts[:i+1]...)
			lead = append([]Trivia(nil), ts[i+1:]...)
			return trail, lead
		}
	}
	return nil, ts
}

func (l *lexer) scanTrivia() []Trivia {
	var out []Trivia
	for !l.eof() {
		r := l.runes[l.pos]
		switch {
		case r == ' ' || r == '\t':
			out = append(out, l.scanWhitespace())
		case r == '\n' || r == '\r':
			out = append(out, l.scanNewline())
		case r == '/' && l.pos+1 < len(l.runes) && l.runes[l.pos+1] == '/':
			out = append(out, l.scanLineComment())
		case r == '{' && (l.mode == modeCode || l.mode == modeRaw):
			out = append(out, l.scanBlockComment())
		default:
			return out
		}
	}
	return out
}

func (l *lexer) scanWhitespace() Trivia {
	start := l.markAt()
	for !l.eof() && (l.runes[l.pos] == ' ' || l.runes[l.pos] == '\t') {
		l.bump()
	}
	return Trivia{Kind: TriviaWhitespace, Span: l.spanFrom(start), Text: l.slice(start.pos, l.pos)}
}

func (l *lexer) scanNewline() Trivia {
	start := l.markAt()
	if l.runes[l.pos] == '\r' {
		l.pos++
		if !l.eof() && l.runes[l.pos] == '\n' {
			l.pos++
		}
		l.line++
		l.col = 1
	} else {
		l.bump()
	}
	return Trivia{Kind: TriviaEndOfLine, Span: l.spanFrom(start), Text: l.slice(start.pos, l.pos)}
}

func (l *lexer) scanLineComment() Trivia {
	start := l.markAt()
	l.bump()
	l.bump()
	for !l.eof() && l.runes[l.pos] != '\n' && l.runes[l.pos] != '\r' {
		l.bump()
	}
	return Trivia{Kind: TriviaLineComment, Span: l.spanFrom(start), Text: l.slice(start.pos, l.pos)}
}

func (l *lexer) scanBlockComment() Trivia {
	start := l.markAt()
	l.bump() // {
	for !l.eof() && l.runes[l.pos] != '}' {
		if l.runes[l.pos] == '\r' || l.runes[l.pos] == '\n' {
			l.scanNewline()
			continue
		}
		l.bump()
	}
	if !l.eof() && l.runes[l.pos] == '}' {
		l.bump()
	}
	return Trivia{Kind: TriviaBlockComment, Span: l.spanFrom(start), Text: l.slice(start.pos, l.pos)}
}

func (l *lexer) scanToken() Token {
	start := l.markAt()
	r := l.runes[l.pos]
	switch r {
	case '{', '}', '(', ')', ',', ';', '*', '/', '+', '.', '-':
		l.bump()
		return l.make(TokSymbol, start)
	case ':':
		l.bump()
		if !l.eof() && l.runes[l.pos] == '=' {
			l.bump()
		}
		return l.make(TokSymbol, start)
	case '=':
		l.bump()
		return l.make(TokSymbol, start)
	case '<':
		l.bump()
		if !l.eof() && (l.runes[l.pos] == '=' || l.runes[l.pos] == '>') {
			l.bump()
		}
		return l.make(TokSymbol, start)
	case '>':
		l.bump()
		if !l.eof() && l.runes[l.pos] == '=' {
			l.bump()
		}
		return l.make(TokSymbol, start)
	case '\'':
		return l.scanString(start)
	default:
		if isIdentStart(r) {
			return l.scanWord(start)
		}
		if unicode.IsDigit(r) {
			return l.scanNumber(start)
		}
		l.bump()
		return l.make(TokError, start)
	}
}

func (l *lexer) scanString(start mark) Token {
	l.bump() // opening '
	closed := false
	for !l.eof() {
		r := l.runes[l.pos]
		if r == '\n' || r == '\r' {
			break
		}
		if r == '\'' {
			if l.pos+1 < len(l.runes) && l.runes[l.pos+1] == '\'' {
				l.bump()
				l.bump()
				continue
			}
			l.bump()
			closed = true
			break
		}
		l.bump()
	}
	kind := TokString
	if !closed {
		kind = TokError
	}
	return l.make(kind, start)
}

func (l *lexer) scanNumber(start mark) Token {
	for !l.eof() && unicode.IsDigit(l.runes[l.pos]) {
		l.bump()
	}
	return l.make(TokNumber, start)
}

func (l *lexer) scanWord(start mark) Token {
	for !l.eof() && isIdentContinueLetter(l.runes[l.pos]) {
		l.bump()
	}
	letters := l.slice(start.pos, l.pos)
	if l.tryHyphenKeyword(letters) {
		return l.make(TokKeyword, start)
	}
	// DIV1 is the keyword DIV followed by the number 1. A non-keyword
	// such as ErrorText001 keeps its trailing digits.
	if !l.eof() && unicode.IsDigit(l.runes[l.pos]) && l.isOperatorKeyword(letters) && l.isKeyword(letters) {
		return l.make(TokKeyword, start)
	}
	for !l.eof() && unicode.IsDigit(l.runes[l.pos]) {
		l.bump()
	}
	word := l.slice(start.pos, l.pos)
	kind := TokIdentifier
	if l.isKeyword(word) {
		kind = TokKeyword
	}
	return l.make(kind, start)
}

func (l *lexer) tryHyphenKeyword(letters string) bool {
	if l.mode != modeStructure || l.eof() || l.runes[l.pos] != '-' {
		return false
	}
	saved := l.markAt()
	l.bump()
	wordStart := l.pos
	if l.eof() || !isIdentStart(l.runes[l.pos]) {
		l.restore(saved)
		return false
	}
	for !l.eof() && isIdentContinueLetter(l.runes[l.pos]) {
		l.bump()
	}
	combined := letters + "-" + l.slice(wordStart, l.pos)
	if structureKeywords[strings.ToUpper(combined)] {
		return true
	}
	l.restore(saved)
	return false
}

func (l *lexer) make(kind TokenKind, start mark) Token {
	return Token{Kind: kind, Text: l.slice(start.pos, l.pos), Span: l.spanFrom(start)}
}

func (l *lexer) isKeyword(word string) bool {
	key := strings.ToUpper(word)
	if l.mode == modeCode || l.mode == modeDeclaration {
		return codeKeywords[key]
	}
	return structureKeywords[key]
}

func (l *lexer) isOperatorKeyword(word string) bool {
	switch strings.ToUpper(word) {
	case "AND", "OR", "XOR", "DIV", "MOD":
		return true
	default:
		return false
	}
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentContinueLetter(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

var structureKeywords = map[string]bool{
	"OBJECT": true, "OBJECT-PROPERTIES": true, "PROPERTIES": true,
	"FIELDS": true, "KEYS": true, "FIELDGROUPS": true, "CODE": true,
	"CONTROLS": true, "DATASET": true, "REQUESTPAGE": true, "LABELS": true,
	"RDLDATA": true, "ELEMENTS": true, "EVENTS": true, "MENUNODES": true,
	"BEGIN": true, "END": true, "VAR": true,
	"TABLE": true, "CODEUNIT": true, "PAGE": true, "REPORT": true,
	"XMLPORT": true, "QUERY": true, "MENUSUITE": true,
}

var codeKeywords = func() map[string]bool {
	m := make(map[string]bool, len(structureKeywords)+32)
	for k, v := range structureKeywords {
		m[k] = v
	}
	for _, k := range []string{
		"IF", "THEN", "ELSE", "CASE", "OF", "FOR", "TO", "DOWNTO", "DO",
		"WHILE", "REPEAT", "UNTIL", "WITH", "EXIT",
		"AND", "OR", "XOR", "DIV", "MOD", "NOT",
		"PROCEDURE", "LOCAL", "TRUE", "FALSE", "ASSERTERROR", "TEMPORARY", "ARRAY",
	} {
		m[k] = true
	}
	return m
}()
