package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/algorand/msgpackdiff/msgpackdiff"
)

var brief = flag.Bool("brief", false, "Disable comparison report")
var ignoreEmpty = flag.Bool("ignore-empty", false, "Treat missing fields as empty objects for comparison")
var ignoreOrder = flag.Bool("ignore-order", false, "Ignore ordering of fields for comparison")
var flexibleTypes = flag.Bool("flexible-types", false, "Compare all numerical values regardless of their type. May be inaccurate")

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Must specify exactly two objects to compare")
		os.Exit(2)
	}

	binA, err := msgpackdiff.GetBinary(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to extract first object: %v\n", err)
		os.Exit(2)
	}

	binB, err := msgpackdiff.GetBinary(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to extract second object: %v\n", err)
		os.Exit(2)
	}

	options := msgpackdiff.CompareOptions{
		Brief:         *brief,
		IgnoreEmpty:   *ignoreEmpty,
		IgnoreOrder:   *ignoreOrder,
		FlexibleTypes: *flexibleTypes,
	}

	result, err := msgpackdiff.Compare(binA, binB, options)

	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
		os.Exit(2)
	}

	result.PrintReport(os.Stdout)

	if !result.Equal {
		fmt.Println("Objects are not equal")
		os.Exit(1)
	}

	fmt.Println("Objects are equal")
}
