package onion

import (
	"encoding/json"
	"fmt"
	"strings"
)

func NewPackagesSet(pkgNames ...string) *PackagesSet {
	x := &PackagesSet{set: map[string]bool{}}
	for _, pkg := range pkgNames {
		x.add(pkg)
	}
	return x
}

// PackagesSet is an ordered set of packages.
type PackagesSet struct {
	xs  []string
	set map[string]bool
}

var _ interface {
	json.Marshaler
	json.Unmarshaler
} = &PackagesSet{}

func (ps PackagesSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.items())
}

func (ps *PackagesSet) UnmarshalJSON(b []byte) error {
	var xs []string
	if err := json.Unmarshal(b, &xs); err != nil {
		return err
	}
	*ps = PackagesSet{set: map[string]bool{}}
	for _, x := range xs {
		ps.add(x)
	}
	return nil
}

func (s *PackagesSet) items() []string {
	return s.xs
}

func (s *PackagesSet) add(pkg string) {
	s.xs = append(s.xs, pkg)
	s.set[pkg] = true
}

func (s *PackagesSet) GoString() string {
	b := new(strings.Builder)
	b.WriteString("PackagesSet(")
	for _, pkg := range s.items() {
		fmt.Fprintf(b, " %q", pkg)
	}
	b.WriteString(" )")
	return b.String()
}

func (s *PackagesSet) contains(pkgName string) bool {
	return s.set[pkgName]
}

// Layer is a named set of packages.
type Layer struct {
	Name     string
	Packages *PackagesSet
}

func (l *Layer) GoString() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "Layer( %q %#v )", l.Name, l.Packages)
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

func (r *Rule) determinate(layers *LayersSet) Decision {
	x := DecisionAllow
	for _, layer := range layers.items() {
		x = x.And(r.deter(layer))
	}
	return x
}

func NewLayersSet(layers ...*Layer) *LayersSet {
	x := LayersSet{set: map[string]int{}}
	for _, layer := range layers {
		x.add(layer)
	}
	return &x
}

// LayersSet is an ordered set of layers.
type LayersSet struct {
	xs  []*Layer
	set map[string]int
}

var _ interface {
	json.Marshaler
	json.Unmarshaler
} = &LayersSet{}

func (l LayersSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.items())
}

func (l *LayersSet) UnmarshalJSON(b []byte) error {
	var xs []*Layer
	if err := json.Unmarshal(b, &xs); err != nil {
		return err
	}
	*l = LayersSet{set: map[string]int{}}
	for _, x := range xs {
		x := x
		l.add(x)
	}
	return nil
}

func (ls *LayersSet) items() []*Layer {
	return ls.xs
}

func (ls *LayersSet) GoString() string {
	b := new(strings.Builder)
	b.WriteString("LayersSet(")
	for _, layer := range ls.items() {
		fmt.Fprintf(b, " %#v", layer)
	}
	b.WriteString(" )")
	return b.String()
}

func (s *LayersSet) add(layer *Layer) {
	s.xs = append(s.xs, layer)
	s.set[layer.Name] = len(s.xs) - 1
}

func (s *LayersSet) findByPackagePath(pkgPath string) *Layer {
	for _, layer := range s.items() {
		if layer.Packages.contains(pkgPath) {
			return layer
		}
	}
	return nil
}

type Config struct {
	Layers *LayersSet
	Rules  []*Rule
}

func layersForPackages(layers *LayersSet, pkgs *PackagesSet) *LayersSet {
	x := LayersSet{set: map[string]int{}}
	for _, pkgName := range pkgs.items() {
		for _, layer := range layers.items() {
			layer := layer
			if layer.Packages.contains(pkgName) {
				x.add(layer)
			}
		}
	}
	return &x
}

func (c *Config) CanDepend(dependantLayerName string, dependencyPkgs []string) Decision {
	depPkgs := NewPackagesSet(dependencyPkgs...)
	layers := layersForPackages(c.Layers, depPkgs)
	fmt.Printf("layers: %#v\n", layers)
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
