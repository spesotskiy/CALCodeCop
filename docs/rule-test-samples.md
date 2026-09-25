# Rule test samples

Minimal wrong/good snippets taken from `internal/syntax/testdata/rule00*` fixtures used by `TestRule001*` / `TestRule002*` / `TestRule003*` in `internal/syntax/first_test.go`. Not copied from the rule spec docs.

Each fixture is a full codeunit; the table shows only the OnRun statement line under test. Source path: `internal/syntax/testdata/rule00N/<fixture>.txt`.

Ok fixtures have no violation; the Wrong column is `—`.

## Rule 001 — operator and comma spacing

| Wrong | Good | Description |
| --- | --- | --- |
| `Amount:=Price;` | `Amount := Price;` | `assign-no-space.txt` — no spaces around `:=` |
| `Amount:= Price;` | `Amount := Price;` | `assign-space-after.txt` — missing space before `:=` |
| `Amount :=Price;` | `Amount := Price;` | `assign-space-before.txt` — missing space after `:=` |
| `A+B` | `A + B` | `plus-tight.txt` — no spaces around `+` |
| `A +B` | `A + B` | `plus-space-before.txt` — missing space after `+` |
| `A+ B` | `A + B` | `plus-space-after.txt` — missing space before `+` |
| `A  +  B` | `A + B` | `plus-extra-spaces.txt` — more than one space on each side of `+` |
| `(A>0)AND(B<10)` | `(A > 0) AND (B < 10)` | `compare-and-tight.txt` — tight `>`, `AND`, `<` (3 findings) |
| `Qty*UnitCost DIV1` | `Qty * UnitCost DIV 1` | `mul-div-tight.txt` — tight `*` and `DIV` (2 findings) |
| `A<>B;` | `A <> B;` | `neq-tight.txt` — no spaces around `<>` |
| `Func(Param1,Param2);` | `Func(Param1, Param2);` | `comma-no-space-after.txt` — missing space after `,` |
| `Func(Param1 ,Param2);` | `Func(Param1, Param2);` | `comma-both-sides.txt` — space before `,` and missing space after (2 findings) |
| `Func(Param1 , Param2);` | `Func(Param1, Param2);` | `comma-space-before.txt` — unexpected space before `,` |
| `Func(Param1,  Param2);` | `Func(Param1, Param2);` | `comma-two-spaces-after.txt` — two spaces after `,` |
| `Func(Param1,	Param2);` | `Func(Param1, Param2);` | `comma-tab-after.txt` — tab after `,` (shown as a tab between comma and `Param2`) |
| — | `Amount := Price + Tax;` | `assign-spaces-ok.txt` — compliant `:=` / `+` |
| — | `(A > 0) AND (B < 10);` | `binary-and-ok.txt` — compliant comparisons and `AND` |
| — | `MESSAGE('%1', Name);` | `call-comma-ok.txt` — compliant call commas |
| — | `NOT Flag;` | `unary-not.txt` — `NOT` is not a binary op for 001 |
| — | `MESSAGE('a,b');` | `comma-in-string.txt` — comma inside string ignored |

## Rule 002 — unary operator spacing

| Wrong | Good | Description |
| --- | --- | --- |
| `Total := - Amount;` | `Total := -Amount;` | `unary-minus-space.txt` — space after unary `-` |
| `Delta := - (A + B);` | `Delta := -(A + B);` | `unary-minus-paren-space.txt` — space between `-` and `(` |
| `N := + 1;` | `N := +1;` | `unary-plus-space.txt` — space after unary `+` |
| — | `Total := -Amount;` | `unary-minus-ok.txt` — unary `-` glued to operand |
| — | `Delta := -(A + B);` | `unary-minus-paren-ok.txt` — unary `-` glued to `(` |
| — | `N := +1;` | `unary-plus-ok.txt` — unary `+` glued to operand |
| — | `Total := A - B;` | `binary-minus-not-002.txt` — binary `-` not Rule 002 |
| — | `NOT Found;` | `not-keyword-ignored.txt` — `NOT` not enforced by 002 |
| — | `NOT(Found);` | `not-paren-ignored.txt` — `NOT(` not enforced by 002 |
| — | `Ok := Quantity < -1;` | `compare-unary-ok.txt` — compare with unary `-1` |

## Rule 003 — no spaces around `[]` and `::`

| Wrong | Good | Description |
| --- | --- | --- |
| `Qty := Arr [i];` | `Qty := Arr[i];` | `index-space-before-bracket.txt` — space before `[` |
| `Qty := Arr[ i];` | `Qty := Arr[i];` | `index-space-after-bracket.txt` — space after `[` |
| `Qty := Arr[i ];` | `Qty := Arr[i];` | `index-space-before-rbracket.txt` — space before `]` |
| `Qty := Arr[ i ];` | `Qty := Arr[i];` | `index-space-inside.txt` — spaces inside brackets (2 findings) |
| `Cell := Arr[ x, y ];` | `Cell := Arr[x, y];` | `index-multidim-space-around.txt` — spaces after `[` / before `]` |
| `Type := Choice ::Open;` | `Type := Choice::Open;` | `option-space-before.txt` — space before `::` |
| `Type := Choice:: Open;` | `Type := Choice::Open;` | `option-space-after.txt` — space after `::` |
| `Type := Choice :: Open;` | `Type := Choice::Open;` | `option-spaces.txt` — spaces on both sides of `::` (2 findings) |
| `Type := "Document Type" ::Order;` | `Type := "Document Type"::Order;` | `option-quoted-space.txt` — space before `::` after quoted name |
| — | `Qty := Arr[i];` | `index-ok.txt` — tight brackets |
| — | `Qty := Arr[i + 1];` | `index-plus-ok.txt` — `+` spacing is Rule 001 |
| — | `Cell := Arr[x, y];` | `index-multidim-ok.txt` — multidim index ok for 003 |
| — | `Type := Choice::Open;` | `option-ok.txt` — tight `::` |
| — | `Type := "Document Type"::Order;` | `option-quoted-ok.txt` — quoted option access |

## Coverage note

Rules **004+** are specified under the project store but do not yet have `testdata` fixtures or `TestRule00N*` cases in this repository. Add rows here when those tests land.
