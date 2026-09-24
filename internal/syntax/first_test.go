package syntax_test

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/spesotskiy/CALCodeCop/internal/rules"
	"github.com/spesotskiy/CALCodeCop/internal/syntax"
)

const helloPath = "internal/syntax/testdata/codeunit-hello.txt"

func TestCodeunitHello(t *testing.T) {
	data, err := os.ReadFile("testdata/codeunit-hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte{'\r'}) {
		t.Fatal("line endings must be LF (got CR); check out with .gitattributes eol=lf, or git config core.autocrlf=false")
	}
	if len(data) != 252 {
		t.Fatalf("file is %d bytes, want 252", len(data))
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("file must not start with a BOM")
	}
	if bytes.Count(data, []byte{'\n'}) != 21 {
		t.Fatalf("newline count = %d, want 21", bytes.Count(data, []byte{'\n'}))
	}
	if !utf8.Valid(data) {
		t.Fatal("file is not UTF-8")
	}

	tree := syntax.Parse(string(data), helloPath, syntax.ParseOptions{})
	if len(tree.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", tree.Diagnostics)
	}
	if tree.Path != helloPath {
		t.Fatalf("path = %q", tree.Path)
	}
	if len(tree.Tokens) != 52 {
		dumpTokens(t, tree.Tokens)
		t.Fatalf("tokens = %d, want 52", len(tree.Tokens))
	}
	assertSpan(t, "file", tree.Span(), 1, 1, 21, 2)
	assertSpan(t, "file full", tree.FullSpan(), 1, 1, 22, 1)
	if tree.FirstToken() != 0 || tree.LastToken() != 51 {
		t.Fatalf("file tokens %d–%d", tree.FirstToken(), tree.LastToken())
	}
	if tree.FullSpan().End.Offset != utf8.RuneCountInString(string(data)) {
		t.Fatalf("full span end offset = %d", tree.FullSpan().End.Offset)
	}

	wantTok := []string{
		"OBJECT", "Codeunit", "50000", "Hello", "{",
		"OBJECT-PROPERTIES", "{",
		"Date", "=", "24", ".", "09", ".", "26", ";",
		"Time", "=", "12", ":", "00", ":", "00", ";",
		"Modified", "=", "Yes", ";",
		"Version", "List", "=", ";",
		"}",
		"PROPERTIES", "{",
		"OnRun", "=", "BEGIN",
		"MESSAGE", "(", "'Hello!'", ")", ";",
		"END", ";",
		"}",
		"CODE", "{",
		"BEGIN", "END", ".",
		"}",
		"}",
	}
	if len(wantTok) != 52 {
		t.Fatalf("oracle length %d", len(wantTok))
	}
	for i, want := range wantTok {
		if tree.Tokens[i].Text != want {
			dumpTokens(t, tree.Tokens)
			t.Fatalf("token %d = %q, want %q", i, tree.Tokens[i].Text, want)
		}
	}
	assertSpan(t, "OBJECT", tree.Tokens[0].Span, 1, 1, 1, 7)
	assertSpan(t, "(", tree.Tokens[38].Span, 13, 20, 13, 21)
	assertSpan(t, "string", tree.Tokens[39].Span, 13, 21, 13, 29)
	assertSpan(t, ")", tree.Tokens[40].Span, 13, 29, 13, 30)
	assertSpan(t, "stmt semi", tree.Tokens[41].Span, 13, 30, 13, 31)
	assertSpan(t, "property semi", tree.Tokens[43].Span, 14, 14, 14, 15)
	helloAt := strings.Index(string(data), "'Hello!'")
	if tree.Tokens[39].Span.Start.Offset != helloAt {
		t.Fatalf("string offset = %d, want %d", tree.Tokens[39].Span.Start.Offset, helloAt)
	}

	if len(tree.Objects) != 1 {
		t.Fatalf("objects = %d", len(tree.Objects))
	}
	obj := tree.Objects[0]
	assertSpan(t, "object", obj.Span(), 1, 1, 21, 2)
	assertSpan(t, "object full", obj.FullSpan(), 1, 1, 22, 1)
	if obj.FirstToken() != 0 || obj.LastToken() != 51 {
		t.Fatalf("object tokens %d–%d", obj.FirstToken(), obj.LastToken())
	}
	if obj.Type == nil || obj.Type.Text != "Codeunit" || obj.Type.Kind() != syntax.KindType {
		t.Fatalf("type = %+v", obj.Type)
	}
	assertSpan(t, "type", obj.Type.Span(), 1, 8, 1, 16)
	if obj.Type.FirstToken() != 1 {
		t.Fatalf("type token %d", obj.Type.FirstToken())
	}
	if obj.ID == nil || obj.ID.Text != "50000" || obj.ID.Kind() != syntax.KindID {
		t.Fatalf("id = %+v", obj.ID)
	}
	assertSpan(t, "id", obj.ID.Span(), 1, 17, 1, 22)
	if obj.ID.FirstToken() != 2 {
		t.Fatalf("id token %d", obj.ID.FirstToken())
	}
	if obj.Name == nil || obj.Name.Text != "Hello" || obj.Name.Kind() != syntax.KindName {
		t.Fatalf("name = %+v", obj.Name)
	}
	assertSpan(t, "name", obj.Name.Span(), 1, 23, 1, 28)
	if obj.Name.FirstToken() != 3 {
		t.Fatalf("name token %d", obj.Name.FirstToken())
	}

	date := mustProp(t, obj.ObjectProperties, "Date")
	assertSpan(t, "Date", date.Span(), 5, 5, 5, 19)
	assertSpan(t, "Date full", date.FullSpan(), 5, 1, 6, 1)
	if date.FirstToken() != 7 || date.LastToken() != 14 {
		t.Fatalf("Date tokens %d–%d", date.FirstToken(), date.LastToken())
	}
	dateLit := mustLit(t, date.Value)
	if dateLit.Text != "24.09.26" || dateLit.FirstToken() != 9 || dateLit.LastToken() != 13 {
		t.Fatalf("Date literal %+v tokens %d–%d", dateLit.Text, dateLit.FirstToken(), dateLit.LastToken())
	}
	assertSpan(t, "Date value", dateLit.Span(), 5, 10, 5, 18)

	timeProp := mustProp(t, obj.ObjectProperties, "Time")
	assertSpan(t, "Time", timeProp.Span(), 6, 5, 6, 19)
	if timeProp.FirstToken() != 15 || timeProp.LastToken() != 22 {
		t.Fatalf("Time tokens %d–%d", timeProp.FirstToken(), timeProp.LastToken())
	}
	timeLit := mustLit(t, timeProp.Value)
	if timeLit.Text != "12:00:00" || timeLit.FirstToken() != 17 || timeLit.LastToken() != 21 {
		t.Fatalf("Time literal %q tokens %d–%d", timeLit.Text, timeLit.FirstToken(), timeLit.LastToken())
	}
	assertSpan(t, "Time value", timeLit.Span(), 6, 10, 6, 18)

	mod := mustProp(t, obj.ObjectProperties, "Modified")
	assertSpan(t, "Modified", mod.Span(), 7, 5, 7, 18)
	if mod.FirstToken() != 23 || mod.LastToken() != 26 {
		t.Fatalf("Modified tokens %d–%d", mod.FirstToken(), mod.LastToken())
	}
	modLit := mustLit(t, mod.Value)
	if modLit.Text != "Yes" || modLit.FirstToken() != 25 {
		t.Fatalf("Modified literal %q token %d", modLit.Text, modLit.FirstToken())
	}
	assertSpan(t, "Modified value", modLit.Span(), 7, 14, 7, 17)

	ver := mustProp(t, obj.ObjectProperties, "Version List")
	assertSpan(t, "Version List", ver.Span(), 8, 5, 8, 19)
	if ver.FirstToken() != 27 || ver.LastToken() != 30 {
		t.Fatalf("Version List tokens %d–%d", ver.FirstToken(), ver.LastToken())
	}
	verLit := mustLit(t, ver.Value)
	if verLit.Text != "" || verLit.FirstToken() != -1 || verLit.LastToken() != -1 {
		t.Fatalf("Version List literal %q tokens %d–%d", verLit.Text, verLit.FirstToken(), verLit.LastToken())
	}
	assertSpan(t, "Version List value", verLit.Span(), 8, 18, 8, 18)

	onRun := mustProp(t, obj.Properties, "OnRun")
	assertSpan(t, "OnRun", onRun.Span(), 12, 5, 14, 15)
	if onRun.FirstToken() != 34 || onRun.LastToken() != 43 {
		t.Fatalf("OnRun tokens %d–%d", onRun.FirstToken(), onRun.LastToken())
	}
	cv, ok := onRun.Value.(*syntax.CodeValue)
	if !ok {
		t.Fatalf("OnRun value %T", onRun.Value)
	}
	if len(cv.Locals) != 0 {
		t.Fatalf("OnRun locals %d", len(cv.Locals))
	}
	assertSpan(t, "OnRun code", cv.Span(), 12, 11, 14, 14)
	if cv.FirstToken() != 36 || cv.LastToken() != 42 {
		t.Fatalf("OnRun code tokens %d–%d", cv.FirstToken(), cv.LastToken())
	}
	if cv.Body == nil {
		t.Fatal("missing OnRun block")
	}
	assertSpan(t, "OnRun block", cv.Body.Span(), 12, 11, 14, 14)
	if cv.Body.FirstToken() != 36 || cv.Body.LastToken() != 42 {
		t.Fatalf("block tokens %d–%d", cv.Body.FirstToken(), cv.Body.LastToken())
	}
	if len(cv.Body.Statements) != 1 {
		t.Fatalf("statements %d", len(cv.Body.Statements))
	}
	call, ok := cv.Body.Statements[0].(*syntax.CallExpr)
	if !ok {
		t.Fatalf("statement %T", cv.Body.Statements[0])
	}
	assertSpan(t, "call", call.Span(), 13, 13, 13, 31)
	if call.FirstToken() != 37 || call.LastToken() != 41 {
		t.Fatalf("call tokens %d–%d", call.FirstToken(), call.LastToken())
	}
	ident, ok := call.Callee.(*syntax.Identifier)
	if !ok || ident.Text != "MESSAGE" || ident.FirstToken() != 37 {
		t.Fatalf("callee %+v", call.Callee)
	}
	assertSpan(t, "MESSAGE", ident.Span(), 13, 13, 13, 20)
	if len(call.Args) != 1 {
		t.Fatalf("args %d", len(call.Args))
	}
	arg := mustLit(t, call.Args[0])
	if arg.Text != "Hello!" || arg.FirstToken() != 39 {
		t.Fatalf("arg %q token %d", arg.Text, arg.FirstToken())
	}
	assertSpan(t, "Hello!", arg.Span(), 13, 21, 13, 29)
	if _, ok := call.CommaTok(0); ok {
		t.Fatal("hello call has a comma")
	}

	if len(obj.Members) != 1 {
		t.Fatalf("members %d", len(obj.Members))
	}
	code, ok := obj.Members[0].(*syntax.CodeSection)
	if !ok {
		t.Fatalf("member %T", obj.Members[0])
	}
	assertSpan(t, "CODE", code.Span(), 16, 3, 20, 4)
	if code.FirstToken() != 45 || code.LastToken() != 50 {
		t.Fatalf("CODE tokens %d–%d", code.FirstToken(), code.LastToken())
	}
	if len(code.Globals) != 0 || len(code.Procedures) != 0 {
		t.Fatalf("globals %d procedures %d", len(code.Globals), len(code.Procedures))
	}
	if code.ObjectTrigger == nil || code.ObjectTrigger.Body == nil {
		t.Fatal("missing object trigger")
	}
	if len(code.ObjectTrigger.Locals) != 0 || len(code.ObjectTrigger.Body.Statements) != 0 {
		t.Fatal("object trigger is not empty")
	}
	assertSpan(t, "trigger", code.ObjectTrigger.Span(), 18, 5, 19, 9)
	assertSpan(t, "trigger block", code.ObjectTrigger.Body.Span(), 18, 5, 19, 9)
	if code.ObjectTrigger.FirstToken() != 47 || code.ObjectTrigger.LastToken() != 49 {
		t.Fatalf("trigger tokens %d–%d", code.ObjectTrigger.FirstToken(), code.ObjectTrigger.LastToken())
	}
	if len(obj.OpaqueSections) != 0 {
		t.Fatalf("opaque %d", len(obj.OpaqueSections))
	}

	var forbidden []string
	var walk func(syntax.Node)
	walk = func(n syntax.Node) {
		if n == nil {
			return
		}
		switch n.Kind() {
		case syntax.KindProcedure, syntax.KindVariable, syntax.KindParen, syntax.KindError:
			forbidden = append(forbidden, n.Kind().String())
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(tree)
	if len(forbidden) != 0 {
		t.Fatalf("unexpected nodes: %s", strings.Join(forbidden, ", "))
	}

	findings := rules.Run(tree, []rules.Rule{rules.NewRule001(), rules.NewRule002()})
	if len(findings) != 0 {
		t.Fatalf("Rule findings = %d, want 0: %+v", len(findings), findings)
	}
}

func TestRule001BadSamples(t *testing.T) {
	hello, err := os.ReadFile("testdata/codeunit-hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		file     string
		count    int
		spans    []string
		messages []string
	}{
		{"assign-no-space.txt", 1, []string{":="}, []string{"Expected one space on each side of ':='."}},
		{"assign-space-after.txt", 1, []string{":="}, []string{"Expected one space on each side of ':='."}},
		{"assign-space-before.txt", 1, []string{":="}, []string{"Expected one space on each side of ':='."}},
		{"plus-tight.txt", 1, []string{"+"}, []string{"Expected one space on each side of '+'."}},
		{"plus-space-before.txt", 1, []string{"+"}, []string{"Expected one space on each side of '+'."}},
		{"plus-space-after.txt", 1, []string{"+"}, []string{"Expected one space on each side of '+'."}},
		{"plus-extra-spaces.txt", 1, []string{"+"}, []string{"Expected one space on each side of '+'."}},
		{"compare-and-tight.txt", 3, []string{">", "AND", "<"}, []string{
			"Expected one space on each side of '>'.",
			"Expected one space on each side of 'AND'.",
			"Expected one space on each side of '<'.",
		}},
		{"mul-div-tight.txt", 2, []string{"*", "DIV"}, []string{
			"Expected one space on each side of '*'.",
			"Expected one space on each side of 'DIV'.",
		}},
		{"comma-no-space-after.txt", 1, []string{","}, []string{"Expected a single space after ','."}},
		{"comma-both-sides.txt", 2, []string{",", ","}, []string{
			"Unexpected space before ','.",
			"Expected a single space after ','.",
		}},
		{"comma-space-before.txt", 1, []string{","}, []string{"Unexpected space before ','."}},
		{"comma-two-spaces-after.txt", 1, []string{","}, []string{"Expected a single space after ','."}},
		{"comma-tab-after.txt", 1, []string{","}, []string{"Expected a single space after ','."}},
	}
	total := 0
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			path := "testdata/rule001/" + tc.file
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(data, hello[:bytes.Index(hello, []byte("MESSAGE"))]) {
				t.Fatal("sample is not a copy of the codeunit")
			}
			if bytes.Contains(data, []byte("MESSAGE('Hello!');")) {
				t.Fatal("sample still contains the hello call")
			}
			tree := syntax.Parse(string(data), "internal/syntax/"+path, syntax.ParseOptions{})
			if len(tree.Diagnostics) != 0 {
				t.Fatalf("diagnostics = %+v", tree.Diagnostics)
			}
			findings := rules.Run(tree, []rules.Rule{rules.NewRule001()})
			if len(findings) != tc.count {
				t.Fatalf("findings = %d, want %d: %+v", len(findings), tc.count, findings)
			}
			got := make([]string, len(findings))
			gotMsg := make([]string, len(findings))
			for i, f := range findings {
				if f.RuleID != "001" {
					t.Fatalf("rule id %q", f.RuleID)
				}
				if f.Severity != rules.SeverityError {
					t.Fatalf("severity %d", f.Severity)
				}
				if len(f.Related) != 0 {
					t.Fatalf("related spans %+v", f.Related)
				}
				text := spanText(string(data), f.Span)
				got[i] = text
				gotMsg[i] = f.Message
				if !messageOK(f, text) {
					t.Fatalf("message %q for %q", f.Message, text)
				}
			}
			sort.Strings(got)
			want := append([]string(nil), tc.spans...)
			sort.Strings(want)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("spans %q, want %q", got, want)
			}
			sort.Strings(gotMsg)
			wantMsg := append([]string(nil), tc.messages...)
			sort.Strings(wantMsg)
			if strings.Join(gotMsg, "\n") != strings.Join(wantMsg, "\n") {
				t.Fatalf("messages %q, want %q", gotMsg, wantMsg)
			}
		})
		total += tc.count
	}
	if len(cases) != 14 {
		t.Fatalf("samples %d", len(cases))
	}
	if total != 18 {
		t.Fatalf("total findings %d, want 18", total)
	}
}

func TestRule002BadSamples(t *testing.T) {
	hello, err := os.ReadFile("testdata/codeunit-hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		file    string
		count   int
		spans   []string
		message string
	}{
		{"unary-minus-space.txt", 1, []string{"-"}, "Unexpected space between unary '-' and its operand."},
		{"unary-minus-paren-space.txt", 1, []string{"-"}, "Unexpected space between unary '-' and its operand."},
		{"unary-plus-space.txt", 1, []string{"+"}, "Unexpected space between unary '+' and its operand."},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			path := "testdata/rule002/" + tc.file
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(data, hello[:bytes.Index(hello, []byte("MESSAGE"))]) {
				t.Fatal("sample is not a copy of the codeunit")
			}
			if bytes.Contains(data, []byte("MESSAGE('Hello!');")) {
				t.Fatal("sample still contains the hello call")
			}
			tree := syntax.Parse(string(data), "internal/syntax/"+path, syntax.ParseOptions{})
			if len(tree.Diagnostics) != 0 {
				t.Fatalf("diagnostics = %+v", tree.Diagnostics)
			}
			findings := rules.Run(tree, []rules.Rule{rules.NewRule002()})
			if len(findings) != tc.count {
				t.Fatalf("findings = %d, want %d: %+v", len(findings), tc.count, findings)
			}
			for i, f := range findings {
				if f.RuleID != "002" {
					t.Fatalf("rule id %q", f.RuleID)
				}
				if f.Severity != rules.SeverityError {
					t.Fatalf("severity %d", f.Severity)
				}
				if len(f.Related) != 0 {
					t.Fatalf("related spans %+v", f.Related)
				}
				text := spanText(string(data), f.Span)
				if text != tc.spans[i] {
					t.Fatalf("span text %q, want %q", text, tc.spans[i])
				}
				if f.Message != tc.message {
					t.Fatalf("message %q, want %q", f.Message, tc.message)
				}
			}
		})
	}
}

func TestRule002OkSamples(t *testing.T) {
	hello, err := os.ReadFile("testdata/codeunit-hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	files := []string{
		"unary-minus-ok.txt",
		"unary-minus-paren-ok.txt",
		"unary-plus-ok.txt",
		"binary-minus-not-002.txt",
		"not-keyword-ignored.txt",
		"not-paren-ignored.txt",
		"compare-unary-ok.txt",
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			path := "testdata/rule002/" + name
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(data, hello[:bytes.Index(hello, []byte("MESSAGE"))]) {
				t.Fatal("sample is not a copy of the codeunit")
			}
			if bytes.Contains(data, []byte("MESSAGE('Hello!');")) {
				t.Fatal("sample still contains the hello call")
			}
			tree := syntax.Parse(string(data), "internal/syntax/"+path, syntax.ParseOptions{})
			if len(tree.Diagnostics) != 0 {
				t.Fatalf("diagnostics = %+v", tree.Diagnostics)
			}
			findings := rules.Run(tree, []rules.Rule{rules.NewRule002()})
			if len(findings) != 0 {
				t.Fatalf("findings = %d, want 0: %+v", len(findings), findings)
			}
		})
	}
}

func TestUnaryParse(t *testing.T) {
	data, err := os.ReadFile("testdata/rule002/unary-minus-ok.txt")
	if err != nil {
		t.Fatal(err)
	}
	tree := syntax.Parse(string(data), "internal/syntax/testdata/rule002/unary-minus-ok.txt", syntax.ParseOptions{})
	if len(tree.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", tree.Diagnostics)
	}
	onRun := mustProp(t, tree.Objects[0].Properties, "OnRun")
	cv := onRun.Value.(*syntax.CodeValue)
	as, ok := cv.Body.Statements[0].(*syntax.AssignStmt)
	if !ok {
		t.Fatalf("statement %T", cv.Body.Statements[0])
	}
	u, ok := as.Right.(*syntax.UnaryExpr)
	if !ok {
		t.Fatalf("right %T", as.Right)
	}
	if u.OpKind != syntax.UnaryMinus {
		t.Fatalf("op kind %d", u.OpKind)
	}
	id, ok := u.X.(*syntax.Identifier)
	if !ok || id.Text != "Amount" {
		t.Fatalf("operand %+v", u.X)
	}
	if tree.Tokens[u.Op].Text != "-" {
		t.Fatalf("op token %q", tree.Tokens[u.Op].Text)
	}
}

func messageOK(f rules.Finding, text string) bool {
	switch f.Message {
	case "Unexpected space before ','.":
		return text == ","
	case "Expected a single space after ','.":
		return text == ","
	default:
		return f.Message == fmt.Sprintf("Expected one space on each side of '%s'.", text)
	}
}

func spanText(src string, sp syntax.SourceSpan) string {
	runes := []rune(src)
	if sp.Start.Offset < 0 || sp.End.Offset > len(runes) || sp.End.Offset < sp.Start.Offset {
		return ""
	}
	return string(runes[sp.Start.Offset:sp.End.Offset])
}

func mustProp(t *testing.T, list *syntax.PropertyList, name string) *syntax.Property {
	t.Helper()
	if list == nil {
		t.Fatalf("missing property list for %s", name)
	}
	for _, p := range list.Properties {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("missing property %s", name)
	return nil
}

func mustLit(t *testing.T, n syntax.Node) *syntax.Literal {
	t.Helper()
	lit, ok := n.(*syntax.Literal)
	if !ok {
		t.Fatalf("got %T, want literal", n)
	}
	return lit
}

func assertSpan(t *testing.T, name string, got syntax.SourceSpan, sl, sc, el, ec int) {
	t.Helper()
	if got.Start.Line != sl || got.Start.Column != sc || got.End.Line != el || got.End.Column != ec {
		t.Fatalf("%s span = %d:%d–%d:%d, want %d:%d–%d:%d (offsets %d–%d)",
			name, got.Start.Line, got.Start.Column, got.End.Line, got.End.Column, sl, sc, el, ec,
			got.Start.Offset, got.End.Offset)
	}
}

func dumpTokens(t *testing.T, toks []syntax.Token) {
	t.Helper()
	for _, tok := range toks {
		t.Logf("%d %q %d:%d–%d:%d", tok.Index, tok.Text, tok.Span.Start.Line, tok.Span.Start.Column, tok.Span.End.Line, tok.Span.End.Column)
	}
}
