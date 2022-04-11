package main

import (
	"strings"

	"github.com/aereal/pkgboundaries/analyzer"
	"golang.org/x/tools/go/analysis"
)

var flags string

type analyzerPlugin int

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	if flags != "" {
		if err := analyzer.Analyzer.Flags.Parse(strings.Split(flags, " ")); err != nil {
			panic("failed to parse: " + err.Error())
		}
	}
	return []*analysis.Analyzer{
		analyzer.Analyzer,
	}
}
