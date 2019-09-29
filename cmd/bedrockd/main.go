package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/icelolly/go-errors"
	"github.com/seeruk/bedrockd/internal"
	"github.com/seeruk/bedrockd/internal/daemon"
)

func main() {
	resolver := internal.NewResolver()
	logger := resolver.ResolveLogger()

	ctx, cfn := context.WithCancel(context.Background())
	defer cfn()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("starting background threads...")

	process := resolver.ResolveBedrockProcess()
	processDone := daemon.NewBackgroundThread(ctx, process)

	select {
	case sig := <-signals:
		logger.Infow("caught signal, stopping threads", "signal", sig)
	case res := <-processDone:
		logger.Fatalw("error starting process thread", "error", res.Error)
	}

	cfn()

	go func() {
		time.AfterFunc(10*time.Second, func() {
			logger.Error("took too long stopping, sending kill signals and exiting")

			if err := process.Kill(); err != nil {
				logger.Errorw("failed to kill process thread", "stack", errors.Stack(err))
			}

			os.Exit(1)
		})
	}()

	<-processDone

	logger.Info("threads stopped successfully, exiting")
}
