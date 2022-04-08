package onion_test

import (
	"testing"

	"github.com/aereal/onion"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	clean := onion.SetConfigPathForTesting("./testdata/config.json")
	defer clean()
	analysistest.Run(t, testdata, onion.Analyzer, "github.com/aereal/a", "github.com/aereal/b")
}

func Test_withEmptyConfig(t *testing.T) {
	testdata := analysistest.TestData()
	clean := onion.SetConfigPathForTesting("./testdata/empty.json")
	defer clean()
	analysistest.Run(t, testdata, onion.Analyzer, "github.com/aereal/empty")
}
