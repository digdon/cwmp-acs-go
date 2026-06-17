package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cwmp-acs/internal/events"
)

var EventsControllerPath = "/events"

// EventsHandler streams ACS events to clients using Server-Sent Events (SSE).
// Clients connect with GET /events and receive a stream of JSON-encoded events.
// An optional "device_id" query parameter filters events to a specific device.
func EventsHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	deviceFilter := r.URL.Query().Get("device_id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable buffering in nginx proxies

	var filter func(events.Event) bool

	if deviceFilter != "" {
		fmt.Printf("client subscribed to events for device_id=%s\n", deviceFilter)
		filter = func(e events.Event) bool {
			return e.DeviceID == deviceFilter
		}
	} else {
		fmt.Println("client subscribed to all events")
	}
	id, ch := events.GlobalBroker.SubscribeWithFilter(filter)
	defer events.GlobalBroker.Unsubscribe(id)

	for {
		select {
		case <-r.Context().Done():
			fmt.Println("client disconnected from events stream")
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(evt)
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				flusher.Flush()
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
			flusher.Flush()
		}
	}
}
