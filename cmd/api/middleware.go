package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// This is replaces with our ip based rate limiter
// func (app *application) rateLimt(next http.Handler) http.Handler {
// 	// initialise new rate limiter that allows average 2 requests per second
// 	// with max burst of 4
// 	limiter := rate.NewLimiter(2, 4)
//
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !limiter.Allow() {
// 			app.rateLimitExceededResponse(w, r)
// 			return
// 		}
//
// 		next.ServeHTTP(w, r)
// 	})
// }

func (app *application) rateLimit(next http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Lock the mutex to prevent this code from being executed
		// concurrently
		mu.Lock()

		if _, found := clients[ip]; !found {
			clients[ip] = rate.NewLimiter(2, 4)
		}

		// Call the Allow() method on the rate limiter for the current IP.
		// If its not allowed, unlock the mutex and send 429
		if !clients[ip].Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		// Very important to unluck the mutex before calling next
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
