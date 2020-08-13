package config

import (
	"context"
	"os"
	"os/signal"

	"github.com/olympus-protocol/ogen/pkg/logger"
)

var shutdownRequestChannel = make(chan struct{})

var interruptSignals = []os.Signal{os.Interrupt}

func InterruptListener(log *logger.Logger, cancel context.CancelFunc) {
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, interruptSignals...)
		select {
		case sig := <-interruptChannel:
			log.Warnf("Received signal (%s).  Shutting down...",
				sig)
		case <-shutdownRequestChannel:
			log.Warn("Shutdown requested.  Shutting down...")
		}
		cancel()
		for {
			select {
			case sig := <-interruptChannel:
				log.Warnf("Received signal (%s).  Already "+
					"shutting down...", sig)

			case <-shutdownRequestChannel:
				log.Warn("Shutdown requested.  Already " +
					"shutting down...")
			}
		}
	}()
}
