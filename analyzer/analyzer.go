package analyzer

import (
	"encoding/json"
	"go/ast"
	"os"
	"strconv"
	"strings"

	"github.com/aereal/pkgboundaries"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "pkgboundaries",
	Doc:  "check package boundaries",
	Run:  run,
}

var (
	configPath       string
	skipTestPackages bool
)

func init() {
	Analyzer.Flags.StringVar(&configPath, "config", "pkgboundaries.json", "config file path")
	Analyzer.Flags.BoolVar(&skipTestPackages, "skip-test", false, "skip validating for test pacakges")
}

func SetConfigPathForTesting(path string) func() {
	orig := configPath
	configPath = path
	return func() {
		configPath = orig
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	currentPkg := pass.Pkg.Path()
	if strings.HasSuffix(currentPkg, ".test") && skipTestPackages {
		return nil, nil
	}
	var cfg *pkgboundaries.Config
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	if cfg.Layers == nil {
		return nil, nil
	}
	currentLayer := (*pkgboundaries.LayersSet)(cfg.Layers).FindByPackagePath(pass.Pkg.Path())
	if currentLayer == nil {
		return nil, nil
	}

	processFile := func(file *ast.File) {
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}
			decision := cfg.CanDepend(currentLayer.Name, pkgboundaries.Package(importPath))
			if decision == pkgboundaries.DecisionDeny {
				pass.Reportf(spec.Pos(), "%s cannot be imported by %s", spec.Path.Value, currentLayer.Name)
			}
		}
	}
	for _, file := range pass.Files {
		if tf := pass.Fset.File(file.Pos()); tf != nil {
			if skipTestPackages && strings.HasSuffix(tf.Name(), "_test.go") {
				continue
			}
		}
		processFile(file)
	}
	return nil, nil
}
