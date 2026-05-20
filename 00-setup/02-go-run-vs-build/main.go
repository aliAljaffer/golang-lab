package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("compiled for %s/%s using %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
}
