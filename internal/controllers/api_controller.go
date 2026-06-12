package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"cwmp-acs/db"
	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/session"
)

var ApiControllerPath = "/api"

// ApiRequest is the JSON body accepted by the API controller.
type ApiRequest struct {
	DeviceID       string          `json:"device_id"`
	RPC            string          `json:"rpc"`
	Parameters     json.RawMessage `json:"parameters,omitempty"`
	TriggerSession bool            `json:"trigger_session,omitempty"`
}

// Parameter structs for each supported outbound CWMP message type.

type GetParameterValuesParams struct {
	Names []string `json:"names"`
}

type GetParameterNamesParams struct {
	ParameterPath string `json:"parameter_path"`
	NextLevel     bool   `json:"next_level"`
}

type RebootParams struct {
	CommandKey string `json:"command_key"`
}

type SetParameterAttributeEntry struct {
	Name               string   `json:"name"`
	NotificationChange bool     `json:"notification_change"`
	Notification       int      `json:"notification"`
	AccessListChange   bool     `json:"access_list_change"`
	AccessList         []string `json:"access_list"`
}

type SetParameterAttributesParams struct {
	Attributes []SetParameterAttributeEntry `json:"attributes"`
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	var req ApiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}
	if req.RPC == "" {
		http.Error(w, "RPC name is required", http.StatusBadRequest)
		return
	}

	msg, err := buildCwmpMessage(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to build message: %v", err), http.StatusBadRequest)
		return
	}

	if err := db.QueueRPC(r.Context(), req.DeviceID, msg.GetID(), req.RPC, req.Parameters); err != nil {
		http.Error(w, "failed to enqueue message", http.StatusInternalServerError)
		return
	}

	if req.TriggerSession {
		sessionInfo, ok := session.DeviceIdStringActiveSessions[req.DeviceID]
		if !ok || sessionInfo == nil || sessionInfo.SessionState == session.TERMINATED {
			fmt.Printf("No active session for %s, so triggering a connection request\n", req.DeviceID)

			if err := sendConnectionRequest(r.Context(), req.DeviceID); err != nil {
				http.Error(w, fmt.Sprintf("failed to trigger connection request: %v", err), http.StatusInternalServerError)
				return
			}
		}
	} else {
		fmt.Printf("RPC queued and will wait for next session from device %s\n", req.DeviceID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "queued", "message_id": msg.GetID()})
}

func newMessageID() string {
	return uuid.NewString()
}

func buildCwmpMessage(req ApiRequest) (cwmp.CwmpMessageInterface, error) {
	return buildCwmpMessageFromParts(newMessageID(), req.RPC, req.Parameters)
}

func buildCwmpMessageFromParts(id, name string, params json.RawMessage) (cwmp.CwmpMessageInterface, error) {
	switch name {
	case "GetParameterValues":
		var p GetParameterValuesParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid parameters for GetParameterValues: %w", err)
		}
		if len(p.Names) == 0 {
			return nil, fmt.Errorf("GetParameterValues requires at least one name")
		}
		return &cwmp.GetParameterValues{
			CwmpMessage:    cwmp.CwmpMessage{Name: "GetParameterValues", CwmpHeader: cwmp.CwmpHeader{ID: id}},
			ParameterNames: p.Names,
		}, nil

	case "GetParameterNames":
		var p GetParameterNamesParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid parameters for GetParameterNames: %w", err)
		}
		return &cwmp.GetParameterNames{
			CwmpMessage:   cwmp.CwmpMessage{Name: "GetParameterNames", CwmpHeader: cwmp.CwmpHeader{ID: id}},
			ParameterPath: p.ParameterPath,
			NextLevel:     p.NextLevel,
		}, nil

	case "Reboot":
		var p RebootParams
		if len(params) > 0 && string(params) != "null" {
			if err := json.Unmarshal(params, &p); err != nil {
				return nil, fmt.Errorf("invalid parameters for Reboot: %w", err)
			}
		}
		return &cwmp.Reboot{
			CwmpMessage: cwmp.CwmpMessage{Name: "Reboot", CwmpHeader: cwmp.CwmpHeader{ID: id}},
			CommandKey:  p.CommandKey,
		}, nil

	case "GetRPCMethods":
		return &cwmp.GetRPCMethods{
			CwmpMessage: cwmp.CwmpMessage{Name: "GetRPCMethods", CwmpHeader: cwmp.CwmpHeader{ID: id}},
		}, nil

	case "SetParameterAttributes":
		var p SetParameterAttributesParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid parameters for SetParameterAttributes: %w", err)
		}
		if len(p.Attributes) == 0 {
			return nil, fmt.Errorf("SetParameterAttributes requires at least one attribute entry")
		}
		attrs := make([]cwmp.SetParameterAttributeStruct, len(p.Attributes))
		for i, a := range p.Attributes {
			attrs[i] = cwmp.SetParameterAttributeStruct{
				Name:               a.Name,
				NotificationChange: a.NotificationChange,
				Notification:       a.Notification,
				AccessListChange:   a.AccessListChange,
				AccessList:         a.AccessList,
			}
		}
		return &cwmp.SetParameterAttributes{
			CwmpMessage:   cwmp.CwmpMessage{Name: "SetParameterAttributes", CwmpHeader: cwmp.CwmpHeader{ID: id}},
			ParameterList: attrs,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported message type: %q", name)
	}
}

func sendConnectionRequest(ctx context.Context, deviceID string) error {
	deviceInfo, err := db.GetDeviceByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to fetch device info: %w", err)
	}
	if deviceInfo == nil {
		return fmt.Errorf("device not found")
	}

	if deviceInfo.ConnectionRequestURL == "" {
		return fmt.Errorf("device does not have a connection request URL")
	}

	fmt.Printf("Sending connection request to %s\n", deviceInfo.ConnectionRequestURL)
	return sendHttpConnectionRequest(deviceInfo.ConnectionRequestURL)
}

func sendHttpConnectionRequest(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create connection request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("connection request returned non-2xx status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read connection request response: %w", err)
	}
	fmt.Printf("connection request response (%d): %s\n", resp.StatusCode, body)

	return nil
}
