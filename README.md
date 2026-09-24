# CALCodeCop

Style checks for Dynamics NAV 2017 (C/AL) object text exports. This slice parses a codeunit and runs Rule 001 (operator and comma spacing) and Rule 002 (unary operator spacing).

## Requirements

- Go 1.24 or newer

## Build and test

```bash
go build ./...
go test ./...
```

Every push and pull request runs the same commands in [GitHub Actions](.github/workflows/build.yml).

Fixtures under `internal/syntax/testdata` are UTF-8 with LF line endings (enforced by `.gitattributes`). On Windows, if `go test` complains about CR or file size, refresh the working tree after pull: `git add --renormalize . && git checkout -- .`

`internal/syntax` parses UTF-8 object text into an AST with a retained token stream. `internal/rules` runs Rule 001 and Rule 002 on that tree.
