package main

import (
	"github.com/olympus-protocol/ogen/cli"
	"github.com/olympus-protocol/ogen/utils/logger"
	"os"
)

var log = logger.New(os.Stdin)

func main() {
	cli.Execute()
}
