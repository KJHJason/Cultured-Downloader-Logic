package cdlogic

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Catch SIGINT/SIGTERM signal and cancel the context when received
func catchInterruptSignal(cancel context.CancelFunc) (stopSignal func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	return func() {
		signal.Stop(sigs)
	}
}
