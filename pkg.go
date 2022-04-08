package pkgboundaries

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aereal/pkgboundaries/internal/sets"
	"github.com/itchyny/rassemble-go"
)

func NewPackagesSet(pkgNames ...string) *sets.OrderedSet[Package] {
	x := sets.NewOrderedSet[Package]()
	for _, pkg := range pkgNames {
		x.Add(Package(pkg))
	}
	return x
}

type Package string

func (p Package) Key() string { return string(p) }

func NewPackagePatternSet(patterns ...PackagePattern) *PackagePatternSet {
	ps := &PackagePatternSet{}
	ps.set = sets.NewOrderedSet(patterns...)
	return ps
}

type PackagePatternSet struct {
	set             *sets.OrderedSet[PackagePattern]
	compiledPattern *regexp.Regexp
	compileErr      error
	compiled        bool
}

func (ps *PackagePatternSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.set)
}

func (ps *PackagePatternSet) UnmarshalJSON(b []byte) error {
	var x sets.OrderedSet[PackagePattern]
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	*ps = PackagePatternSet{set: &x}
	return nil
}

func (ps *PackagePatternSet) compileOnce() {
	if ps.compiled {
		return
	}
	defer func() {
		ps.compiled = true
	}()
	patterns := ps.set.Items()
	exprs := make([]string, len(patterns))
	for i, pattern := range patterns {
		exprs[i] = string(pattern)
	}
	var assembled string
	assembled, ps.compileErr = rassemble.Join(exprs)
	if ps.compileErr != nil {
		return
	}
	ps.compiledPattern, ps.compileErr = regexp.Compile(assembled)
	if ps.compileErr != nil {
		return
	}
}

func (ps *PackagePatternSet) match(pkg Package) bool {
	ps.compileOnce()
	if ps.compiledPattern == nil {
		return false
	}
	return ps.compiledPattern.MatchString(string(pkg))
}

type PackagePattern string

func (p PackagePattern) Key() string {
	return string(p)
}

func containPackage(pkgNames *sets.OrderedSet[Package], pkgPatterns *PackagePatternSet, pkg Package) bool {
	if pkgNames != nil {
		if pkgNames.Has(pkg) {
			return true
		}
	}
	if pkgPatterns != nil {
		if pkgPatterns.match(pkg) {
			return true
		}
	}
	return false
}

// Layer is a named set of packages.
type Layer struct {
	Name                string
	PackageNames        *sets.OrderedSet[Package]
	PackageNamePatterns *PackagePatternSet
}

func (l *Layer) Key() string {
	return l.Name
}

func (l *Layer) GoString() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "Layer( %q %#v )", l.Name, l.PackageNames)
	return b.String()
}

// Rule is a pair of Layer and allowed/denied layers list.
type Rule struct {
	// Layer is a layer name applies the rule.
	Layer string

	// Allowed is layer names list that can be appeared in dependency list.
	Allowed []string

	// Denied is layer names list that can NOT be appeared in dependency list.
	Denied []string
}

func (r *Rule) determinate(layers *sets.OrderedSet[*Layer]) Decision {
	x := DecisionAllow
	for _, allowedLayer := range r.Allowed {
		if layers.HasKey(allowedLayer) || allowedLayer == "*" {
			x = x.And(DecisionAllow)
		}
	}
	for _, deniedLayer := range r.Denied {
		if layers.HasKey(deniedLayer) || deniedLayer == "*" {
			x = x.And(DecisionDeny)
		}
	}
	return x
}

func NewLayersSet(layers ...*Layer) *sets.OrderedSet[*Layer] {
	return sets.NewOrderedSet(layers...)
}

// LayersSet is an ordered set of layers.
type LayersSet sets.OrderedSet[*Layer]

func (s *LayersSet) toSet() *sets.OrderedSet[*Layer] {
	a := *s
	b := sets.OrderedSet[*Layer](a)
	return &b
}

func (s *LayersSet) findByPackagePath(pkgPath string) *Layer {
	for _, layer := range s.toSet().Items() {
		if containPackage(layer.PackageNames, layer.PackageNamePatterns, Package(pkgPath)) {
			return layer
		}
	}
	return nil
}

type Config struct {
	Layers *sets.OrderedSet[*Layer]
	Rules  []*Rule
}

func layersForPackages(layers *sets.OrderedSet[*Layer], pkg Package) *sets.OrderedSet[*Layer] {
	x := sets.NewOrderedSet[*Layer]()
	for _, layer := range layers.Items() {
		layer := layer
		if containPackage(layer.PackageNames, layer.PackageNamePatterns, pkg) {
			x.Add(layer)
		}
	}
	return x
}

func (c *Config) CanDepend(dependantLayerName string, dependency Package) Decision {
	layers := layersForPackages(c.Layers, dependency)
	for _, rule := range c.Rules {
		if rule.Layer != dependantLayerName {
			continue
		}
		return rule.determinate(layers)
	}
	return DecisionDeny
}

type Decision int

func (e Decision) String() string {
	switch e {
	case DecisionAllow:
		return "EffectAllow"
	case DecisionDeny:
		return "EffectDeny"
	default:
		panic("bug")
	}
}

func (d Decision) And(other Decision) Decision {
	if d == DecisionAllow && other == DecisionAllow {
		return DecisionAllow
	}
	return DecisionDeny
}

const (
	DecisionDeny Decision = iota
	DecisionAllow
)
