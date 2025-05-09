// Package main provides the entry point for the terraform-registry-builder tool.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ikedam/terraform-registry-builder/builder"
)

func main() {
	// Parse command line arguments
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s SRC DST\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  SRC: Directory containing provider binaries or packages\n")
		fmt.Fprintf(os.Stderr, "  DST: Directory for the Terraform registry namespace\n")
		os.Exit(1)
	}

	srcDir := args[0]
	dstDir := args[1]

	// Create and run the builder
	b := builder.New(srcDir, dstDir)
	err := b.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Build completed successfully.")
}
