package main

import (
	"os"

	"github.com/prabath/nexperf/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
