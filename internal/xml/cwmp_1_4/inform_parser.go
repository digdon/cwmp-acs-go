package cwmp_1_4

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/errors"
	cwmpxml "cwmp-acs/internal/xml"
)

type SOAPElement = cwmpxml.SOAPElement

var eventStructRegex = regexp.MustCompile(`^[^:]+:EventStruct\[(\d+)\]$`)
var parameterValueStructRegex = regexp.MustCompile(`^[^:]+:ParameterValueStruct\[(\d+)\]$`)

func ParseInform(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	fmt.Println("Parsing Inform")

	inform := cwmp.Inform{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "Inform", CwmpHeader: cpeHeader,
		},
	}

	for _, child := range elem.Children {
		switch child.Name.Local {
		case "DeviceId":
			deviceID, err := parseDeviceId(child)
			if err != nil {
				return nil, err
			}
			inform.DeviceId = deviceID

		case "Event":
			events, err := parseEventList(child)
			if err != nil {
				return nil, err
			}
			inform.Events = events

		case "MaxEnvelopes":
			maxEnvelopes, err := strconv.Atoi(child.Text)
			if err != nil {
				return nil, &errors.IncomingMessageError{
					Source:      cwmp.FaultSourceCPE,
					FaultCode:   8003, // Invalid arguments
					FaultString: "Inform message contains invalid MaxEnvelopes value",
				}
			}
			inform.MaxEnvelopes = maxEnvelopes

		case "CurrentTime":
			inform.CurrentTime = child.Text

		case "RetryCount":
			retryCount, err := strconv.Atoi(child.Text)
			if err != nil {
				return nil, &errors.IncomingMessageError{
					Source:      cwmp.FaultSourceCPE,
					FaultCode:   8003, // Invalid arguments
					FaultString: "Inform message contains invalid RetryCount value",
				}
			}
			inform.RetryCount = retryCount

		case "ParameterList":
			paramList, err := parseParameterList(child)
			if err != nil {
				return nil, err
			}
			inform.ParamList = paramList
		}
	}

	if !isValid(inform) {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: "Inform message is not valid",
		}
	}

	return &inform, nil
}

func parseDeviceId(elem SOAPElement) (cwmp.DeviceId, *errors.IncomingMessageError) {
	deviceID := cwmp.DeviceId{}

	for _, deviceIDChild := range elem.Children {
		switch deviceIDChild.Name.Local {
		case "Manufacturer":
			deviceID.Manufacturer = deviceIDChild.Text
		case "OUI":
			deviceID.OUI = deviceIDChild.Text
		case "ProductClass":
			deviceID.ProductClass = deviceIDChild.Text
		case "SerialNumber":
			deviceID.SerialNumber = deviceIDChild.Text
			// default:
			// 	return cwmp.DeviceId{}, fmt.Errorf("unexpected element in DeviceId: %s", deviceIDChild.Name.Local)
		}
	}

	if deviceID.Manufacturer == "" {
		return cwmp.DeviceId{}, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: "Inform is missing Manufacturer element or it contains an empty value",
		}
	}

	if deviceID.OUI == "" {
		return cwmp.DeviceId{}, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: "Inform is missing OUI element or it contains an empty value",
		}
	}

	if deviceID.SerialNumber == "" {
		return cwmp.DeviceId{}, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: "Inform is missing SerialNumber element or it contains an empty value",
		}
	}

	return deviceID, nil
}

func parseEventList(elem SOAPElement) ([]cwmp.Event, *errors.IncomingMessageError) {
	eventList := []cwmp.Event{}

	// Start by looking for the arrayType attribute to confirm this is an array of EventStruct
	var arrayType string
	for _, attr := range elem.Attrs {
		if attr.Name.Local == "arrayType" {
			arrayType = attr.Value
			break
		}
	}

	if arrayType == "" {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: "Inform Event list is missing arrayType attribute",
		}
	}

	// Now verify the arrayType is correct for an array of EventStruct
	matches := eventStructRegex.FindStringSubmatch(arrayType)
	if len(matches) != 2 {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("invalid arrayType for Event list: %s", arrayType),
		}
	}

	// Capture the item count from the arrayType for later verification
	itemCount, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("invalid item count in arrayType: %v", err),
		}
	}

	for _, eventChild := range elem.Children {
		if eventChild.Name.Local != "EventStruct" {
			return nil, &errors.IncomingMessageError{
				Source:      cwmp.FaultSourceCPE,
				FaultCode:   8003, // Invalid arguments
				FaultString: fmt.Sprintf("unexpected element in Event list: %s", eventChild.Name.Local),
			}
		}

		var event cwmp.Event
		for _, eventStructChild := range eventChild.Children {
			switch eventStructChild.Name.Local {
			case "EventCode":
				event.EventCode = eventStructChild.Text
			case "CommandKey":
				event.CommandKey = eventStructChild.Text
			default:
				return nil, &errors.IncomingMessageError{
					Source:      cwmp.FaultSourceCPE,
					FaultCode:   8003, // Invalid arguments
					FaultString: fmt.Sprintf("unexpected element in EventStruct: %s", eventStructChild.Name.Local),
				}
			}
		}
		eventList = append(eventList, event)
	}

	if len(eventList) != itemCount {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("Event item count mismatch: expected %d, got %d", itemCount, len(eventList)),
		}
	}

	return eventList, nil
}

func parseParameterList(elem SOAPElement) ([]cwmp.ParameterValueStruct, *errors.IncomingMessageError) {
	paramList := []cwmp.ParameterValueStruct{}

	// Start by looking for the arrayType attribute to confirm this is an array of ParameterValueStruct
	var arrayType string
	for _, attr := range elem.Attrs {
		if attr.Name.Local == "arrayType" {
			arrayType = attr.Value
			break
		}
	}

	if arrayType == "" {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: "Inform ParameterList is missing arrayType attribute",
		}
	}

	// Now verify the arrayType is correct for an array of ParameterValueStruct
	matches := parameterValueStructRegex.FindStringSubmatch(arrayType)
	if len(matches) != 2 {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("invalid arrayType for ParameterList: %s", arrayType),
		}
	}

	// Capture the item count from the arrayType for later verification
	itemCount, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("invalid item count in arrayType: %v", err),
		}
	}

	for _, paramChild := range elem.Children {
		if paramChild.Name.Local != "ParameterValueStruct" {
			return nil, &errors.IncomingMessageError{
				Source:      cwmp.FaultSourceCPE,
				FaultCode:   8003, // Invalid arguments
				FaultString: fmt.Sprintf("unexpected element in ParameterList: %s", paramChild.Name.Local),
			}
		}

		var param cwmp.ParameterValueStruct
		for _, paramStructChild := range paramChild.Children {
			switch paramStructChild.Name.Local {
			case "Name":
				param.Name = paramStructChild.Text
			case "Value":
				param.Value = paramStructChild.Text
				for _, attr := range paramStructChild.Attrs {
					if attr.Name.Local == "type" {
						pos := strings.LastIndex(attr.Value, ":")
						param.Type = attr.Value[pos+1:]
						break
					}
				}
			default:
				return nil, &errors.IncomingMessageError{
					Source:      cwmp.FaultSourceCPE,
					FaultCode:   8003, // Invalid arguments
					FaultString: fmt.Sprintf("unexpected element in ParameterValueStruct: %s", paramStructChild.Name.Local),
				}
			}
		}
		paramList = append(paramList, param)
	}

	if len(paramList) != itemCount {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("ParameterList item count mismatch: expected %d, got %d", itemCount, len(paramList)),
		}
	}

	return paramList, nil
}

func isValid(inform cwmp.Inform) bool {
	if len(inform.ParamList) == 0 {
		return false
	}

	hasCrUrl := false
	for _, param := range inform.ParamList {
		if strings.HasSuffix(param.Name, "ConnectionRequestURL") {
			hasCrUrl = true
			break
		}
	}

	return hasCrUrl
}
