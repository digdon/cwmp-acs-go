package db

type PendingMessage struct {
	MessageID   string
	MessageName string
	Parameters  []byte
	SendCount   int
}

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

// Table definitions
const createDeviceTableSQL = `CREATE TABLE IF NOT EXISTS device (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL UNIQUE,
		manufacturer TEXT NOT NULL,
		oui TEXT NOT NULL,
		product_class TEXT NOT NULL,
		serial_number TEXT NOT NULL,
		connection_request_url TEXT NOT NULL,
		parameter_key TEXT,
		provisioning_code TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

const createRPCQueueTableSQL = `CREATE TABLE IF NOT EXISTS rpc_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message_id TEXT NOT NULL,
		device_id TEXT NOT NULL,
		message_name TEXT NOT NULL,
		parameters TEXT,
		send_count INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
