package cwmp_1_4

import (
	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/xml"
	"cwmp-acs/internal/xml/dom"
	"fmt"
	"sort"
	"strconv"
)

var CwmpMessageGenerators = map[string]xml.MessageGenerator{
	"AddObject":    GenerateAddObject,
	"DeleteObject": GenerateDeleteObject,
	// "Download":                 GenerateDownload,
	"Fault":                  GenerateFault,
	"GetParameterAttributes": GenerateGetParameterAttributes,
	"GetParameterNames":      GenerateGetParameterNames,
	"GetParameterValues":     GenerateGetParameterValues,
	"GetRPCMethods":          GenerateGetRPCMethods,
	"GetRPCMethodsResponse":  GenerateGetRPCMethodsResponse,
	"InformResponse":         GenerateInformResponse,
	"Reboot":                 GenerateReboot,
	"SetParameterAttributes": GenerateSetParameterAttributes,
	"SetParameterValues":     GenerateSetParameterValues,
	// "TransferCompleteResponse": GenerateTransferCompleteResponse,
}

type messageParts struct {
	Envelope *dom.Element
	Header   *dom.Element
	Body     *dom.Element
	RPC      *dom.Element
}

func generateMessageParts(namespaces map[xml.NamespaceID]xml.Namespace, message cwmp.CwmpMessageInterface) *messageParts {
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
			SetText(message.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	rpc := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":" + message.GetName()))

	return &messageParts{
		Envelope: envelope,
		Header:   header,
		Body:     body,
		RPC:      rpc,
	}
}

func GenerateAddObject(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	addObject, ok := message.(*cwmp.AddObject)
	if !ok {
		return "", fmt.Errorf("message is not of type AddObject")
	}

	messageParts := generateMessageParts(namespaces, addObject)
	ao := messageParts.RPC
	ao.AddChild(dom.NewElement("ObjectName")).SetText(addObject.ObjectName)
	ao.AddChild(dom.NewElement("ParameterKey")).SetText(addObject.ParameterKey)

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateDeleteObject(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	deleteObject, ok := message.(*cwmp.DeleteObject)
	if !ok {
		return "", fmt.Errorf("message is not of type DeleteObject")
	}

	messageParts := generateMessageParts(namespaces, deleteObject)
	do := messageParts.RPC
	do.AddChild(dom.NewElement("ObjectName")).SetText(deleteObject.ObjectName)
	do.AddChild(dom.NewElement("ParameterKey")).SetText(deleteObject.ParameterKey)

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateFault(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	fault, ok := message.(*cwmp.Fault)
	if !ok {
		return "", fmt.Errorf("message is not of type Fault")
	}

	cwmpNS := namespaces[xml.CWMP]

	messageParts := generateMessageParts(namespaces, fault)
	faultEl := messageParts.RPC
	faultEl.AddChild(dom.NewElement("faultcode").SetText(fault.Source.String()))
	faultEl.AddChild(dom.NewElement("faultstring").SetText("CWMP fault"))

	detail := faultEl.AddChild(dom.NewElement("detail"))
	cwmpFault := detail.AddChild(dom.NewElement(cwmpNS.Prefix + ":Fault"))
	cwmpFault.AddChild(dom.NewElement("FaultCode").SetText(strconv.Itoa(fault.FaultCode)))
	cwmpFault.AddChild(dom.NewElement("FaultString").SetText(fault.FaultString))

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateGetParameterAttributes(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getParameterAttributes, ok := message.(*cwmp.GetParameterAttributes)
	if !ok {
		return "", fmt.Errorf("message is not of type GetParameterAttributes")
	}

	soapenc := namespaces[xml.SOAPENC]
	xsd := namespaces[xml.XSD]

	messageParts := generateMessageParts(namespaces, getParameterAttributes)
	gpa := messageParts.RPC
	paramList := gpa.AddChild(dom.NewElement("ParameterNames"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(getParameterAttributes.ParameterNames)))
	for _, name := range getParameterAttributes.ParameterNames {
		paramList.AddChild(dom.NewElement("string")).SetText(name)
	}

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateGetParameterNames(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getParameterNames, ok := message.(*cwmp.GetParameterNames)
	if !ok {
		return "", fmt.Errorf("message is not of type GetParameterNames")
	}

	messageParts := generateMessageParts(namespaces, getParameterNames)
	gpn := messageParts.RPC
	gpn.AddChild(dom.NewElement("ParameterPath")).SetText(getParameterNames.ParameterPath)
	gpn.AddChild(dom.NewElement("NextLevel")).SetText(strconv.FormatBool(getParameterNames.NextLevel))

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateGetParameterValues(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getParameterValues, ok := message.(*cwmp.GetParameterValues)
	if !ok {
		return "", fmt.Errorf("message is not of type GetParameterValues")
	}

	soapenc := namespaces[xml.SOAPENC]
	xsd := namespaces[xml.XSD]

	messageParts := generateMessageParts(namespaces, getParameterValues)
	gpv := messageParts.RPC
	paramList := gpv.AddChild(dom.NewElement("ParameterNames"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(getParameterValues.ParameterNames)))
	for _, name := range getParameterValues.ParameterNames {
		paramList.AddChild(dom.NewElement("string")).SetText(name)
	}

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateGetRPCMethods(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getRPCMethods, ok := message.(*cwmp.GetRPCMethods)
	if !ok {
		return "", fmt.Errorf("message is not of type GetRPCMethods")
	}

	messageParts := generateMessageParts(namespaces, getRPCMethods)

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateGetRPCMethodsResponse(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getRPCMethodsResponse, ok := message.(*cwmp.GetRPCMethodsResponse)
	if !ok {
		return "", fmt.Errorf("message is not of type GetRPCMethodsResponse")
	}

	soapenc := namespaces[xml.SOAPENC]
	xsd := namespaces[xml.XSD]

	messageParts := generateMessageParts(namespaces, getRPCMethodsResponse)
	grmr := messageParts.RPC
	methodList := grmr.AddChild(dom.NewElement("MethodList"))
	methodList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(getRPCMethodsResponse.MethodList)))
	for _, method := range getRPCMethodsResponse.MethodList {
		methodList.AddChild(dom.NewElement("string")).SetText(method)
	}
	sort.Strings(getRPCMethodsResponse.MethodList)

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateInformResponse(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	informResponse, ok := message.(*cwmp.InformResponse)
	if !ok {
		return "", fmt.Errorf("message is not of type InformResponse")
	}

	soapenv := namespaces[xml.SOAPENV]
	cwmpNS := namespaces[xml.CWMP]

	messageParts := generateMessageParts(namespaces, informResponse)
	header := messageParts.Header
	if v := informResponse.CwmpHeader.UseCWMPVersion; v != "" {
		header.AddChild(
			dom.NewElement(cwmpNS.Prefix+":UseCWMPVersion").
				AddAttr(soapenv.Prefix+":mustUnderstand", "1").
				SetText(v),
		)
	}

	ir := messageParts.RPC
	ir.AddChild(dom.NewElement("MaxEnvelopes").SetText(strconv.Itoa(informResponse.MaxEnvelopes)))

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateReboot(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	reboot, ok := message.(*cwmp.Reboot)
	if !ok {
		return "", fmt.Errorf("message is not of type Reboot")
	}

	messageParts := generateMessageParts(namespaces, reboot)
	rebootElem := messageParts.RPC
	rebootElem.AddChild(dom.NewElement("CommandKey").SetText(reboot.CommandKey))

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateSetParameterAttributes(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	setParameterAttributes, ok := message.(*cwmp.SetParameterAttributes)
	if !ok {
		return "", fmt.Errorf("message is not of type SetParameterAttributes")
	}

	soapenc := namespaces[xml.SOAPENC]
	xsd := namespaces[xml.XSD]
	cwmpNS := namespaces[xml.CWMP]

	messageParts := generateMessageParts(namespaces, setParameterAttributes)
	spa := messageParts.RPC
	paramList := spa.AddChild(dom.NewElement("ParameterList"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:SetParameterAttributesStruct[%d]", cwmpNS.Prefix, len(setParameterAttributes.ParameterList)))
	for _, param := range setParameterAttributes.ParameterList {
		paramEl := paramList.AddChild(dom.NewElement("SetParameterAttributesStruct"))
		paramEl.AddChild(dom.NewElement("Name").SetText(param.Name))
		paramEl.AddChild(dom.NewElement("NotificationChange").SetText(strconv.FormatBool(param.NotificationChange)))
		paramEl.AddChild(dom.NewElement("Notification").SetText(strconv.Itoa(param.Notification)))
		paramEl.AddChild(dom.NewElement("AccessListChange").SetText(strconv.FormatBool(param.AccessListChange)))
		accessListEl := paramEl.AddChild(dom.NewElement("AccessList"))
		accessListEl.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(param.AccessList)))
		for _, access := range param.AccessList {
			accessListEl.AddChild(dom.NewElement("string").SetText(access))
		}
	}

	return dom.NewDocument(messageParts.Envelope).Serialize()
}

func GenerateSetParameterValues(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	setParameterValues, ok := message.(*cwmp.SetParameterValues)
	if !ok {
		return "", fmt.Errorf("message is not of type SetParameterValues")
	}

	soapenc := namespaces[xml.SOAPENC]
	xsd := namespaces[xml.XSD]
	xsi := namespaces[xml.XSI]
	cwmpNS := namespaces[xml.CWMP]

	messageParts := generateMessageParts(namespaces, setParameterValues)
	spv := messageParts.RPC
	paramList := spv.AddChild(dom.NewElement("ParameterList"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:ParameterValueStruct[%d]", cwmpNS.Prefix, len(setParameterValues.ParameterList)))
	for _, param := range setParameterValues.ParameterList {
		paramEl := paramList.AddChild(dom.NewElement("ParameterValueStruct"))
		paramEl.AddChild(dom.NewElement("Name").SetText(param.Name))
		paramEl.AddChild(dom.NewElement("Value").SetText(param.Value)).AddAttr(xsi.Prefix+":type", fmt.Sprintf("%s:%s", xsd.Prefix, param.Type))
	}
	spv.AddChild(dom.NewElement("ParameterKey").SetText(setParameterValues.ParameterKey))

	return dom.NewDocument(messageParts.Envelope).Serialize()
}
