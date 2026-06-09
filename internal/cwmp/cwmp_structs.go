package cwmp

type CwmpHeader struct {
	ID                    string
	HoldRequests          string // deprecated - only comes from ACS
	SessionTimeout        string // deprecated in CWMP 1.4 - only comes from CPEs, only in Informs
	SupportedCWMPVersions string // only comes from CPEs, only in Informs
	UseCWMPVersion        string // only comes from ACS
	// Unknown               []cwmpxml.SOAPElement
}

type CwmpMessage struct {
	Name string
	CwmpHeader
}

type CwmpMessageInterface interface {
	GetName() string
	GetID() string
}

func (c *CwmpMessage) GetName() string {
	return c.Name
}

func (c *CwmpMessage) GetID() string {
	return c.ID
}

type FaultSource int

const (
	FaultSourceCPE FaultSource = iota
	FaultSourceServer
)

func (f FaultSource) String() string {
	switch f {
	case FaultSourceCPE:
		return "Client"
	case FaultSourceServer:
		return "Server"
	default:
		return "Unknown"
	}
}

// Message substructures

type DeviceId struct {
	Manufacturer string
	OUI          string
	ProductClass string
	SerialNumber string
}

type Event struct {
	EventCode  string
	CommandKey string
}

type ParameterValueStruct struct {
	Name  string
	Value string
	Type  string
}

// Full message structures

type Fault struct {
	CwmpMessage
	Source      FaultSource
	FaultCode   int
	FaultString string
}

type GetRPCMethods struct {
	CwmpMessage
}

type GetRPCMethodsResponse struct {
	CwmpMessage
	MethodList []string
}

type Inform struct {
	CwmpMessage
	DeviceId     DeviceId
	Events       []Event
	MaxEnvelopes int
	CurrentTime  string
	RetryCount   int
	Parameters   []ParameterValueStruct
}

type InformResponse struct {
	CwmpMessage
	MaxEnvelopes int
}
