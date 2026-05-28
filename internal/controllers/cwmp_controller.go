package controllers

import (
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
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
	// Check to see if incoming request is part of an existing session
	var sessionID string
	var sessionInfo *session.SessionInfo

	sessionIDCookie, _ := r.Cookie(session.SessionCookieName)

	if sessionIDCookie != nil {
		sessionID = sessionIDCookie.Value
	}

	fmt.Println("Incoming session ID:", sessionID)

	if sessionID != "" {
		// Look in active session info table
		sessionInfo = session.SessionIdActiveSessions[sessionID]
		fmt.Println("Found session info:", sessionInfo)
	}

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

		if sessionID == "" || sessionInfo == nil {
			// No session ID or session info, so probably just a wayward empty post - we'll ignore it
			fmt.Println("Wayward empty post - ignoring")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if sessionInfo.SessionState != session.AWAITING_CPE_EMPTY_POST {
			// Shouldn't be getting an empty post at this stage in a session - better abandon the session
			fmt.Println(sessionID, ": got an empty post when we shouldn't - terminating session")
			sessionInfo.SessionState = session.TERMINATED
			delete(session.SessionIdActiveSessions, sessionID)
			delete(session.DeviceInfoActiveSessions, sessionInfo.DeviceID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// This means the CPE is done sending it's RPCs, so now the ACS can send some
		sessionInfo.SessionState = session.SENDING_ACS_RPCS
		sessionInfo.LastMessageTime = time.Now().Unix()

		outgoingMsg = findOutgoingRpc(sessionInfo)
	} else {
		if sessionInfo == nil {
			// No session info (which is possible when getting an Inform or if a CPE tries to send the session ID
			// cookie from a previous session), so we should probably think about setting up a new one

			if sessionID != "" {
				// No session info, but we got a session ID cookie - this is probably an expired session ID,
				// so we'll just log it and move on to creating a new session
				fmt.Println(sessionID + ": this is perhaps an expired session ID")
			}

			sessionInfo = session.CreateNewSession()
			sessionID = sessionInfo.SessionID
		}

		// Try parsing the incoming message
		incomingMsg, parseErr := parseIncomingMessage(xmlBytes, sessionInfo)
		if parseErr != nil {
			if xmlErr, ok := stderrors.AsType[*errors.XmlParsingError](parseErr); ok {
				// XML parsing error - just return a 400 with the error message
				errMsg := fmt.Sprintf("XML parsing error: %s", xmlErr.Message)
				fmt.Println(errMsg)
				http.Error(w, errMsg, http.StatusBadRequest)
				// w.WriteHeader(http.StatusBadRequest)
				return
			}

			if incomingMsgErr, ok := stderrors.AsType[*errors.IncomingMessageError](parseErr); ok {
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
			}
		}

		fmt.Println("Parsed incoming message:", incomingMsg)

		if incomingMsg.GetName() == "Inform" {
			// Inform gets initial special processing since it's kicking off a new session

			if sessionInfo.SessionState != session.NEW {
				// Picked up session ID/session info from a previous session
				// This shouldn't happen, but in case it does, we'll log it and then create a whole new session
				fmt.Printf("[%s] got an Inform message for a session that was already active (state: %+v) - this is unexpected, so terminating old session and starting new one\n", sessionID, sessionInfo.SessionState)

				sessionInfo.SessionState = session.TERMINATED
				delete(session.SessionIdActiveSessions, sessionID)
				delete(session.DeviceInfoActiveSessions, sessionInfo.DeviceID)

				newSessionInfo := session.CreateNewSession()
				newSessionInfo.CwmpVersion = sessionInfo.CwmpVersion
				newSessionInfo.XmlNamespaces = sessionInfo.XmlNamespaces
				fmt.Printf("New session info: %+v\n", newSessionInfo)
				sessionID = newSessionInfo.SessionID
				sessionInfo = newSessionInfo
			}

			sessionInfo.SessionState = session.INITIATING
			sessionInfo.DeviceID = incomingMsg.(*cwmp.Inform).DeviceId
			sessionInfo.LastMessageTime = time.Now().Unix()

			// Add session info to active session tables
			session.SessionIdActiveSessions[sessionID] = sessionInfo
			session.DeviceInfoActiveSessions[sessionInfo.DeviceID] = sessionInfo
		}

		sessionInfo.LastMessageTime = time.Now().Unix()

		if _, ok := incomingRequestNames[incomingMsg.GetName()]; ok {
			// This is a request from the CPE that requires a response from the ACS
			outgoingMsg = processIncomingRequest(sessionInfo, incomingMsg)
		} else {
			// This is a response from the CPE to a previous ACS request
			processIncomingResponse(sessionInfo, incomingMsg)

			// Look for the next ACS RPC to send back to the CPE (if any)
			outgoingMsg = findOutgoingRpc(sessionInfo)
		}
	}

	if outgoingMsg != nil {
		sendOutgoingMsg(w, sessionInfo, outgoingMsg)

		if outgoingMsg.GetName() == "InformResponse" {
			// After sending the InformResponse, we're waiting for an empty post from the CPE to indicate it's done sending its RPCs
			sessionInfo.SessionState = session.AWAITING_CPE_EMPTY_POST
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

func parseIncomingMessage(xmlBytes []byte, sessionInfo *session.SessionInfo) (cwmp.CwmpMessageInterface, error) {
	parsedEnv, err := xml.ParseSOAPEnvelope(xmlBytes)
	if err != nil {
		return nil, &errors.XmlParsingError{Message: fmt.Sprintf("Failed to parse SOAP envelope: %v", err)}
	}

	fmt.Printf("\nEnvelope: %s (%s)\n", parsedEnv.Envelope.Name.Local, parsedEnv.Envelope.Name.Space)
	fmt.Printf("Header:   %s (%s)\n", parsedEnv.Header.Name.Local, parsedEnv.Header.Name.Space)
	fmt.Printf("Body:     %s (%s)\n", parsedEnv.Body.Name.Local, parsedEnv.Body.Name.Space)
	fmt.Println("\nNamespaces:")
	for prefix, uri := range parsedEnv.Namespaces {
		if prefix == "" {
			fmt.Printf("  (default) => %s\n", uri)
			continue
		}
		fmt.Printf("  %s => %s\n", prefix, uri)
	}

	namespaceMap := xml.BuildNamespaceMap(parsedEnv.Namespaces)
	sessionInfo.XmlNamespaces = namespaceMap

	// _, cwmpNS := findCwmpNamespace(parsedEnv.Namespaces)
	cpeHeader := xml.ParseCPEHeader(parsedEnv.Header, namespaceMap[xml.CWMP].URL)

	fmt.Println("\nParsed CPE header:", cpeHeader)

	rpcName := parsedEnv.Body.Children[0].Name.Local
	fmt.Println("RPC Name:", rpcName)

	cwmpVersion := determineCwmpVersion(parsedEnv.Namespaces, cpeHeader)
	fmt.Println(cwmpVersion)

	if rpcName == "Inform" {
		sessionInfo.CwmpVersion = cwmpVersion
	} else if cwmpVersion != sessionInfo.CwmpVersion {
		// Cwmp version of incoming message doesn't match what's already been established in the session info,
		// so let's return a fault
		return nil, &errors.IncomingMessageError{
			Header:      cpeHeader,
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8801, // Version mismatch
			FaultString: fmt.Sprintf("CWMP version mismatch - expected %s but got %s", sessionInfo.CwmpVersion.String(), cwmpVersion.String()),
		}
	}

	return parseCwmpMessageViaMap(sessionInfo, rpcName, parsedEnv.Body.Children[0], cpeHeader)
}

var cwmpNSPattern = regexp.MustCompile(`^urn:dslforum-org:cwmp-\d+-\d+$`)

func findCwmpNamespace(namespaces map[string]string) (string, string) {
	for prefix, uri := range namespaces {
		if cwmpNSPattern.MatchString(uri) {
			return prefix, uri
		}
	}

	return "", ""
}

func determineCwmpVersion(namespaces map[string]string, header cwmp.CwmpHeader) cwmp.SupportedCwmpVersion {
	cwmpVersion := cwmp.UNKNOWN_CWMP_VERSION

	// Start by looking for incoming SupportedCWMPVersions header, which means CPE supports at least 1.4
	if header.SupportedCWMPVersions != "" {
		return cwmp.MaxCommonVersion(header.SupportedCWMPVersions)
	}

	// No SupportedCWMPVersions header, so now we have to infer via namespace URI and other headers
	for _, uri := range namespaces {
		switch uri {
		case xml.CWMP_1_0_NS_URL:
			return cwmp.CWMP_V1_0
		case xml.CWMP_1_1_NS_URL:
			return cwmp.CWMP_V1_1
		case xml.CWMP_1_2_NS_URL:
			// With this URI, the version could be 1.2 or 1.3
			if header.SessionTimeout != "" {
				// SessionTimeout was added in 1.3, so if it's present then we know it's 1.3
				return cwmp.CWMP_V1_3
			} else {
				// No SessionTimeout, so assume 1.2
				return cwmp.CWMP_V1_2
			}
		}
	}

	return cwmpVersion
}

func parseCwmpMessageViaMap(sessionInfo *session.SessionInfo, rpcName string, elem xml.SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	// First try to find a parser specific to the specified version and message name
	if versionParsers, ok := cwmpMessageParsers[sessionInfo.CwmpVersion]; ok {
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
