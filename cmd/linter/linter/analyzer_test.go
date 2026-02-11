package linter_test

import (
	"testing"

	"github.com/mrhyman/shortner/cmd/linter/linter"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, linter.Analyzer, "a", "b")
}
