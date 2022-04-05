package main

import (
	"github.com/aereal/onion"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(onion.Analyzer)
}
