[![status][ci-status-badge]][ci-status]
[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev]

# pkgboundaries

pkgboundaries find layering violations.

## Synopsis

```sh
go install github.com/aereal/pkgboundaries@latest
go vet -vettool=$(which pkgboundaries) -pkgboundaries.config=path/to/pkgboundaries.json ./...
```

See testdata/config.json.

## License

See LICENSE file.

[pkg-go-dev]: https://pkg.go.dev/github.com/aereal/pkgboundaries
[pkg-go-dev-badge]: https://pkg.go.dev/badge/aereal/pkgboundaries
[ci-status-badge]: https://github.com/aereal/pkgboundaries/workflows/CI/badge.svg?branch=main
[ci-status]: https://github.com/aereal/pkgboundaries/actions/workflows/CI
