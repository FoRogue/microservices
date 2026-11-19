package eventing

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func Serve(server *http.Server, workers ...func(context.Context)) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var running sync.WaitGroup
	for _, work := range workers {
		running.Go(func() { work(ctx) })
	}
	serverResult := make(chan error, 1)
	go func() { serverResult <- server.ListenAndServe() }()

	var err error
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = server.Shutdown(shutdownCtx)
		if err != nil {
			server.Close()
		}
	case err = <-serverResult:
		stop()
	}
	running.Wait()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
