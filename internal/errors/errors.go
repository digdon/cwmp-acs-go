package errors

import (
	"fmt"

	"cwmp-acs/internal/cwmp"
)

type IncomingMessageError struct {
	Header      cwmp.CwmpHeader
	Source      cwmp.FaultSource
	FaultCode   int
	FaultString string
}

func (e *IncomingMessageError) Error() string {
	return fmt.Sprintf("IncomingMessageError: %s - %d: %s", e.Source.String(), e.FaultCode, e.FaultString)
}

type XmlParsingError struct {
	Message string
}

func (e *XmlParsingError) Error() string {
	return fmt.Sprintf("XmlParsingError: %s", e.Message)
}
