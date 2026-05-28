package cwmp_1_4

import (
	"cwmp-acs/internal/cwmp"
	"cwmp-acs/internal/xml"
	"fmt"
	"strconv"
)

func GenerateInformResponse(message cwmp.CwmpMessageInterface, namespaces map[xml.NamespaceID]xml.Namespace) (string, error) {
	informResponse, ok := message.(*cwmp.InformResponse)
	if !ok {
		return "", fmt.Errorf("invalid message type")
	}

	responseXML := `<soap_env:Envelope
	xmlns:soap_env="http://schemas.xmlsoap.org/soap/envelope/"
	xmlns:soap_enc="http://schemas.xmlsoap.org/soap/encoding/"
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xmlns:` + namespaces[xml.CWMP].Prefix + `="` + namespaces[xml.CWMP].URL + `">
	<soap_env:Header>
		<` + namespaces[xml.CWMP].Prefix + `:ID soap_env:mustUnderstand="1">` + informResponse.GetID() + `</` + namespaces[xml.CWMP].Prefix + `:ID>
		<` + namespaces[xml.CWMP].Prefix + `:UseCWMPVersion soap_env:mustUnderstand="1">` + informResponse.CwmpHeader.UseCWMPVersion + `</` + namespaces[xml.CWMP].Prefix + `:UseCWMPVersion>
	</soap_env:Header>
	<soap_env:Body>
		<` + namespaces[xml.CWMP].Prefix + `:InformResponse>
			<` + namespaces[xml.CWMP].Prefix + `:MaxEnvelopes>` + strconv.Itoa(informResponse.MaxEnvelopes) + `</` + namespaces[xml.CWMP].Prefix + `:MaxEnvelopes>
		</` + namespaces[xml.CWMP].Prefix + `:InformResponse>
	</soap_env:Body>
</soap_env:Envelope>`

	return responseXML, nil
}
