[![status][ci-status-badge]][ci-status]
[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev]

# onion

onion find layering violations.

## Synopsis

```sh
go install github.com/aereal/onion@latest
go vet -vettool=$(which onion) -onion.config=path/to/onion.json ./...
```

See testdata/config.json.

## License

See LICENSE file.

[pkg-go-dev]: https://pkg.go.dev/github.com/aereal/onion
[pkg-go-dev-badge]: https://pkg.go.dev/badge/aereal/onion
[ci-status-badge]: https://github.com/aereal/onion/workflows/CI/badge.svg?branch=main
[ci-status]: https://github.com/aereal/onion/actions/workflows/CI
