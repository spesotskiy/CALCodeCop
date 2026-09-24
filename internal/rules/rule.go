package rules

import (
	"strings"

	"github.com/spesotskiy/CALCodeCop/internal/syntax"
)

// Severity is the severity of a rule finding.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

// Finding is one rule result. Parse diagnostics stay on the syntax tree.
type Finding struct {
	RuleID   string
	Severity Severity
	Message  string
	Span     syntax.SourceSpan
	Related  []syntax.SourceSpan
}

// Context is the finished tree and token stream a rule reads.
type Context struct {
	File     *syntax.ObjectFile
	Tokens   []syntax.Token
	Findings []Finding
}

// NewContext builds a context from one parse. Rules do not call Parse.
func NewContext(file *syntax.ObjectFile) *Context {
	ctx := &Context{File: file}
	if file != nil {
		ctx.Tokens = file.Tokens
	}
	return ctx
}

// Report records one finding.
func (c *Context) Report(f Finding) {
	if c == nil {
		return
	}
	c.Findings = append(c.Findings, f)
}

// Rule is one style check. ID is stable.
type Rule interface {
	ID() string
	Analyze(ctx *Context)
}

// Run applies each rule to a tree that has already been parsed.
func Run(file *syntax.ObjectFile, ruleSet []Rule) []Finding {
	ctx := NewContext(file)
	for _, rule := range ruleSet {
		if rule == nil {
			continue
		}
		rule.Analyze(ctx)
	}
	return ctx.Findings
}

// Walk visits n and then its children. Error nodes are skipped.
func Walk(n syntax.Node, fn func(syntax.Node)) {
	if n == nil || fn == nil {
		return
	}
	if n.Kind() == syntax.KindError {
		return
	}
	fn(n)
	for _, child := range n.Children() {
		Walk(child, fn)
	}
}

// Gap returns the text strictly between left and right
// (trailing trivia of left plus leading trivia of right).
func Gap(left, right syntax.Token) string {
	var b strings.Builder
	for _, t := range left.TrailingTrivia {
		b.WriteString(t.Text)
	}
	for _, t := range right.LeadingTrivia {
		b.WriteString(t.Text)
	}
	return b.String()
}

// IsExactlyOneSpace reports whether gap is a single U+0020.
func IsExactlyOneSpace(gap string) bool { return gap == " " }

// IsEmpty reports whether gap has no characters.
func IsEmpty(gap string) bool { return len(gap) == 0 }
