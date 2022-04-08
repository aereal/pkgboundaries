package main

import (
	"github.com/aereal/pkgboundaries"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(pkgboundaries.Analyzer)
}
