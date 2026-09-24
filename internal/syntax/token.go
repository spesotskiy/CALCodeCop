package syntax

// SourcePosition is a point in a UTF-8 object file.
// Offset is a 0-based rune index. Line and Column are 1-based.
// A column is one rune; a tab is one column.
type SourcePosition struct {
	Offset int
	Line   int
	Column int
}

// SourceSpan is a half-open range [Start, End).
type SourceSpan struct {
	Start SourcePosition
	End   SourcePosition
}

type TokenKind int

const (
	TokEOF TokenKind = iota
	TokKeyword
	TokIdentifier
	TokNumber
	TokString
	TokSymbol
	TokError
)

type TriviaKind int

const (
	TriviaWhitespace TriviaKind = iota
	TriviaEndOfLine
	TriviaLineComment
	TriviaBlockComment
	TriviaSkipped
)

// Trivia is whitespace or a comment attached to a token.
// Text is the exact source characters, including the newline sequence.
type Trivia struct {
	Kind TriviaKind
	Span SourceSpan
	Text string
}

// Token is one lexical element. Layout lives on LeadingTrivia and TrailingTrivia.
type Token struct {
	Kind           TokenKind
	Text           string
	Span           SourceSpan
	LeadingTrivia  []Trivia
	TrailingTrivia []Trivia
	Index          int
}
