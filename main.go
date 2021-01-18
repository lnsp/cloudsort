package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/lnsp/cloudsort/cmd"
	"go.uber.org/zap"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
