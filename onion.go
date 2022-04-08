package onion

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/itchyny/rassemble-go"
)

func NewPackagesSet(pkgNames ...string) *OrderedSet[Package] {
	x := initOrderedSet[Package]()
	for _, pkg := range pkgNames {
		x.add(Package(pkg))
	}
	return x
}

type Package string

func (p Package) Key() string { return string(p) }

func NewPackagePatternSet(patterns ...PackagePattern) *PackagePatternSet {
	ps := &PackagePatternSet{}
	ps.set = initOrderedSet[PackagePattern]()
	for _, pattern := range patterns {
		ps.set.add(pattern)
	}
	return ps
}

type PackagePatternSet struct {
	set             *OrderedSet[PackagePattern]
	compiledPattern *regexp.Regexp
	compileErr      error
	compiled        bool
}

func (ps *PackagePatternSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.set)
}

func (ps *PackagePatternSet) UnmarshalJSON(b []byte) error {
	var x OrderedSet[PackagePattern]
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
	patterns := ps.set.items()
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

func containPackage(pkgNames *OrderedSet[Package], pkgPatterns *PackagePatternSet, pkg Package) bool {
	if pkgNames != nil {
		if pkgNames.contains(pkg) {
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
	PackageNames        *OrderedSet[Package]
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

func (r *Rule) deter(layer *Layer) Decision {
	for _, layerName := range r.Allowed {
		if layer.Name == layerName {
			return DecisionAllow
		}
	}
	for _, layerName := range r.Denied {
		if layer.Name == layerName {
			return DecisionDeny
		}
	}
	return DecisionDeny
}

func (r *Rule) determinate(layers *OrderedSet[*Layer]) Decision {
	x := DecisionAllow
	for _, layer := range layers.items() {
		x = x.And(r.deter(layer))
	}
	return x
}

func NewLayersSet(layers ...*Layer) *OrderedSet[*Layer] {
	x := initOrderedSet[*Layer]()
	for _, layer := range layers {
		x.add(layer)
	}
	return x
}

// LayersSet is an ordered set of layers.
type LayersSet OrderedSet[*Layer]

func (s *LayersSet) toSet() *OrderedSet[*Layer] {
	a := *s
	b := OrderedSet[*Layer](a)
	return &b
}

func (s *LayersSet) findByPackagePath(pkgPath string) *Layer {
	for _, layer := range s.toSet().items() {
		if containPackage(layer.PackageNames, layer.PackageNamePatterns, Package(pkgPath)) {
			return layer
		}
	}
	return nil
}

type Config struct {
	Layers *OrderedSet[*Layer]
	Rules  []*Rule
}

func layersForPackages(layers *OrderedSet[*Layer], pkg Package) *OrderedSet[*Layer] {
	x := initOrderedSet[*Layer]()
	for _, layer := range layers.items() {
		layer := layer
		if containPackage(layer.PackageNames, layer.PackageNamePatterns, pkg) {
			x.add(layer)
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

func initOrderedSet[T interface{ Key() string }]() *OrderedSet[T] {
	return &OrderedSet[T]{set: map[string]int{}}
}

func NewOrderedSet[T interface{ Key() string }](xs ...T) *OrderedSet[T] {
	set := initOrderedSet[T]()
	for _, x := range xs {
		set.add(x)
	}
	return set
}

type OrderedSet[T interface{ Key() string }] struct {
	xs  []T
	set map[string]int
}

func (s *OrderedSet[T]) add(x T) {
	if _, found := s.set[x.Key()]; found {
		return
	}
	s.xs = append(s.xs, x)
	s.set[x.Key()] = len(s.xs) - 1
}

func (s *OrderedSet[T]) items() []T {
	return s.xs
}

func (s *OrderedSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.items())
}

func (s *OrderedSet[T]) UnmarshalJSON(b []byte) error {
	var vals []T
	if err := json.Unmarshal(b, &vals); err != nil {
		return err
	}
	xs := &OrderedSet[T]{set: map[string]int{}}
	for _, x := range vals {
		xs.add(x)
	}
	*s = *xs
	return nil
}

func (s *OrderedSet[T]) contains(x T) bool {
	return s.containsByKey(x.Key())
}

func (s *OrderedSet[T]) containsByKey(key string) bool {
	_, ok := s.set[key]
	return ok
}
