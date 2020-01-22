package config

import (
	"github.com/grupokindynos/ogen/logger"
	"os"
	"os/signal"
)

var shutdownRequestChannel = make(chan struct{})

var interruptSignals = []os.Signal{os.Interrupt}

func InterruptListener(log *logger.Logger) <-chan struct{} {
	c := make(chan struct{})
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
		close(c)
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
	return c
}
