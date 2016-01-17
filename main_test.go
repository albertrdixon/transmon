package main

import (
	"os"

	"github.com/albertrdixon/gearbox/logger"
)

func init() {
	logger.Configure("debug", "[transmon] ", os.Stdout)
}
