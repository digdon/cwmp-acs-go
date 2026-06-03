package cwmp_1_4

import "cwmp-acs/internal/cwmp"

func ParseGetRPCMethods(elem SOAPElement, cpeHeader cwmp.CwmpHeader) (cwmp.CwmpMessageInterface, error) {
	getRPCMethods := cwmp.GetRPCMethods{
		CwmpMessage: cwmp.CwmpMessage{
			Name: "GetRPCMethods", CwmpHeader: cpeHeader,
		},
	}

	return &getRPCMethods, nil
}
