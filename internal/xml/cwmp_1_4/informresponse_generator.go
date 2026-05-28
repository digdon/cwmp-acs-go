package cwmp_1_4

import (
	"cwmp-acs/internal/cwmp"
	"fmt"
	"strconv"
)

func GenerateInformResponse(message cwmp.CwmpMessageInterface) (string, error) {
	informResponse, ok := message.(*cwmp.InformResponse)
	if !ok {
		return "", fmt.Errorf("invalid message type")
	}

	cwmpNamespace := "urn:dslforum-org:cwmp-" + namespaceVersionSuffix(informResponse.CwmpHeader.UseCWMPVersion)

	responseXML := `<soap_env:Envelope
	xmlns:soap_env="http://schemas.xmlsoap.org/soap/envelope/"
	xmlns:soap_enc="http://schemas.xmlsoap.org/soap/encoding/"
	xmlns:xsd="http://www.w3.org/2001/XMLSchema"
	xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xmlns:cwmp="` + cwmpNamespace + `">
	<soap_env:Header>
		<cwmp:ID soap_env:mustUnderstand="1">` + informResponse.GetID() + `</cwmp:ID>
	</soap_env:Header>
	<soap_env:Body>
		<cwmp:InformResponse>
			<MaxEnvelopes>` + strconv.Itoa(informResponse.MaxEnvelopes) + `</MaxEnvelopes>
		</cwmp:InformResponse>
	</soap_env:Body>
</soap_env:Envelope>`

	return responseXML, nil
}

func namespaceVersionSuffix(version string) string {
	if version == "" || version == "unknown" {
		return "1-2"
	}

	if len(version) == 3 && version[1] == '.' {
		return string(version[0]) + "-" + string(version[2])
	}

	return "1-2"
}
