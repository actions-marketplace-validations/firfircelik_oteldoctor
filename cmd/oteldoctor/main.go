package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/firfircelik/oteldoctor/internal/cli"
)

func main() {
	err := cli.Execute()
	if err == nil {
		os.Exit(0)
	}

	var exitErr *cli.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.Code)
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(2)
}
