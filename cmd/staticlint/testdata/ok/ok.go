package main

import "os"

func main() {
	os.Getenv("")
	ok()
}

func ok() {
	os.Exit(1)
}
