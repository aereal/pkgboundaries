package onion_test

import (
	"testing"

	"github.com/aereal/onion"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testCases := []struct {
		configPath string
		patterns   []string
	}{
		{"./testdata/config.json", []string{"github.com/aereal/a"}},
		{"./testdata/b.json", []string{"github.com/aereal/b"}},
		{"./testdata/empty.json", []string{"github.com/aereal/empty"}},
	}
	for _, tc := range testCases {
		t.Run(tc.configPath, func(t *testing.T) {
			testdata := analysistest.TestData()
			clean := onion.SetConfigPathForTesting(tc.configPath)
			defer clean()
			analysistest.Run(t, testdata, onion.Analyzer, tc.patterns...)
		})
	}
}
