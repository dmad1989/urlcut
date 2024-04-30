// Package staticlint implements set of static checks.
//
// # Set of statichecks are described in config.json
//
// Current checklist are included:
//
// 1. All checks from golang.org/x/tools/go/analysis/passes
//
// 2. All SA checks from https://staticcheck.io/docs/checks/
//
// 3. All ST checks from https://staticcheck.io/docs/checks/
//
// 4. Check wrapping errors https://github.com/fatih/errwrap
//
// 5. Check for calling os.Exit in main func of main package
//
// Example:
//
//	staticlint -SA1000 <project path>
//
// Perform SA1000 analysis for given project.
// For more details run:
//
//	staticlint -help
//
// exitanalyzer investigates main package for calling os.Exit from main function. Run this check with following command:
//
//	staticlint -exitanalyzer
package main
