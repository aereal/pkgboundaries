package onion_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aereal/onion"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestConfig_CanDepend(t *testing.T) {
	cfg := &onion.Config{
		Layers: onion.NewLayersSet(
			&onion.Layer{Name: "a", Packages: onion.NewPackagesSet("pkg/1", "pkg/2")},
			&onion.Layer{Name: "b", Packages: onion.NewPackagesSet("pkg/3", "pkg/4")},
			&onion.Layer{Name: "c", Packages: onion.NewPackagesSet("pkg/5", "pkg/6")},
		),
		Rules: []*onion.Rule{
			{Layer: "a", Allowed: []string{"b"}},
		},
	}
	type args struct {
		dependantLayerName string
		dependency         string
	}
	testCases := []struct {
		name       string
		args       args
		wantEffect onion.Decision
	}{
		{"ok", args{"a", "pkg/3"}, onion.DecisionAllow},
		{"ng", args{"a", "pkg/5"}, onion.DecisionDeny},
		{"ng (unknown)", args{"a", "pkg/x"}, onion.DecisionAllow},
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
		x    onion.Decision
		y    onion.Decision
		want onion.Decision
	}{
		{onion.DecisionAllow, onion.DecisionAllow, onion.DecisionAllow},
		{onion.DecisionAllow, onion.DecisionDeny, onion.DecisionDeny},
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
	var fromData onion.Config
	if err := json.NewDecoder(f).Decode(&fromData); err != nil {
		t.Fatal(err)
	}
	wantBytes, err := json.MarshalIndent(fromData, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	want := &onion.Config{
		Rules: []*onion.Rule{
			{
				Layer:   "App",
				Allowed: []string{"Errors"},
				Denied:  []string{"Print"},
			},
		},
		Layers: onion.NewLayersSet(
			&onion.Layer{
				Name:     "App",
				Packages: onion.NewPackagesSet("github.com/aereal/a"),
			},
			&onion.Layer{
				Name:     "Errors",
				Packages: onion.NewPackagesSet("errors"),
			},
			&onion.Layer{
				Name:     "Print",
				Packages: onion.NewPackagesSet("fmt", "log"),
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

	var got onion.Config
	if err := json.Unmarshal(marshaled, &got); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, &got, cmpopts.IgnoreUnexported(onion.PackagesSet{}, onion.LayersSet{})); diff != "" {
		t.Errorf("json.Unmarshal (-want, +got):\n%s", diff)
	}
}
