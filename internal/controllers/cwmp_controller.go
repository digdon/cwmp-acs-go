package controllers

import (
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/errors"
	"cwmp-acs/internal/session"
	"cwmp-acs/internal/xml"
	cwmp14 "cwmp-acs/internal/xml/cwmp_1_4"
)

var CwmpControllerPath = "/cwmp"

// Set up the versioned parsers map
var cwmpMessageParsers = map[cwmp.SupportedCwmpVersion]map[string]xml.MessageParser{
	cwmp.CWMP_V1_4: cwmp14.CwmpMessageParsers,
}

var defaultCwmpMessageParsers = cwmp14.CwmpMessageParsers

// Set up the versioned generators map
var cwmpMessageGenerators = map[cwmp.SupportedCwmpVersion]map[string]xml.MessageGenerator{
	cwmp.CWMP_V1_4: cwmp14.CwmpMessageGenerators,
}

var defaultCwmpMessageGenerators = cwmp14.CwmpMessageGenerators

var incomingRequestNames = map[string]bool{
	"Inform":           true,
	"GetPRCMethods":    true,
	"TransferComplete": true,
}

func CwmpHandler(w http.ResponseWriter, r *http.Request) {
	// Gather up any existing session info, based on session ID cookie
	var sessionID string
	var sessionInfo *session.SessionInfo

	sessionIDCookie, _ := r.Cookie(session.SessionCookieName)

	if sessionIDCookie != nil {
		sessionID = sessionIDCookie.Value
	}

	fmt.Println("Incoming session ID:", sessionID)

	// Read in the incoming message body as bytes (we'll parse it later)
	xmlBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Debug stuff - incoming message stuff
	xmlString := string(xmlBytes)
	if len(xmlString) == 0 {
		xmlString = "<empty>"
	}
	fmt.Printf("Received message:\n%s\n", xmlString)

	var outgoingMsg cwmp.CwmpMessageInterface

	if len(xmlBytes) == 0 {
		// Empty post
		fmt.Printf("[%s] received an empty post\n", sessionID)

		if sessionID != "" {
			// Look for session in active session info table
			sessionInfo = session.SessionIdActiveSessions[sessionID]
			fmt.Printf("[%s] Found session info: %+v\n", sessionID, sessionInfo)
		}

		if sessionID == "" || sessionInfo == nil {
			// No session ID or session info, so probably just a wayward empty post - we'll ignore it
			fmt.Printf("[%s] Wayward empty post - ignoring\n", sessionID)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if sessionInfo.SessionState != session.RECEIVING_CPE_RPCS {
			// Shouldn't be getting an empty post at this stage in a session - better abandon the session
			// Although, perhaps we don't want to actually kill the session, instead returning an error
			// and letting the automated session cleaner eventually clean it up.
			fmt.Printf("[%s] got an empty post when we shouldn't - terminating session\n", sessionID)
			sessionInfo.SessionState = session.TERMINATED
			delete(session.SessionIdActiveSessions, sessionID)
			delete(session.DeviceInfoActiveSessions, sessionInfo.DeviceID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// This means the CPE is done sending it's RPCs, so now the ACS can send some
		sessionInfo.SessionState = session.SENDING_ACS_RPCS
		sessionInfo.LastIncomingMessageTime = time.Now().Unix()

		outgoingMsg = findOutgoingRpc(sessionInfo)
	} else {
		outgoingMsg, sessionInfo, err = handleIncomingMessage(xmlBytes, sessionID)
		if err != nil {
			if xmlErr, ok := stderrors.AsType[*errors.XmlParsingError](err); ok {
				// XML parsing error - just return a 400 with the error message
				errMsg := fmt.Sprintf("XML parsing error: %s", xmlErr.Message)
				fmt.Println(errMsg)
				http.Error(w, errMsg, http.StatusBadRequest)
				// w.WriteHeader(http.StatusBadRequest)
				return
			} else if incomingMsgErr, ok := stderrors.AsType[*errors.IncomingMessageError](err); ok {
				// CWMP message parsing error - we need to return a SOAP fault
				fault := cwmp.Fault{
					CwmpMessage: cwmp.CwmpMessage{
						Name:       "Fault",
						CwmpHeader: incomingMsgErr.Header,
					},
					Source:      incomingMsgErr.Source,
					FaultCode:   incomingMsgErr.FaultCode,
					FaultString: incomingMsgErr.FaultString,
				}
				sendOutgoingMsg(w, sessionInfo, &fault)
				return
			} else {
				// Some other kind of error - we'll just return a 500 with the error message
				errMsg := fmt.Sprintf("Internal server error: %v", err)
				fmt.Println(errMsg)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}
		}
	}

	if outgoingMsg != nil {
		sendOutgoingMsg(w, sessionInfo, outgoingMsg)
		sessionInfo.LastOutgoingMessageTime = time.Now().Unix()

		if outgoingMsg.GetName() == "InformResponse" {
			// After sending the InformResponse, we're ready to accept incoming RPCs from the CPE
			sessionInfo.SessionState = session.RECEIVING_CPE_RPCS
		}
	} else {
		// No outgoing RPCs, so we're done. Let's terminate the session
		fmt.Printf("[%s] no outgoing messages - terminating session\n", sessionID)
		sessionInfo.SessionState = session.TERMINATED
		delete(session.SessionIdActiveSessions, sessionID)
		delete(session.DeviceInfoActiveSessions, sessionInfo.DeviceID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleIncomingMessage(xmlBytes []byte, sessionID string) (cwmp.CwmpMessageInterface, *session.SessionInfo, error) {
	parsedEnv, err := xml.ParseSOAPEnvelope(xmlBytes)
	if err != nil {
		return nil, nil, &errors.XmlParsingError{Message: fmt.Sprintf("Failed to parse SOAP message: %v", err)}
	}

	namespaceMap := xml.BuildNamespaceMap(parsedEnv.Namespaces)

	if !isValidNamespaceMap(namespaceMap) {
		return nil, nil, &errors.XmlParsingError{Message: "Invalid namespaces in SOAP message"}
	}

	// Parse out the Header contents
	cpeHeader := xml.ParseCPEHeader(parsedEnv.Header, namespaceMap[xml.CWMP].URL)

	// Extract the RPC name from the Body
	rpcName := parsedEnv.Body.Children[0].Name.Local

	// Check to see if the incoming Header content is valid
	if valid, faultCode, faultString := isHeaderValidForMessageType(cpeHeader, rpcName); !valid {
		return nil, nil, &errors.IncomingMessageError{
			Header:      cpeHeader,
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   faultCode,
			FaultString: faultString,
		}
	}

	var sessionInfo *session.SessionInfo

	if sessionID != "" {
		// Look in active session info table
		sessionInfo = session.SessionIdActiveSessions[sessionID]
		fmt.Printf("[%s] Found session info: %+v\n", sessionID, sessionInfo)
	}

	var parsedMsg cwmp.CwmpMessageInterface

	if rpcName == "Inform" {
		// Informs get special handling, since they kick off new CWMP sessions

		if sessionInfo != nil {
			// Got session info from a previous session - this shouldn't happen, but we'll log it
			fmt.Printf("[%s] got an Inform message with a session ID cookie that points to an existing session - this is unexpected [%+v]\n", sessionID, sessionInfo.SessionState)

			// Do we, at some point, want to terminate the old session? For now we'll let the automated session cleaner eventually handle this
		} else if sessionID != "" {
			// Got a session ID from a previous session - this also shouldn't happen, but again, we'll log it
			fmt.Printf("[%s] got an Inform message with a session ID cookie from a different session - this is unexpected\n", sessionID)
		}

		// Set up a new session
		sessionInfo = session.CreateNewSession()
		sessionID = sessionInfo.SessionID

		fmt.Printf("[%s] Created new session info: %+v\n", sessionID, sessionInfo)

		// Since this is possibly an Inform message, let's work out which version of CWMP to use
		cwmpVersion := determineCwmpVersion(cpeHeader, namespaceMap[xml.CWMP].URL)

		sessionInfo.CwmpVersion = cwmpVersion
		sessionInfo.XmlNamespaces = namespaceMap

		fmt.Printf("[%s] Determined CWMP version: %s\n", sessionID, cwmpVersion)

		parsedMsg, err = parseCwmpMessageViaMap(cwmpVersion, rpcName, parsedEnv.Body.Children[0], cpeHeader)
		if err != nil {
			return nil, sessionInfo, err
		}

		// Successful parse, so let's register the session
		sessionInfo.DeviceID = parsedMsg.(*cwmp.Inform).DeviceId
		sessionInfo.SessionState = session.INITIATING
		session.SessionIdActiveSessions[sessionID] = sessionInfo
		session.DeviceInfoActiveSessions[sessionInfo.DeviceID] = sessionInfo
	} else {
		// Everything else
		if sessionID == "" || sessionInfo == nil {
			// No session ID or session info, so we don't know what to do with this message - we'll just return an error
			fmt.Printf("[%s] got a %s message, but there's no valid session ID cookie\n", sessionID, rpcName)
			return nil, nil, &errors.IncomingMessageError{
				Header:      cpeHeader,
				Source:      cwmp.FaultSourceCPE,
				FaultCode:   8001, // Invalid session state
				FaultString: "Invalid session state",
			}
		}

		if sessionInfo.SessionState != session.RECEIVING_CPE_RPCS && sessionInfo.SessionState != session.SENDING_ACS_RPCS {
			// Session is in a state where we shouldn't be getting messages from the CPE, so we'll return an error
			fmt.Printf("[%s] got a %s message, but the current session state is invalid: %+v\n", sessionID, rpcName, sessionInfo.SessionState)
			return nil, nil, &errors.IncomingMessageError{
				Header:      cpeHeader,
				Source:      cwmp.FaultSourceCPE,
				FaultCode:   8001, // Invalid session state
				FaultString: "Invalid session state",
			}
		}

		parsedMsg, err = parseCwmpMessageViaMap(sessionInfo.CwmpVersion, rpcName, parsedEnv.Body.Children[0], cpeHeader)
		if err != nil {
			return nil, sessionInfo, err
		}
	}

	sessionInfo.LastIncomingMessageTime = time.Now().Unix()

	var outgoingMsg cwmp.CwmpMessageInterface

	if _, ok := incomingRequestNames[parsedMsg.GetName()]; ok {
		// This is a request from the CPE that requires a response from the ACS
		outgoingMsg = processIncomingRequest(sessionInfo, parsedMsg)
	} else {
		// This is a response from the CPE to a previous ACS request
		processIncomingResponse(sessionInfo, parsedMsg)

		// Look for the next ACS RPC to send back to the CPE (if any)
		outgoingMsg = findOutgoingRpc(sessionInfo)
	}

	return outgoingMsg, sessionInfo, nil
}

func isValidNamespaceMap(namespaceMap map[xml.NamespaceID]xml.Namespace) bool {
	if _, ok := namespaceMap[xml.SOAPENV]; !ok {
		return false
	}
	if _, ok := namespaceMap[xml.SOAPENC]; !ok {
		return false
	}
	if _, ok := namespaceMap[xml.XSD]; !ok {
		return false
	}
	if _, ok := namespaceMap[xml.XSI]; !ok {
		return false
	}
	if _, ok := namespaceMap[xml.CWMP]; !ok {
		return false
	}

	return true
}

func isHeaderValidForMessageType(header cwmp.CwmpHeader, messageType string) (bool, int, string) {
	// Start with items in the header that can never be sent by the CPE for any message type
	if header.UseCWMPVersion != "" {
		return false, 8003, "UseCWMPVersion cannot be sent by CPE"
	} else if header.HoldRequests != "" {
		return false, 8003, "HoldRequests cannot be sent by CPE"
	}

	if messageType != "Inform" {
		if header.SupportedCWMPVersions != "" {
			return false, 8003, "SupportedCWMPVersions cannot be sent by CPE for non-Inform messages"
		} else if header.SessionTimeout != "" {
			return false, 8003, "SessionTimeout cannot be sent by CPE for non-Inform messages"
		}
	}

	return true, 0, ""
}

func determineCwmpVersion(cpeHeader cwmp.CwmpHeader, cwmpNSUrl string) cwmp.SupportedCwmpVersion {
	cwmpVersion := cwmp.UNKNOWN_CWMP_VERSION

	// Start by looking for incoming SupportedCWMPVersions header, which means CPE supports at least 1.4
	if cpeHeader.SupportedCWMPVersions != "" {
		return cwmp.MaxCommonVersion(cpeHeader.SupportedCWMPVersions)
	}

	// No SupportedCWMPVersions header, so now we have to infer via namespace URI and (possibly) other headers
	switch cwmpNSUrl {
	case xml.CWMP_1_0_NS_URL:
		return cwmp.CWMP_V1_0
	case xml.CWMP_1_1_NS_URL:
		return cwmp.CWMP_V1_1
	case xml.CWMP_1_2_NS_URL:
		// With this URI, the version could be 1.2 or 1.3
		if cpeHeader.SessionTimeout != "" {
			// SessionTimeout was added in 1.3, so if it's present then we know it's 1.3
			return cwmp.CWMP_V1_3
		} else {
			// No SessionTimeout, so assume 1.2
			return cwmp.CWMP_V1_2
		}
	}

	return cwmpVersion
}

func processIncomingRequest(sessionInfo *session.SessionInfo, incomingMsg cwmp.CwmpMessageInterface) cwmp.CwmpMessageInterface {
	switch incomingMsg.GetName() {
	case "Inform":
		informResponse := &cwmp.InformResponse{
			CwmpMessage: cwmp.CwmpMessage{
				Name: "InformResponse",
				CwmpHeader: cwmp.CwmpHeader{
					ID:             incomingMsg.GetID(),
					UseCWMPVersion: sessionInfo.CwmpVersion.String(),
				},
			},
			MaxEnvelopes: 1,
		}
		return informResponse
	default:
		// For now, we'll just return a generic response for any unsupported RPCs, but in the future we might want to return a fault or something more specific
		fmt.Printf("Received unsupported RPC method: %s\n", incomingMsg.GetName())
		return nil
	}
}

func processIncomingResponse(sessionInfo *session.SessionInfo, incomingMsg cwmp.CwmpMessageInterface) {
	// This is where we would implement any special processing for incoming responses to previous ACS requests
	// For now, we don't have any ACS requests that expect responses, so we'll just log the incoming response

	fmt.Printf("Processing incoming response: %+v\n", incomingMsg)
}

func findOutgoingRpc(sessionInfo *session.SessionInfo) cwmp.CwmpMessageInterface {
	// This is where we would look for the next ACS RPC to send back to the CPE (if any)
	// For now, we don't have any ACS RPCs to send, so we'll just return nil

	return nil
}

func parseCwmpMessageViaMap(cwmpVersion cwmp.SupportedCwmpVersion, rpcName string, elem xml.SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	// First try to find a parser specific to the specified version and message name
	if versionParsers, ok := cwmpMessageParsers[cwmpVersion]; ok {
		if parser, ok := versionParsers[rpcName]; ok {
			return parser(elem, cpeHeader)
		}
	}

	// No version-specific parser, so now try to find a parser in the default set
	if parser, ok := defaultCwmpMessageParsers[rpcName]; ok {
		return parser(elem, cpeHeader)
	}

	return nil, &errors.IncomingMessageError{
		Header:      cpeHeader,
		Source:      cwmp.FaultSourceCPE,
		FaultCode:   8000, // Method not supported
		FaultString: fmt.Sprintf("Unsupported RPC method: %s", rpcName),
	}
}

func generateXmlMessageViaMap(sessionInfo *session.SessionInfo, message cwmp.CwmpMessageInterface) (string, error) {
	// First try to find a generator specific to the specified version and message name
	if versionGenerators, ok := cwmpMessageGenerators[sessionInfo.CwmpVersion]; ok {
		if generator, ok := versionGenerators[message.GetName()]; ok {
			return generator(message, sessionInfo.XmlNamespaces)
		}
	}

	// No version-specific generator, so now try to find a generator in the default set
	if generator, ok := defaultCwmpMessageGenerators[message.GetName()]; ok {
		return generator(message, sessionInfo.XmlNamespaces)
	}

	return "", fmt.Errorf("no generator found for message: %s", message.GetName())
}

func sendOutgoingMsg(w http.ResponseWriter, sessionInfo *session.SessionInfo, message cwmp.CwmpMessageInterface) {
	fmt.Printf("Outgoing message: %+v\n", message)

	xmlString, err := generateXmlMessageViaMap(sessionInfo, message)
	if err != nil {
		fmt.Printf("Error generating XML: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("Generated XML:\n%s\n", xmlString)

	if message.GetName() == "InformResponse" {
		http.SetCookie(w, &http.Cookie{
			Name:  session.SessionCookieName,
			Value: sessionInfo.SessionID,
			Path:  CwmpControllerPath,
			// HttpOnly: true,
		})
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xmlString))
}
