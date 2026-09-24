package rules

import (
	"github.com/spesotskiy/CALCodeCop/internal/syntax"
)

// Rule003 checks that indexing brackets and option-access '::' have no
// surrounding trivia. Severity stays Error.
type Rule003 struct{}

// NewRule003 returns Rule 003.
func NewRule003() *Rule003 { return &Rule003{} }

// ID is the stable rule id.
func (r *Rule003) ID() string { return "003" }

// Analyze reports one finding per forbidden gap beside '[', ']', or '::'.
func (r *Rule003) Analyze(ctx *Context) {
	if ctx == nil || ctx.File == nil {
		return
	}
	Walk(ctx.File, func(n syntax.Node) {
		switch n := n.(type) {
		case *syntax.IndexExpr:
			r.checkIndex(ctx, n)
		case *syntax.OptionAccessExpr:
			r.checkOption(ctx, n)
		}
	})
}

func (r *Rule003) checkIndex(ctx *Context, idx *syntax.IndexExpr) {
	if idx == nil || !operandOK(idx.X) {
		return
	}
	lbrack, okL := tokAt(ctx.Tokens, idx.Lbrack)
	rbrack, okR := tokAt(ctx.Tokens, idx.Rbrack)
	if !okL || !okR {
		return
	}
	recv, okRecv := tokAt(ctx.Tokens, idx.X.LastToken())
	if okRecv && !IsEmpty(Gap(recv, lbrack)) {
		ctx.Report(Finding{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Unexpected space around '[' or ']'.",
			Span:     lbrack.Span,
		})
	}
	if len(idx.Args) == 0 {
		if !IsEmpty(Gap(lbrack, rbrack)) {
			ctx.Report(Finding{
				RuleID:   r.ID(),
				Severity: SeverityError,
				Message:  "Unexpected space around '[' or ']'.",
				Span:     lbrack.Span,
			})
		}
		return
	}
	firstArg := idx.Args[0]
	lastArg := idx.Args[len(idx.Args)-1]
	if !operandOK(firstArg) || !operandOK(lastArg) {
		return
	}
	firstTok, okF := tokAt(ctx.Tokens, firstArg.FirstToken())
	lastTok, okLast := tokAt(ctx.Tokens, lastArg.LastToken())
	if !okF || !okLast {
		return
	}
	if !IsEmpty(Gap(lbrack, firstTok)) {
		ctx.Report(Finding{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Unexpected space around '[' or ']'.",
			Span:     lbrack.Span,
		})
	}
	if !IsEmpty(Gap(lastTok, rbrack)) {
		ctx.Report(Finding{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Unexpected space around '[' or ']'.",
			Span:     rbrack.Span,
		})
	}
}

func (r *Rule003) checkOption(ctx *Context, oa *syntax.OptionAccessExpr) {
	if oa == nil || !operandOK(oa.X) || !operandOK(oa.Sel) {
		return
	}
	cc, ok := tokAt(ctx.Tokens, oa.ColonColon)
	if !ok {
		return
	}
	left, okL := tokAt(ctx.Tokens, oa.X.LastToken())
	right, okR := tokAt(ctx.Tokens, oa.Sel.FirstToken())
	if !okL || !okR {
		return
	}
	if !IsEmpty(Gap(left, cc)) {
		ctx.Report(Finding{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Unexpected space around '::'.",
			Span:     cc.Span,
		})
	}
	if !IsEmpty(Gap(cc, right)) {
		ctx.Report(Finding{
			RuleID:   r.ID(),
			Severity: SeverityError,
			Message:  "Unexpected space around '::'.",
			Span:     cc.Span,
		})
	}
}
