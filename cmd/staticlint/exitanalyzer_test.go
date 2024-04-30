package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ExitCheckAnalyzer, "./...")
}
