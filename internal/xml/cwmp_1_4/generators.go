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

func GenerateAddObject(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	addObject, ok := message.(*cwmp.AddObject)
	if !ok {
		return "", fmt.Errorf("message is not of type AddObject")
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
			SetText(addObject.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	ao := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":AddObject"))
	ao.AddChild(dom.NewElement("ObjectName")).SetText(addObject.ObjectName)
	ao.AddChild(dom.NewElement("ParameterKey")).SetText(addObject.ParameterKey)

	return dom.NewDocument(envelope).Serialize()
}

func GenerateDeleteObject(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	deleteObject, ok := message.(*cwmp.DeleteObject)
	if !ok {
		return "", fmt.Errorf("message is not of type DeleteObject")
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
			SetText(deleteObject.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	do := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":DeleteObject"))
	do.AddChild(dom.NewElement("ObjectName")).SetText(deleteObject.ObjectName)
	do.AddChild(dom.NewElement("ParameterKey")).SetText(deleteObject.ParameterKey)

	return dom.NewDocument(envelope).Serialize()
}

func GenerateFault(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	fault, ok := message.(*cwmp.Fault)
	if !ok {
		return "", fmt.Errorf("message is not of type Fault")
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

func GenerateGetParameterAttributes(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getParameterAttributes, ok := message.(*cwmp.GetParameterAttributes)
	if !ok {
		return "", fmt.Errorf("message is not of type GetParameterAttributes")
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
			SetText(getParameterAttributes.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	gpa := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":GetParameterAttributes"))
	paramList := gpa.AddChild(dom.NewElement("ParameterNames"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(getParameterAttributes.ParameterNames)))
	for _, name := range getParameterAttributes.ParameterNames {
		paramList.AddChild(dom.NewElement("string")).SetText(name)
	}

	return dom.NewDocument(envelope).Serialize()
}

func GenerateGetParameterNames(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getParameterNames, ok := message.(*cwmp.GetParameterNames)
	if !ok {
		return "", fmt.Errorf("message is not of type GetParameterNames")
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
			SetText(getParameterNames.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	gpn := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":GetParameterNames"))
	gpn.AddChild(dom.NewElement("ParameterPath")).SetText(getParameterNames.ParameterPath)
	gpn.AddChild(dom.NewElement("NextLevel")).SetText(strconv.FormatBool(getParameterNames.NextLevel))

	return dom.NewDocument(envelope).Serialize()
}

func GenerateGetParameterValues(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getParameterValues, ok := message.(*cwmp.GetParameterValues)
	if !ok {
		return "", fmt.Errorf("message is not of type GetParameterValues")
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
			SetText(getParameterValues.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	gpv := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":GetParameterValues"))
	paramList := gpv.AddChild(dom.NewElement("ParameterNames"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(getParameterValues.ParameterNames)))
	for _, name := range getParameterValues.ParameterNames {
		paramList.AddChild(dom.NewElement("string")).SetText(name)
	}

	return dom.NewDocument(envelope).Serialize()
}

func GenerateGetRPCMethods(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getRPCMethods, ok := message.(*cwmp.GetRPCMethods)
	if !ok {
		return "", fmt.Errorf("message is not of type GetRPCMethods")
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
			SetText(getRPCMethods.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	body.AddChild(dom.NewElement(cwmpNS.Prefix + ":GetRPCMethods"))

	return dom.NewDocument(envelope).Serialize()
}

func GenerateGetRPCMethodsResponse(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	getRPCMethodsResponse, ok := message.(*cwmp.GetRPCMethodsResponse)
	if !ok {
		return "", fmt.Errorf("message is not of type GetRPCMethodsResponse")
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
	methodList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:string[%d]", xsd.Prefix, len(getRPCMethodsResponse.MethodList)))
	for _, method := range getRPCMethodsResponse.MethodList {
		methodList.AddChild(dom.NewElement("string")).SetText(method)
	}
	sort.Strings(getRPCMethodsResponse.MethodList)

	return dom.NewDocument(envelope).Serialize()
}

func GenerateInformResponse(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	informResponse, ok := message.(*cwmp.InformResponse)
	if !ok {
		return "", fmt.Errorf("message is not of type InformResponse")
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
			SetText(informResponse.GetID()),
	)
	if v := informResponse.CwmpHeader.UseCWMPVersion; v != "" {
		header.AddChild(
			dom.NewElement(cwmpNS.Prefix+":UseCWMPVersion").
				AddAttr(soapenv.Prefix+":mustUnderstand", "1").
				SetText(v),
		)
	}

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	ir := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":InformResponse"))
	ir.AddChild(dom.NewElement("MaxEnvelopes").SetText(strconv.Itoa(informResponse.MaxEnvelopes)))

	return dom.NewDocument(envelope).Serialize()
}

func GenerateReboot(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	reboot, ok := message.(*cwmp.Reboot)
	if !ok {
		return "", fmt.Errorf("message is not of type Reboot")
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
			SetText(reboot.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	rebootElem := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":Reboot"))
	rebootElem.AddChild(dom.NewElement("CommandKey").SetText(reboot.CommandKey))

	return dom.NewDocument(envelope).Serialize()
}

func GenerateSetParameterAttributes(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	setParameterAttributes, ok := message.(*cwmp.SetParameterAttributes)
	if !ok {
		return "", fmt.Errorf("message is not of type SetParameterAttributes")
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
			SetText(setParameterAttributes.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	spa := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":SetParameterAttributes"))
	paramList := spa.AddChild(dom.NewElement("ParameterList"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:SetParameterAttributeStruct[%d]", cwmpNS.Prefix, len(setParameterAttributes.ParameterList)))
	for _, param := range setParameterAttributes.ParameterList {
		paramEl := paramList.AddChild(dom.NewElement("SetParameterAttributeStruct"))
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

	return dom.NewDocument(envelope).Serialize()
}

func GenerateSetParameterValues(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	setParameterValues, ok := message.(*cwmp.SetParameterValues)
	if !ok {
		return "", fmt.Errorf("message is not of type SetParameterValues")
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
			SetText(setParameterValues.GetID()),
	)

	body := envelope.AddChild(dom.NewElement(soapenv.Prefix + ":Body"))
	spv := body.AddChild(dom.NewElement(cwmpNS.Prefix + ":SetParameterValues"))
	paramList := spv.AddChild(dom.NewElement("ParameterList"))
	paramList.AddAttr(soapenc.Prefix+":arrayType", fmt.Sprintf("%s:ParameterValueStruct[%d]", cwmpNS.Prefix, len(setParameterValues.ParameterList)))
	for _, param := range setParameterValues.ParameterList {
		paramEl := paramList.AddChild(dom.NewElement("ParameterValueStruct"))
		paramEl.AddChild(dom.NewElement("Name").SetText(param.Name))
		paramEl.AddChild(dom.NewElement("Value").SetText(param.Value)).AddAttr(xsi.Prefix+":type", fmt.Sprintf("%s:%s", xsd.Prefix, param.Type))
	}
	spv.AddChild(dom.NewElement("ParameterKey").SetText(setParameterValues.ParameterKey))

	return dom.NewDocument(envelope).Serialize()
}
