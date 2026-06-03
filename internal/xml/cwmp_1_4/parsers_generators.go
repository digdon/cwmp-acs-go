package cwmp_1_4

import (
	"cwmp-acs/internal/xml"
)

var CwmpMessageParsers = map[string]xml.MessageParser{
	"GetRPCMethods": ParseGetRPCMethods,
	"Inform":        ParseInform,
}

var CwmpMessageGenerators = map[string]xml.MessageGenerator{
	"Fault":                 GenerateFault,
	"GetRPCMethodsResponse": GenerateGetRPCMethodsResponse,
	"InformResponse":        GenerateInformResponse,
}
