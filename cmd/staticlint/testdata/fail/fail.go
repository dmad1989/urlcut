package main

import "os"

func main() {
	os.Setenv("", "")
	os.Exit(1)                            // want "os.Exit not allowed in func main"
	defer os.Exit(1)                      // want "os.Exit not allowed in func main"
	f := func(code int) { os.Exit(code) } // want "os.Exit not allowed in func main"
	f(1)
}
