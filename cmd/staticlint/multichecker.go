package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

const configFile = "config.json"

type ConfigData struct {
	Staticcheck Staticcheck
}

type Staticcheck struct {
	groups []string
	checks []string
}

func main() {
	mychecks := []*analysis.Analyzer{
		shadow.Analyzer,
		appends.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		defers.Analyzer,
		errorsas.Analyzer,
		httpmux.Analyzer,
		lostcancel.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unusedresult.Analyzer,
		ExitCheckAnalyzer,
	}
	cfg := confidData()

analyzersLoop:
	for _, v := range staticcheck.Analyzers {
		// add groups of linters
		for _, g := range cfg.Staticcheck.groups {
			if strings.HasPrefix(v.Analyzer.Name, g) {
				mychecks = append(mychecks, v.Analyzer)
				continue analyzersLoop
			}
		}
		// add  linter by name
		for _, c := range cfg.Staticcheck.checks {
			if v.Analyzer.Name == c {
				mychecks = append(mychecks, v.Analyzer)
				continue analyzersLoop
			}
		}
	}
	multichecker.Main(mychecks...)

}

func confidData() (cfg ConfigData) {
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), configFile))
	if err != nil {
		panic(err)
	}
	// var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	return
}
