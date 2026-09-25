# Rule test samples

Minimal wrong/good snippets taken from `internal/syntax/testdata/rule00*` fixtures used by `TestRule001*` / `TestRule002*` / `TestRule003*` in `internal/syntax/first_test.go`. Not copied from the rule spec docs.

Each fixture is a full codeunit; the table shows only the OnRun statement line under test. Source path: `internal/syntax/testdata/rule00N/<fixture>.txt`.

Ok fixtures have no violation; the Wrong column is `—`.

## Rule 001 — operator and comma spacing

| Rule | Wrong | Good | Description |
| --- | --- | --- | --- |
| 001 | `Amount:=Price;` | `Amount := Price;` | `assign-no-space.txt` — no spaces around `:=` |
| 001 | `Amount:= Price;` | `Amount := Price;` | `assign-space-after.txt` — missing space before `:=` |
| 001 | `Amount :=Price;` | `Amount := Price;` | `assign-space-before.txt` — missing space after `:=` |
| 001 | `A+B` | `A + B` | `plus-tight.txt` — no spaces around `+` |
| 001 | `A +B` | `A + B` | `plus-space-before.txt` — missing space after `+` |
| 001 | `A+ B` | `A + B` | `plus-space-after.txt` — missing space before `+` |
| 001 | `A  +  B` | `A + B` | `plus-extra-spaces.txt` — more than one space on each side of `+` |
| 001 | `(A>0)AND(B<10)` | `(A > 0) AND (B < 10)` | `compare-and-tight.txt` — tight `>`, `AND`, `<` (3 findings) |
| 001 | `Qty*UnitCost DIV1` | `Qty * UnitCost DIV 1` | `mul-div-tight.txt` — tight `*` and `DIV` (2 findings) |
| 001 | `A<>B;` | `A <> B;` | `neq-tight.txt` — no spaces around `<>` |
| 001 | `Func(Param1,Param2);` | `Func(Param1, Param2);` | `comma-no-space-after.txt` — missing space after `,` |
| 001 | `Func(Param1 ,Param2);` | `Func(Param1, Param2);` | `comma-both-sides.txt` — space before `,` and missing space after (2 findings) |
| 001 | `Func(Param1 , Param2);` | `Func(Param1, Param2);` | `comma-space-before.txt` — unexpected space before `,` |
| 001 | `Func(Param1,  Param2);` | `Func(Param1, Param2);` | `comma-two-spaces-after.txt` — two spaces after `,` |
| 001 | `Func(Param1,	Param2);` | `Func(Param1, Param2);` | `comma-tab-after.txt` — tab after `,` (shown as a tab between comma and `Param2`) |
| 001 | — | `Amount := Price + Tax;` | `assign-spaces-ok.txt` — compliant `:=` / `+` |
| 001 | — | `(A > 0) AND (B < 10);` | `binary-and-ok.txt` — compliant comparisons and `AND` |
| 001 | — | `MESSAGE('%1', Name);` | `call-comma-ok.txt` — compliant call commas |
| 001 | — | `NOT Flag;` | `unary-not.txt` — `NOT` is not a binary op for 001 |
| 001 | — | `MESSAGE('a,b');` | `comma-in-string.txt` — comma inside string ignored |

## Rule 002 — unary operator spacing

| Rule | Wrong | Good | Description |
| --- | --- | --- | --- |
| 002 | `Total := - Amount;` | `Total := -Amount;` | `unary-minus-space.txt` — space after unary `-` |
| 002 | `Delta := - (A + B);` | `Delta := -(A + B);` | `unary-minus-paren-space.txt` — space between `-` and `(` |
| 002 | `N := + 1;` | `N := +1;` | `unary-plus-space.txt` — space after unary `+` |
| 002 | — | `Total := -Amount;` | `unary-minus-ok.txt` — unary `-` glued to operand |
| 002 | — | `Delta := -(A + B);` | `unary-minus-paren-ok.txt` — unary `-` glued to `(` |
| 002 | — | `N := +1;` | `unary-plus-ok.txt` — unary `+` glued to operand |
| 002 | — | `Total := A - B;` | `binary-minus-not-002.txt` — binary `-` not Rule 002 |
| 002 | — | `NOT Found;` | `not-keyword-ignored.txt` — `NOT` not enforced by 002 |
| 002 | — | `NOT(Found);` | `not-paren-ignored.txt` — `NOT(` not enforced by 002 |
| 002 | — | `Ok := Quantity < -1;` | `compare-unary-ok.txt` — compare with unary `-1` |

## Rule 003 — no spaces around `[]` and `::`

| Rule | Wrong | Good | Description |
| --- | --- | --- | --- |
| 003 | `Qty := Arr [i];` | `Qty := Arr[i];` | `index-space-before-bracket.txt` — space before `[` |
| 003 | `Qty := Arr[ i];` | `Qty := Arr[i];` | `index-space-after-bracket.txt` — space after `[` |
| 003 | `Qty := Arr[i ];` | `Qty := Arr[i];` | `index-space-before-rbracket.txt` — space before `]` |
| 003 | `Qty := Arr[ i ];` | `Qty := Arr[i];` | `index-space-inside.txt` — spaces inside brackets (2 findings) |
| 003 | `Cell := Arr[ x, y ];` | `Cell := Arr[x, y];` | `index-multidim-space-around.txt` — spaces after `[` / before `]` |
| 003 | `Type := Choice ::Open;` | `Type := Choice::Open;` | `option-space-before.txt` — space before `::` |
| 003 | `Type := Choice:: Open;` | `Type := Choice::Open;` | `option-space-after.txt` — space after `::` |
| 003 | `Type := Choice :: Open;` | `Type := Choice::Open;` | `option-spaces.txt` — spaces on both sides of `::` (2 findings) |
| 003 | `Type := "Document Type" ::Order;` | `Type := "Document Type"::Order;` | `option-quoted-space.txt` — space before `::` after quoted name |
| 003 | — | `Qty := Arr[i];` | `index-ok.txt` — tight brackets |
| 003 | — | `Qty := Arr[i + 1];` | `index-plus-ok.txt` — `+` spacing is Rule 001 |
| 003 | — | `Cell := Arr[x, y];` | `index-multidim-ok.txt` — multidim index ok for 003 |
| 003 | — | `Type := Choice::Open;` | `option-ok.txt` — tight `::` |
| 003 | — | `Type := "Document Type"::Order;` | `option-quoted-ok.txt` — quoted option access |

## Coverage note

Rules **004+** are specified under the project store but do not yet have `testdata` fixtures or `TestRule00N*` cases in this repository. Add rows here when those tests land.
