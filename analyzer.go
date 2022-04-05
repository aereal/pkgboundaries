package onion

import (
	"encoding/json"
	"go/ast"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "onion",
	Doc:  "check import direction",
	Run:  run,
}

var configPath string

func init() {
	Analyzer.Flags.StringVar(&configPath, "config", "onion.json", "config file path")
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
	if strings.HasSuffix(currentPkg, ".test") {
		return nil, nil
	}
	var cfg *Config
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	currentLayer := (*LayersSet)(cfg.Layers).findByPackagePath(pass.Pkg.Path())
	if currentLayer == nil {
		return nil, nil
	}

	processFile := func(file *ast.File) {
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}
			decision := cfg.CanDepend(currentLayer.Name, Package(importPath))
			if decision == DecisionDeny {
				pass.Reportf(spec.Pos(), "%s cannot be imported by %s", spec.Path.Value, currentLayer.Name)
			}
		}
	}
	for _, file := range pass.Files {
		processFile(file)
	}
	return nil, nil
}
