package syntax

// ParseOptions configures Parse. The first tests use the zero value.
type ParseOptions struct{}

// Parse builds an AST from UTF-8 object text. text is the file with a leading
// UTF-8 BOM already stripped. path is stored on the tree for later findings.
func Parse(text string, path string, _ ParseOptions) *SyntaxTree {
	file := &ObjectFile{Path: path}
	file.kind = KindObjectFile
	file.Diagnostics = []Diagnostic{}
	p := &parser{lex: newLexer(text), file: file}
	p.parseFile()
	p.lex.finish()
	file.Tokens = p.lex.tokens
	if len(file.Tokens) > 0 {
		file.first = 0
		file.last = len(file.Tokens) - 1
	} else {
		file.explicit = true
		file.first = -1
		file.last = -1
		file.span = SourceSpan{
			Start: SourcePosition{Line: 1, Column: 1},
			End:   SourcePosition{Line: 1, Column: 1},
		}
	}
	finalize(file, nil, file.Tokens)
	return file
}

func finalize(n Node, parent Node, tokens []Token) {
	if n == nil {
		return
	}
	n.SetParent(parent)
	if n.explicitSpan() {
		sp := n.Span()
		n.setSpans(sp, sp)
	} else if first, last := n.tokenRange(); first >= 0 && last >= first && last < len(tokens) {
		span := SourceSpan{Start: tokens[first].Span.Start, End: tokens[last].Span.End}
		full := SourceSpan{Start: spanStart(tokens[first]), End: spanEnd(tokens[last])}
		n.setSpans(span, full)
	}
	for _, child := range n.Children() {
		finalize(child, n, tokens)
	}
}

func spanStart(tok Token) SourcePosition {
	if len(tok.LeadingTrivia) == 0 {
		return tok.Span.Start
	}
	return tok.LeadingTrivia[0].Span.Start
}

func spanEnd(tok Token) SourcePosition {
	if len(tok.TrailingTrivia) == 0 {
		return tok.Span.End
	}
	return tok.TrailingTrivia[len(tok.TrailingTrivia)-1].Span.End
}
