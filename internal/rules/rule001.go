package rules

import (
	"fmt"

	"github.com/spesotskiy/CALCodeCop/internal/syntax"
)

// Rule001 checks binary-operator spacing and comma spacing.
// Severity stays Error.
type Rule001 struct{}

// NewRule001 returns Rule 001.
func NewRule001() *Rule001 { return &Rule001{} }

// ID is the stable rule id.
func (r *Rule001) ID() string { return "001" }

// Analyze reports one finding per binary operator whose spacing is wrong,
// and one finding per comma side that fails.
func (r *Rule001) Analyze(ctx *Context) {
	if ctx == nil || ctx.File == nil {
		return
	}
	Walk(ctx.File, func(n syntax.Node) {
		switch n := n.(type) {
		case *syntax.BinaryExpr:
			r.checkAround(ctx, n.Left, n.Op, n.Right)
		case *syntax.AssignStmt:
			r.checkAround(ctx, n.Left, n.Assign, n.Right)
		case *syntax.CallExpr:
			r.checkCommas(ctx, n)
		}
	})
}

func (r *Rule001) checkAround(ctx *Context, left syntax.Node, opIndex int, right syntax.Node) {
	if !operandOK(left) || !operandOK(right) {
		return
	}
	if opIndex < 0 || opIndex >= len(ctx.Tokens) {
		return
	}
	leftTok, okL := tokAt(ctx.Tokens, left.LastToken())
	rightTok, okR := tokAt(ctx.Tokens, right.FirstToken())
	if !okL || !okR {
		return
	}
	op := ctx.Tokens[opIndex]
	if IsExactlyOneSpace(Gap(leftTok, op)) && IsExactlyOneSpace(Gap(op, rightTok)) {
		return
	}
	ctx.Report(Finding{
		RuleID:   r.ID(),
		Severity: SeverityError,
		Message:  fmt.Sprintf("Expected one space on each side of '%s'.", op.Text),
		Span:     op.Span,
	})
}

func (r *Rule001) checkCommas(ctx *Context, call *syntax.CallExpr) {
	if call == nil || len(call.Args) < 2 {
		return
	}
	for i := 0; i < len(call.Args)-1; i++ {
		left := call.Args[i]
		right := call.Args[i+1]
		if !operandOK(left) || !operandOK(right) {
			continue
		}
		comma, ok := call.CommaTok(i)
		if !ok {
			continue
		}
		leftTok, okL := tokAt(ctx.Tokens, left.LastToken())
		rightTok, okR := tokAt(ctx.Tokens, right.FirstToken())
		if !okL || !okR {
			continue
		}
		if !IsEmpty(Gap(leftTok, comma)) {
			ctx.Report(Finding{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Unexpected space before ','.",
				Span:     comma.Span,
			})
		}
		if !IsExactlyOneSpace(Gap(comma, rightTok)) {
			ctx.Report(Finding{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Expected a single space after ','.",
				Span:     comma.Span,
			})
		}
	}
}

func operandOK(n syntax.Node) bool {
	return n != nil && n.Kind() != syntax.KindError
}

func tokAt(tokens []syntax.Token, index int) (syntax.Token, bool) {
	if index < 0 || index >= len(tokens) {
		return syntax.Token{}, false
	}
	return tokens[index], true
}
