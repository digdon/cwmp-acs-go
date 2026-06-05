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

type ParameterAttributeStruct struct {
	Name         string
	Notification int
	AccessList   []string
}

type ParameterInfoStruct struct {
	Name     string
	Writable bool
}

type ParameterValueStruct struct {
	Name  string
	Value string
	Type  string
}

type SetParameterAttributeStruct struct {
	Name               string
	NotificationChange bool
	Notification       int
	AccessListChange   bool
	AccessList         []string
}

// Full message structures

type Fault struct {
	CwmpMessage
	Source      FaultSource
	FaultCode   int
	FaultString string
}

type GetParameterNames struct {
	CwmpMessage
	ParameterPath string
	NextLevel     bool
}

type GetParameterNamesResponse struct {
	CwmpMessage
	ParameterList []ParameterInfoStruct
}

type GetParameterValues struct {
	CwmpMessage
	ParameterNames []string
}

type GetParameterValuesResponse struct {
	CwmpMessage
	ParameterList []ParameterValueStruct
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

type Reboot struct {
	CwmpMessage
	CommandKey string
}

type RebootResponse struct {
	CwmpMessage
}

type SetParameterAttributes struct {
	CwmpMessage
	ParameterList []SetParameterAttributeStruct
}

type SetParameterAttributesResponse struct {
	CwmpMessage
}

type SetParameterValues struct {
	CwmpMessage
	ParameterList []ParameterValueStruct
	ParameterKey  string
}

type SetParameterValuesResponse struct {
	CwmpMessage
	Status int
}

type TransferComplete struct {
	CwmpMessage
	CommandKey string
}
