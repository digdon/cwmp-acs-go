package session

import (
	"github.com/google/uuid"

	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/xml"
)

const SessionCookieName = "CWMP-Session-ID"

type SessionState int

const (
	NEW SessionState = iota
	INITIATING
	AWAITING_CPE_EMPTY_POST
	SENDING_ACS_RPCS
	TERMINATED
)

type SessionInfo struct {
	SessionID       string
	DeviceID        cwmp.DeviceId
	SessionState    SessionState
	CwmpVersion     cwmp.SupportedCwmpVersion
	LastMessageTime int64
	XmlNamespaces   map[xml.NamespaceID]xml.Namespace
}

var SessionIdActiveSessions = map[string]*SessionInfo{}
var DeviceInfoActiveSessions = map[cwmp.DeviceId]*SessionInfo{}

func CreateNewSession() *SessionInfo {
	sessionID := uuid.NewString()

	sessionInfo := &SessionInfo{
		SessionID:       sessionID,
		DeviceID:        cwmp.DeviceId{},
		SessionState:    NEW,
		CwmpVersion:     cwmp.UNKNOWN_CWMP_VERSION,
		LastMessageTime: 0,
		XmlNamespaces:   map[xml.NamespaceID]xml.Namespace{},
	}

	return sessionInfo
}
