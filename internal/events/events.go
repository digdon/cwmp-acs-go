package events

import "time"

type EventType string

const (
	EventInform                         EventType = "Inform"
	EventGetParameterNamesResponse      EventType = "GetParameterNamesResponse"
	EventGetParameterValuesResponse     EventType = "GetParameterValuesResponse"
	EventGetParameterAttributesResponse EventType = "GetParameterAttributesResponse"
	EventSetParameterValuesResponse     EventType = "SetParameterValuesResponse"
	EventSetParameterAttributesResponse EventType = "SetParameterAttributesResponse"
	EventAddObjectResponse              EventType = "AddObjectResponse"
	EventDeleteObjectResponse           EventType = "DeleteObjectResponse"
	EventRebootResponse                 EventType = "RebootResponse"
	EventFault                          EventType = "Fault"
)

type Event struct {
	Type      EventType `json:"type"`
	DeviceID  string    `json:"device_id"`
	SessionID string    `json:"session_id"`
	MessageID string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}
