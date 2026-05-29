package cwmp_1_4

import (
	"bytes"
	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/xml"
	stdxml "encoding/xml"
	"fmt"
)

func GenerateInformResponse(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	informResponse, ok := message.(*cwmp.InformResponse)
	if !ok {
		return "", fmt.Errorf("invalid message type")
	}

	var buf bytes.Buffer
	encoder := stdxml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	envelopeStart := stdxml.StartElement{
		Name: stdxml.Name{Local: namespaces[xml.SOAPENV].Prefix + ":Envelope"},
		Attr: []stdxml.Attr{
			{Name: stdxml.Name{Local: "xmlns:" + namespaces[xml.SOAPENV].Prefix}, Value: namespaces[xml.SOAPENV].URL},
			{Name: stdxml.Name{Local: "xmlns:" + namespaces[xml.SOAPENC].Prefix}, Value: namespaces[xml.SOAPENC].URL},
			{Name: stdxml.Name{Local: "xmlns:" + namespaces[xml.XSD].Prefix}, Value: namespaces[xml.XSD].URL},
			{Name: stdxml.Name{Local: "xmlns:" + namespaces[xml.XSI].Prefix}, Value: namespaces[xml.XSI].URL},
			{Name: stdxml.Name{Local: "xmlns:" + namespaces[xml.CWMP].Prefix}, Value: namespaces[xml.CWMP].URL},
		},
	}

	if err := encoder.EncodeToken(envelopeStart); err != nil {
		return "", err
	}

	headerStart := stdxml.StartElement{Name: stdxml.Name{Local: namespaces[xml.SOAPENV].Prefix + ":Header"}}
	if err := encoder.EncodeToken(headerStart); err != nil {
		return "", err
	}

	idStart := stdxml.StartElement{
		Name: stdxml.Name{Local: namespaces[xml.CWMP].Prefix + ":ID"},
		Attr: []stdxml.Attr{{Name: stdxml.Name{Local: namespaces[xml.SOAPENV].Prefix + ":mustUnderstand"}, Value: "1"}},
	}
	if err := encoder.EncodeElement(informResponse.GetID(), idStart); err != nil {
		return "", err
	}

	useCWMPVersionStart := stdxml.StartElement{
		Name: stdxml.Name{Local: namespaces[xml.CWMP].Prefix + ":UseCWMPVersion"},
		Attr: []stdxml.Attr{{Name: stdxml.Name{Local: namespaces[xml.SOAPENV].Prefix + ":mustUnderstand"}, Value: "1"}},
	}
	if err := encoder.EncodeElement(informResponse.CwmpHeader.UseCWMPVersion, useCWMPVersionStart); err != nil {
		return "", err
	}

	if err := encoder.EncodeToken(headerStart.End()); err != nil {
		return "", err
	}

	bodyStart := stdxml.StartElement{Name: stdxml.Name{Local: namespaces[xml.SOAPENV].Prefix + ":Body"}}
	if err := encoder.EncodeToken(bodyStart); err != nil {
		return "", err
	}

	informResponseStart := stdxml.StartElement{Name: stdxml.Name{Local: namespaces[xml.CWMP].Prefix + ":InformResponse"}}
	if err := encoder.EncodeToken(informResponseStart); err != nil {
		return "", err
	}

	maxEnvelopesStart := stdxml.StartElement{Name: stdxml.Name{Local: "MaxEnvelopes"}}
	if err := encoder.EncodeElement(informResponse.MaxEnvelopes, maxEnvelopesStart); err != nil {
		return "", err
	}

	if err := encoder.EncodeToken(informResponseStart.End()); err != nil {
		return "", err
	}

	if err := encoder.EncodeToken(bodyStart.End()); err != nil {
		return "", err
	}

	if err := encoder.EncodeToken(envelopeStart.End()); err != nil {
		return "", err
	}

	if err := encoder.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
}
