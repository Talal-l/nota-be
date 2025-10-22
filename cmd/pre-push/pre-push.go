package main

import (
	"context"
	"os"

	"github.com/anotik/anocore/pkg/hook"
	"github.com/anotik/anocore/pkg/logger"
)

func main() {
	ctx := context.Background()
	log, err := logger.NewContextLogger(ctx)
	if err != nil {
		log.Error("failed to create logger", "error", err)
	}
	err = hook.Vet("./...")
	if err != nil {
		log.Error("failed to vet", "error", err)
		os.Exit(1)
	}

	err = hook.RunTests("./...")
	if err != nil {
		log.Error("failed to run tests", "error", err)
		os.Exit(1)
	}
	err = hook.CheckCoreDep()
	if err != nil {
		log.Error("failed to check core dep", "error", err)
		os.Exit(1)
	}
	os.Exit(0)
}
