package cwmp_1_4

import (
	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/errors"
	"cwmp-acs/internal/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var CwmpMessageParsers = map[string]xml.MessageParser{
	// "AddObjectResponse":              ParseAddObjectResponse,
	// "DeleteObjectResponse":           ParseDeleteObjectResponse,
	// "DownloadResponse":               ParseDownloadResponse,
	"Fault": ParseFault,
	// "GetParameterAttributesResponse": ParseGetParameterAttributesResponse,
	"GetParameterNamesResponse": ParseGetParameterNamesResponse,
	// "GetParameterValuesResponse":     ParseGetParameterValuesResponse,
	"GetRPCMethods":                  ParseGetRPCMethods,
	"GetRPCMethodsResponse":          ParseGetRPCMethodsResponse,
	"Inform":                         ParseInform,
	"RebootResponse":                 ParseRebootResponse,
	"SetParameterAttributesResponse": ParseSetParameterAttributesResponse,
	// "SetParameterValuesResponse": ParseSetParameterValuesResponse,
	// "TransferComplete":           ParseTransferComplete,
}

type SOAPElement = xml.SOAPElement

func ParseFault(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	fmt.Println("Parsing Fault")

	fault := cwmp.Fault{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "Fault", CwmpHeader: cpeHeader,
		},
	}

	for _, child := range elem.Children {
		switch child.Name.Local {
		case "faultcode":
			value := child.Text
			switch strings.ToLower(value) {
			case "client":
				fault.Source = cwmp.FaultSourceCPE
			case "server":
				fault.Source = cwmp.FaultSourceServer
			}

		case "faultstring":
			fault.FaultString = child.Text

		case "detail":
			for _, detailChild := range child.Children {
				if detailChild.Name.Local == "Fault" {
					for _, faultChild := range detailChild.Children {
						switch faultChild.Name.Local {
						case "FaultCode":
							faultCode, err := strconv.Atoi(faultChild.Text)
							if err != nil {
								return nil, fmt.Errorf("invalid FaultCode value in Fault detail: %v", err)
							}
							fault.FaultCode = faultCode

						case "FaultString":
							fault.FaultString = faultChild.Text
						}
					}
				}
			}
		}
	}

	return &fault, nil
}

var parameterInfoStructRegex = regexp.MustCompile(`^[^:]+:ParameterInfoStruct\[(\d+)\]$`)

func ParseGetParameterNamesResponse(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	fmt.Println("Parsing GetParameterNamesResponse")

	getParameterNamesResponse := cwmp.GetParameterNamesResponse{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "GetParameterNamesResponse", CwmpHeader: cpeHeader,
		},
	}

	for _, child := range elem.Children {
		switch child.Name.Local {
		case "ParameterList":
			parameters, err := parseParameterInfoStructList(child)
			if err != nil {
				return nil, err
			}
			getParameterNamesResponse.ParameterList = parameters
		}
	}

	return &getParameterNamesResponse, nil
}

func parseParameterInfoStructList(elem SOAPElement) ([]cwmp.ParameterInfoStruct, *errors.IncomingMessageError) {
	paramList := []cwmp.ParameterInfoStruct{}

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
			FaultString: "ParameterList is missing arrayType attribute",
		}
	}

	// Now verify the arrayType is correct for an array of ParameterInfoStruct
	matches := parameterInfoStructRegex.FindStringSubmatch(arrayType)
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

	// Now parse each ParameterInfoStruct item in the list
	for _, paramChild := range elem.Children {
		if paramChild.Name.Local != "ParameterInfoStruct" {
			return nil, &errors.IncomingMessageError{
				Source:      cwmp.FaultSourceCPE,
				FaultCode:   8003, // Invalid arguments
				FaultString: fmt.Sprintf("unexpected element in ParameterList: %s", paramChild.Name.Local),
			}
		}

		var param cwmp.ParameterInfoStruct
		for _, paramStructChild := range paramChild.Children {
			switch paramStructChild.Name.Local {
			case "Name":
				param.Name = paramStructChild.Text
			case "Writable":
				var err error
				param.Writable, err = strconv.ParseBool(strings.TrimSpace(paramStructChild.Text))
				if err != nil {
					return nil, &errors.IncomingMessageError{
						Source:      cwmp.FaultSourceCPE,
						FaultCode:   8003, // Invalid arguments
						FaultString: fmt.Sprintf("invalid boolean value for Writable in ParameterInfoStruct: %v", err),
					}
				}
			default:
				return nil, &errors.IncomingMessageError{
					Source:      cwmp.FaultSourceCPE,
					FaultCode:   8003, // Invalid arguments
					FaultString: fmt.Sprintf("unexpected element in ParameterInfoStruct: %s", paramStructChild.Name.Local),
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

func ParseGetRPCMethods(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	getRPCMethods := cwmp.GetRPCMethods{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "GetRPCMethods", CwmpHeader: cpeHeader,
		},
	}

	return &getRPCMethods, nil
}

func ParseGetRPCMethodsResponse(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	getRPCMethodsResponse := cwmp.GetRPCMethodsResponse{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "GetRPCMethodsResponse", CwmpHeader: cpeHeader,
		},
	}

	for _, child := range elem.Children {
		switch child.Name.Local {
		case "MethodList":
			methods, err := parseMethodList(child)
			if err != nil {
				return nil, err
			}
			getRPCMethodsResponse.MethodList = methods
		}
	}

	return &getRPCMethodsResponse, nil
}

var methodListRegex = regexp.MustCompile(`^[^:]+:string\[(\d+)\]$`)

func parseMethodList(elem SOAPElement) ([]string, *errors.IncomingMessageError) {
	methodList := []string{}

	// Start by looking for the arrayType attribute to confirm this is an array of string
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
			FaultString: "MethodList is missing arrayType attribute",
		}
	}

	// Now verify the arrayType is correct for an array of string
	matches := methodListRegex.FindStringSubmatch(arrayType)
	if len(matches) != 2 {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("invalid arrayType for MethodList: %s", arrayType),
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

	// Now parse each method name in the list
	for _, methodChild := range elem.Children {
		if methodChild.Name.Local != "string" {
			return nil, &errors.IncomingMessageError{
				Source:      cwmp.FaultSourceCPE,
				FaultCode:   8003, // Invalid arguments
				FaultString: fmt.Sprintf("unexpected element in MethodList: %s", methodChild.Name.Local),
			}
		}
		methodList = append(methodList, methodChild.Text)
	}

	if len(methodList) != itemCount {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: fmt.Sprintf("MethodList item count mismatch: expected %d, got %d", itemCount, len(methodList)),
		}
	}

	return methodList, nil
}

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
			inform.Parameters = paramList
		}
	}

	valid, validationError := isValidInform(inform)
	if !valid {
		return nil, &errors.IncomingMessageError{
			Source:      cwmp.FaultSourceCPE,
			FaultCode:   8003, // Invalid arguments
			FaultString: validationError,
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

var eventStructRegex = regexp.MustCompile(`^[^:]+:EventStruct\[(\d+)\]$`)

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
			FaultString: "Event list is missing arrayType attribute",
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

var parameterValueStructRegex = regexp.MustCompile(`^[^:]+:ParameterValueStruct\[(\d+)\]$`)

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
			FaultString: "ParameterList is missing arrayType attribute",
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

	// Now parse each ParameterValueStruct item in the list
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

func isValidInform(inform cwmp.Inform) (bool, string) {
	if inform.DeviceId.Manufacturer == "" {
		return false, "Inform is missing DeviceId"
	}

	if len(inform.Events) == 0 {
		return false, "Inform has no events, but should have at least one"
	}

	if len(inform.Parameters) == 0 {
		return false, "Inform is missing ParameterList"
	}

	hasCrUrl := false
	for _, param := range inform.Parameters {
		if strings.HasSuffix(param.Name, "ConnectionRequestURL") {
			hasCrUrl = true
			break
		}
	}

	for name, found := range checkedParams {
		if !found {
			return false, fmt.Sprintf("Inform is missing required parameter: %s", name)
		}
	}

	return true, ""
}

var minimumForcedInformParameters = []string{
	"DeviceInfo.HardwareVersion",
	"DeviceInfo.SoftwareVersion",
	"DeviceInfo.ProvisioningCode",
	"ManagementServer.ConnectionRequestURL",
	"ManagementServer.ParameterKey",
}

func ParseRebootResponse(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	rebootResponse := cwmp.RebootResponse{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "RebootResponse", CwmpHeader: cpeHeader,
		},
	}

	return &rebootResponse, nil
}

func ParseSetParameterAttributesResponse(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	setParameterAttributesResponse := cwmp.SetParameterAttributesResponse{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "SetParameterAttributesResponse", CwmpHeader: cpeHeader,
		},
	}

	return &setParameterAttributesResponse, nil
}
