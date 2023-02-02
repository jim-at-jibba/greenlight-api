package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {

		// Create quit channel which carries os.Signal values
		// Using buffered channel as signal.Notify does not wait
		// and so we need to be ready
		quit := make(chan os.Signal, 1)

		// use signal.Notify to listen for incoming SIGINT and SIGTERM signals
		// and relay them to the channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will
		// block until the signal is recieved
		s := <-quit

		app.logger.PrintInfo("shutting down server", map[string]string{"signal": s.String()})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Call Shutdown() on out server, passing in the context we just made
		// Shutdown() will return nil if the graceful hsutdown was a succcess or an error
		// due to issues closing hte listneers or not within 20 seconds graceful
		// This is relayed with the shutdownError
		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Calling Shutdown() on our server will cause ListenAndServe() to immediately return
	// a http.ErrServerClosed error. This is good as it indicates that the graceful shutdown
	// has started.

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise we wait to receive the return value from the shutdown() on the shutdownError channel
	// If the return value is an error, we know that there was a problem with the graceful shutdown
	// and return the error
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
