# CALCodeCop

CALCodeCop reads a Dynamics NAV 2017 C/AL object text file and checks style. This slice parses a codeunit and runs Rule 001 (operator and comma spacing).

The module path is `github.com/spesotskiy/CALCodeCop`. Go 1.24 or newer.

```
go test ./...
```

Fixtures under `internal/syntax/testdata` are UTF-8 with LF line endings (enforced by `.gitattributes`). On Windows, if `go test` complains about CR or file size, refresh the working tree after pull: `git add --renormalize . && git checkout -- .`

`internal/syntax` parses UTF-8 object text into an AST with a retained token stream. `internal/rules` runs Rule 001 on that tree. The check is `go test ./...`.
