package session

import (
	"context"

	"github.com/google/uuid"

	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/xml"
)

const SessionCookieName = "CWMP-Session-ID"

type SessionState int

const (
	NEW SessionState = iota
	INITIATING
	RECEIVING_CPE_RPCS
	SENDING_ACS_RPCS
	TERMINATED
)

type SessionInfo struct {
	SessionID               string
	DeviceIdString          string
	DeviceID                cwmp.DeviceId
	SessionState            SessionState
	CwmpVersion             cwmp.SupportedCwmpVersion
	LastIncomingMessageTime int64
	LastOutgoingMessageTime int64
	XmlNamespaces           map[xml.NamespaceID]xml.Namespace
	Context                 context.Context
	ActiveRPC               cwmp.CwmpMessageInterface
}

var SessionIdActiveSessions = map[string]*SessionInfo{}
var DeviceIdStringActiveSessions = map[string]*SessionInfo{}

func CreateNewSession() *SessionInfo {
	sessionID := uuid.NewString()

	sessionInfo := &SessionInfo{
		SessionID:               sessionID,
		DeviceID:                cwmp.DeviceId{},
		SessionState:            NEW,
		CwmpVersion:             cwmp.UNKNOWN_CWMP_VERSION,
		LastIncomingMessageTime: 0,
		LastOutgoingMessageTime: 0,
		XmlNamespaces:           map[xml.NamespaceID]xml.Namespace{},
		ActiveRPC:               nil,
	}

	return sessionInfo
}
