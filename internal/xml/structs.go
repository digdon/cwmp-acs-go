package xml

import (
	"cwmp-acs/internal/cwmp"
	"encoding/xml"
)

type Namespace struct {
	Prefix string
	URL    string
}

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

type MessageGenerator func(message cwmp.CwmpMessageInterface, namespaces map[NamespaceID]Namespace) (string, error)

type NamespaceID int

const (
	SOAPENV NamespaceID = iota
	SOAPENC
	XSI
	XSD
	CWMP
)

var namespaceUrls = map[NamespaceID]string{
	SOAPENV: SOAPENV_NS_URL,
	SOAPENC: SOAPENC_NS_URL,
	XSI:     XSI_NS_URL,
	XSD:     XSD_NS_URL,
	CWMP:    CWMP_1_2_NS_URL, // Default to 1.2, but this can be overridden by the actual namespaces in the message
}

var namespaceUrlToIDMap = map[string]NamespaceID{
	SOAPENV_NS_URL:  SOAPENV,
	SOAPENC_NS_URL:  SOAPENC,
	XSI_NS_URL:      XSI,
	XSD_NS_URL:      XSD,
	CWMP_1_0_NS_URL: CWMP,
	CWMP_1_1_NS_URL: CWMP,
	CWMP_1_2_NS_URL: CWMP,
}

const SOAPENV_NS_URL string = "http://schemas.xmlsoap.org/soap/envelope/"
const SOAPENC_NS_URL string = "http://schemas.xmlsoap.org/soap/encoding/"
const XSI_NS_URL string = "http://www.w3.org/2001/XMLSchema-instance"
const XSD_NS_URL string = "http://www.w3.org/2001/XMLSchema"

const CWMP_NS_PREFIX string = "cwmp"
const CWMP_1_0_NS_URL string = "urn:dslforum-org:cwmp-1-0"
const CWMP_1_1_NS_URL string = "urn:dslforum-org:cwmp-1-1"
const CWMP_1_2_NS_URL string = "urn:dslforum-org:cwmp-1-2"
