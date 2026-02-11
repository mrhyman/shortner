package main

import (
	"github.com/mrhyman/shortner/cmd/linter/linter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
