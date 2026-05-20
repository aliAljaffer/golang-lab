package main

import (
	"fmt"
	"runtime"
)

func main() {
	// runtime constants captured at compile time — these reflect the binary's target.
	fmt.Printf("GOOS=%s GOARCH=%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Go version=%s\n", runtime.Version())
	fmt.Printf("NumCPU=%d (the scheduler will use up to GOMAXPROCS of these)\n", runtime.NumCPU())
	fmt.Printf("NumGoroutine=%d (just this main goroutine right now)\n", runtime.NumGoroutine())

	fmt.Println()
	fmt.Println("For paths and toolchain locations, run: `go env`")
	fmt.Println("Useful keys: GOROOT, GOPATH, GOBIN, GOMODCACHE, GOCACHE, GOPROXY")
}
