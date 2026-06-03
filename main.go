package main

import (
	"context"
	"cwmp-acs/internal/controllers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var serverPort = ":8080"

func main() {
	log.Printf("CWMP ACS starting on %s\n", serverPort)

	srv := &http.Server{
		Addr:    serverPort,
		Handler: http.HandlerFunc(router),
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Error starting server:", err)
		}
	}()

	// Block until interrupt or terminate signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give active connections time to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}

type Route struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
}

var routingTable = []Route{
	{Path: controllers.CwmpControllerPath, Method: http.MethodPost, Handler: controllers.CwmpHandler},
	{Path: controllers.StatusControllerPath, Method: http.MethodGet, Handler: controllers.StatusHandler},
}

func router(w http.ResponseWriter, r *http.Request) {
	// Search the request path routing table for a matching handler
	for _, route := range routingTable {
		if r.URL.Path == route.Path {
			// Path found, but is the method type correct?
			if r.Method != route.Method {
				// Wrong method type - return error
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			} else {
				// Right method - call the handler
				route.Handler(w, r)
				return
			}
		}
	}

	// No matching path found
	http.NotFound(w, r)
}
