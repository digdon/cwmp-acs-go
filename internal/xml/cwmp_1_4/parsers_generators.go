package cwmp_1_4

import (
	"cwmp-acs/internal/xml"
)

var CwmpMessageParsers = map[string]xml.MessageParser{
	"Inform": ParseInform,
}

var CwmpMessageGenerators = map[string]xml.MessageGenerator{
	"Fault":          GenerateFault,
	"InformResponse": GenerateInformResponse,
}
