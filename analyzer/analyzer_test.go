package analyzer_test

import (
	"path/filepath"
	"testing"

	"github.com/aereal/pkgboundaries/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testCases := []struct {
		configPath string
		patterns   []string
	}{
		{"../testdata/config.json", []string{"github.com/aereal/a"}},
		{"../testdata/b.json", []string{"github.com/aereal/b"}},
		{"../testdata/empty.json", []string{"github.com/aereal/empty"}},
		{"../testdata/allow-all.json", []string{"github.com/aereal/c"}},
		{"../testdata/deny-all.json", []string{"github.com/aereal/d"}},
	}
	for _, tc := range testCases {
		t.Run(filepath.Base(tc.configPath), func(t *testing.T) {
			testdata := analysistest.TestData()
			clean := analyzer.SetConfigPathForTesting(tc.configPath)
			defer clean()
			analysistest.Run(t, testdata, analyzer.Analyzer, tc.patterns...)
		})
	}
}
