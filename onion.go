package onion

import (
	"encoding/json"
	"errors"
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

type PackagesSet struct {
	xs  []string
	set map[string]bool
}

var _ interface {
	json.Marshaler
	json.Unmarshaler
} = &PackagesSet{}

func (ps PackagesSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.Items())
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

func (s *PackagesSet) Items() []string {
	return s.xs
}

func (s *PackagesSet) add(pkg string) {
	s.xs = append(s.xs, pkg)
	s.set[pkg] = true
}

func (s *PackagesSet) GoString() string {
	b := new(strings.Builder)
	b.WriteString("PackagesSet(")
	for _, pkg := range s.Items() {
		fmt.Fprintf(b, " %q", pkg)
	}
	b.WriteString(" )")
	return b.String()
}

func (s *PackagesSet) Contains(pkgName string) bool {
	return s.set[pkgName]
}

type Layer struct {
	Name     string
	Packages *PackagesSet
}

func (l *Layer) GoString() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "Layer( %q %#v )", l.Name, l.Packages)
	return b.String()
}

type Rule struct {
	DependantLayer    string
	AllowedLayerNames []string
	DeniedLayerNames  []string
}

func (r *Rule) deter(layer *Layer) Decision {
	for _, layerName := range r.AllowedLayerNames {
		if layer.Name == layerName {
			return DecisionAllow
		}
	}
	for _, layerName := range r.DeniedLayerNames {
		if layer.Name == layerName {
			return DecisionDeny
		}
	}
	return DecisionDeny
}

func (r *Rule) determinate(layers *LayersSet) Decision {
	x := DecisionAllow
	for _, layer := range layers.Items() {
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

type LayersSet struct {
	xs  []*Layer
	set map[string]int
}

var _ interface {
	json.Marshaler
	json.Unmarshaler
} = &LayersSet{}

func (l LayersSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Items())
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

func (ls *LayersSet) Items() []*Layer {
	return ls.xs
}

func (ls *LayersSet) GoString() string {
	b := new(strings.Builder)
	b.WriteString("LayersSet(")
	for _, layer := range ls.Items() {
		fmt.Fprintf(b, " %#v", layer)
	}
	b.WriteString(" )")
	return b.String()
}

func (s *LayersSet) add(layer *Layer) {
	s.xs = append(s.xs, layer)
	s.set[layer.Name] = len(s.xs) - 1
}

func (s *LayersSet) Find(name string) *Layer {
	i, ok := s.set[name]
	if !ok {
		return nil
	}
	return s.xs[i]
}

func (s *LayersSet) FromPackagePath(pkgPath string) *Layer {
	for _, layer := range s.Items() {
		if layer.Packages.Contains(pkgPath) {
			return layer
		}
	}
	return nil
}

type Config struct {
	Layers *LayersSet
	Rules  []*Rule
}

var ErrLayerNotFound = errors.New("layer not found")

func (c *Config) FindLayer(layerName string) *Layer {
	return c.Layers.Find(layerName)
}

func layersForPackages(layers *LayersSet, pkgs *PackagesSet) *LayersSet {
	x := LayersSet{set: map[string]int{}}
	for _, pkgName := range pkgs.Items() {
		for _, layer := range layers.Items() {
			layer := layer
			if layer.Packages.Contains(pkgName) {
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
		if rule.DependantLayer != dependantLayerName {
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
