package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/adrianord/terrapack/pkg/version"
)

func panicHandler() {
	if panicPayload := recover(); panicPayload != nil {
		stack := string(debug.Stack())
		fmt.Fprintln(os.Stderr, "===========================================================================")
		fmt.Fprintln(os.Stderr, "Terrapack has paniced! This is a bug in Terrapack.")
		fmt.Fprintln(os.Stderr, "Please open an issue at: https://github.com/adrianord/terrapack/issues")
		fmt.Fprintln(os.Stderr, "Provide all the information below.")
		fmt.Fprintln(os.Stderr, "===========================================================================")
		fmt.Fprintf(os.Stderr, "Terrapack version: %s\n", version.Version)
		fmt.Fprintf(os.Stderr, "Go Version:        %s\n", runtime.Version())
		fmt.Fprintf(os.Stderr, "Go Compiler:       %s\n", runtime.Compiler)
		fmt.Fprintf(os.Stderr, "Architecture:      %s\n", runtime.GOARCH)
		fmt.Fprintf(os.Stderr, "Operating System:  %s\n", runtime.GOOS)
		fmt.Fprintf(os.Stderr, "Panic:             %s\n\n", panicPayload)
		fmt.Fprintln(os.Stderr, stack)
		os.Exit(1)
	}
}

func main() {
	defer panicHandler()
	if err := newTerrapackCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
		os.Exit(1)
	}
}
