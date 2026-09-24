# CALCodeCop

CALCodeCop reads a Dynamics NAV 2017 C/AL object text file and checks style. This slice parses a codeunit and runs Rule 001 (operator and comma spacing).

The module path is `github.com/spesotskiy/CALCodeCop`. Go 1.24 or newer.

```
go test ./...
```

`internal/syntax` parses UTF-8 object text into an AST with a retained token stream. `internal/rules` runs Rule 001 on that tree. The check is `go test ./...`.
