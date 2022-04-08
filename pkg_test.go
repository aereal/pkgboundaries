package pkgboundaries_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aereal/pkgboundaries"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestConfig_CanDepend(t *testing.T) {
	cfg := &pkgboundaries.Config{
		Layers: pkgboundaries.NewLayersSet(
			&pkgboundaries.Layer{Name: "a", PackageNames: pkgboundaries.NewPackagesSet("pkg/1", "pkg/2")},
			&pkgboundaries.Layer{Name: "b", PackageNames: pkgboundaries.NewPackagesSet("pkg/3", "pkg/4")},
			&pkgboundaries.Layer{Name: "c", PackageNames: pkgboundaries.NewPackagesSet("pkg/5", "pkg/6")},
		),
		Rules: []*pkgboundaries.Rule{
			{Layer: "a", Allowed: []string{"b"}, Denied: []string{"c"}},
		},
	}
	type args struct {
		dependantLayerName string
		dependency         pkgboundaries.Package
	}
	testCases := []struct {
		name       string
		args       args
		wantEffect pkgboundaries.Decision
	}{
		{"ok", args{"a", "pkg/3"}, pkgboundaries.DecisionAllow},
		{"ng", args{"a", "pkg/5"}, pkgboundaries.DecisionDeny},
		{"ng (unknown)", args{"a", "pkg/x"}, pkgboundaries.DecisionAllow},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := cfg.CanDepend(tc.args.dependantLayerName, tc.args.dependency)
			if got != tc.wantEffect {
				t.Errorf("want=%s got=%s", tc.wantEffect, got)
			}
		})
	}
}

func TestEffect_And(t *testing.T) {
	testCases := []struct {
		x    pkgboundaries.Decision
		y    pkgboundaries.Decision
		want pkgboundaries.Decision
	}{
		{pkgboundaries.DecisionAllow, pkgboundaries.DecisionAllow, pkgboundaries.DecisionAllow},
		{pkgboundaries.DecisionAllow, pkgboundaries.DecisionDeny, pkgboundaries.DecisionDeny},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("x=%s y=%s", tc.x, tc.y), func(t *testing.T) {
			got := tc.x.And(tc.y)
			if got != tc.want {
				t.Errorf("got=%s want=%s", got, tc.want)
			}
		})
	}
}

func TestConfig_Marshaling(t *testing.T) {
	f, err := os.Open("./testdata/config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var fromData pkgboundaries.Config
	if err := json.NewDecoder(f).Decode(&fromData); err != nil {
		t.Fatal(err)
	}
	wantBytes, err := json.MarshalIndent(fromData, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	want := &pkgboundaries.Config{
		Rules: []*pkgboundaries.Rule{
			{
				Layer:   "App",
				Allowed: []string{"Errors"},
				Denied:  []string{"Print", "Encoding"},
			},
		},
		Layers: pkgboundaries.NewLayersSet(
			&pkgboundaries.Layer{
				Name:         "App",
				PackageNames: pkgboundaries.NewPackagesSet("github.com/aereal/a"),
			},
			&pkgboundaries.Layer{
				Name:         "Errors",
				PackageNames: pkgboundaries.NewPackagesSet("errors"),
			},
			&pkgboundaries.Layer{
				Name:         "Print",
				PackageNames: pkgboundaries.NewPackagesSet("fmt", "log"),
			},
			&pkgboundaries.Layer{
				Name:                "Encoding",
				PackageNamePatterns: pkgboundaries.NewPackagePatternSet(pkgboundaries.PackagePattern("^encoding/")),
			},
		),
	}
	marshaled, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(wantBytes), string(marshaled)); diff != "" {
		t.Errorf("json.Marshal (-want, +got):\n%s", diff)
	}

	var got pkgboundaries.Config
	if err := json.Unmarshal(marshaled, &got); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, &got, cmpopts.IgnoreUnexported(pkgboundaries.OrderedSet[pkgboundaries.Package]{}, pkgboundaries.OrderedSet[*pkgboundaries.Layer]{})); diff != "" {
		t.Errorf("json.Unmarshal (-want, +got):\n%s", diff)
	}
}
