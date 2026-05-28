package main

import (
	"cwmp-acs/internal/controllers"
	"log"
	"net/http"
)

var serverPort = ":8080"

func main() {
	log.Printf("CWMP ACS starting on %s\n", serverPort)

	err := http.ListenAndServe(serverPort, http.HandlerFunc(router))
	if err != nil {
		log.Println("Error starting server:", err)
	}
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
