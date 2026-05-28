package xml

import (
	"cwmp-acs/internal/cwmp"
	"encoding/xml"
)

type SOAPElement struct {
	Name     xml.Name
	Attrs    []xml.Attr
	Text     string
	Children []SOAPElement
}

type ParsedEnvelope struct {
	Envelope   SOAPElement
	Header     SOAPElement
	Body       SOAPElement
	Namespaces map[string]string
}

type MessageParser func(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error)

type MessageGenerator func(message cwmp.CwmpMessageInterface) (string, error)
