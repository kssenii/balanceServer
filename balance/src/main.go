package main

import (
	"context"
	"handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	l := log.New(os.Stdout, "balance-info ", log.LstdFlags)
	startupHandler := handlers.Startup(l)
	shutdownHandler := handlers.Shutdown(l)
	clientHandler := handlers.NewClient(l)

	serveMux := http.NewServeMux()
	serveMux.Handle("/", startupHandler)
	serveMux.Handle("/shutdown", shutdownHandler)
	serveMux.Handle("/client", clientHandler)

	server := &http.Server{
		Addr:         ":9090",
		Handler:      serveMux,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		l.Println("Starting server on port 9090")

		err := server.ListenAndServe()
		if err != nil {
			l.Printf("Error starting server: %s", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	sig := <-sigChan // Block untill message is available to be consumed
	l.Println(" Received terminate, gracefull shutdown", sig)

	timeoutContext, _ := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(timeoutContext)

}
