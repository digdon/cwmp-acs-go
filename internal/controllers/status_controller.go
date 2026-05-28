package controllers

import (
	"fmt"
	"net/http"
)

var StatusControllerPath = "/status"

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "CWMP ACS is running")
}
