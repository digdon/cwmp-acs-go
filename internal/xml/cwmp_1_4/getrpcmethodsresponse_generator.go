package cwmp_1_4

import (
	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/xml"
	"cwmp-acs/internal/xml/dom"
	"fmt"
	"sort"
)

func GenerateGetRPCMethodsResponse(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getRPCMethodsResponse, ok := message.(*cwmp.GetRPCMethodsResponse)
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
			SetText(getRPCMethodsResponse.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	grmr := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":GetRPCMethodsResponse"))
	methodList := grmr.AddChild(dom.NewElement("MethodList"))
	methodList.AddAttr(soapenv.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(getRPCMethodsResponse.MethodList)))
	for _, method := range getRPCMethodsResponse.MethodList {
		methodList.AddChild(dom.NewElement("string")).SetText(method)
	}
	sort.Strings(getRPCMethodsResponse.MethodList)

	return dom.NewDocument(envelope).Serialize()
}
