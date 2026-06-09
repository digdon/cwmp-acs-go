package db

type DeviceInfo struct {
	DeviceID             string
	Manufacturer         string
	OUI                  string
	ProductClass         string
	SerialNumber         string
	SoftwareVersion      string
	HardwareVersion      string
	ConnectionRequestURL string
	ParameterKey         string
	ProvisioningCode     string
}
