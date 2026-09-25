package rules

import (
	"fmt"

	"github.com/spesotskiy/CALCodeCop/internal/syntax"
)

// Rule002 checks that symbolic unary operators glue to their operand.
// NOT is parsed as unary but is not enforced here. Severity stays Error.
type Rule002 struct{}

// NewRule002 returns Rule 002.
func NewRule002() *Rule002 { return &Rule002{} }

// ID is the stable rule id.
func (r *Rule002) ID() string { return "002" }

// Analyze reports one finding per unary '-' or '+' with a non-empty gap
// before its operand.
func (r *Rule002) Analyze(ctx *Context) {
	if ctx == nil || ctx.File == nil {
		return
	}
	Walk(ctx.File, func(n syntax.Node) {
		u, ok := n.(*syntax.UnaryExpr)
		if !ok {
			return
		}
		if u.OpKind == syntax.UnaryNot {
			return
		}
		if u.X == nil || u.X.Kind() == syntax.KindError {
			return
		}
		if u.Op < 0 || u.Op >= len(ctx.Tokens) {
			return
		}
		op := ctx.Tokens[u.Op]
		operandTok, ok := tokAt(ctx.Tokens, u.X.FirstToken())
		if !ok {
			return
		}
		if IsEmpty(Gap(op, operandTok)) {
			return
		}
		ctx.Report(Finding{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  fmt.Sprintf("Unexpected space between unary '%s' and its operand.", op.Text),
			Span:     op.Span,
		})
	})
}
