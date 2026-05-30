package cwmp_1_4

import (
	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/xml"
	"cwmp-acs/internal/xml/dom"
	"fmt"
	"strconv"
)

func GenerateFault(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	fault, ok := message.(*cwmp.Fault)
	if !ok {
		return "", fmt.Errorf("invalid message type")
	}

	soapenv := namespaces[xml.SOAPENV]
	soapenc := namespaces[xml.SOAPENC]
	xsd := namespaces[xml.XSD]
	xsi := namespaces[xml.XSI]
	cwmpNS := namespaces[xml.CWMP]

	envelope := dom.NewElement(soapenv.Prefix+":Envelope").
		AddAttr("xmlns:"+soapenv.Prefix, soapenv.URL).
		AddAttr("xmlns:"+soapenc.Prefix, soapenc.URL).
		AddAttr("xmlns:"+xsd.Prefix, xsd.URL).
		AddAttr("xmlns:"+xsi.Prefix, xsi.URL).
		AddAttr("xmlns:"+cwmpNS.Prefix, cwmpNS.URL)

	header := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Header"))
	header.AddChild(
		dom.NewElement(cwmpNS.Prefix+":ID").
			AddAttr(soapenv.Prefix+":mustUnderstand", "1").
			SetText(fault.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	faultEl := body.AddChild(dom.NewElement(soapenv.Prefix + ":Fault"))
	faultEl.AddChild(dom.NewElement("faultcode").SetText(fault.Source.String()))
	faultEl.AddChild(dom.NewElement("faultstring").SetText("CWMP fault"))

	detail := faultEl.AddChild(dom.NewElement("detail"))
	cwmpFault := detail.AddChild(dom.NewElement(cwmpNS.Prefix + ":Fault"))
	cwmpFault.AddChild(dom.NewElement("FaultCode").SetText(strconv.Itoa(fault.FaultCode)))
	cwmpFault.AddChild(dom.NewElement("FaultString").SetText(fault.FaultString))

	return dom.NewDocument(envelope).Serialize()
}
