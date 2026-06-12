package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var ErrDeviceNotFound = errors.New("device not found")

var db *sql.DB

func InitDB(ctx context.Context) {
	log.Println("Initializing database connection")
	// Open connection to DB (in this case, open a local SQLite database file, creating it if it doesn't exist)
	var err error
	db, err = sql.Open("sqlite3", "./cwmp_acs.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	// Verify that the DB can be reached
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Set up database for concurrent access (SQLite with WAL allows concurrent reads and writes, but we should still limit connections)
	var journalMode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode=WAL;").Scan(&journalMode); err != nil || journalMode != "wal" {
		log.Fatalf("Failed to enable WAL mode: mode=%s err=%v", journalMode, err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout=5000;"); err != nil {
		log.Fatalf("Failed to set busy_timeout: %v", err)
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	// Set up required tables
	if _, err := db.ExecContext(ctx, createDeviceTableSQL); err != nil {
		log.Fatalf("Create device table failed: %v", err)
	}

	if _, err := db.ExecContext(ctx, createRPCQueueTableSQL); err != nil {
		log.Fatalf("Create rpc_queue table failed: %v", err)
	}
}

// nullableString converts an empty string to a NULL sql value.
func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func CloseDB() {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Println("Error closing database:", err)
		}
	}
}

func AddDevice(ctx context.Context, info DeviceInfo) error {
	log.Printf("Registering device: %s", info.DeviceID)
	upsertSQL := `INSERT INTO device (device_id, manufacturer, oui, product_class, serial_number, connection_request_url, parameter_key, provisioning_code)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(device_id) DO UPDATE SET
			manufacturer = excluded.manufacturer,
			oui = excluded.oui,
			product_class = excluded.product_class,
			serial_number = excluded.serial_number,
			connection_request_url = excluded.connection_request_url,
			parameter_key = excluded.parameter_key,
			provisioning_code = excluded.provisioning_code,
			updated_at = CURRENT_TIMESTAMP;`
	result, err := db.ExecContext(ctx, upsertSQL, info.DeviceID, info.Manufacturer, info.OUI, info.ProductClass, info.SerialNumber, info.ConnectionRequestURL, nullableString(info.ParameterKey), nullableString(info.ProvisioningCode))
	if err != nil {
		log.Printf("Error registering device %s: %v", info.DeviceID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for device %s: %v", info.DeviceID, err)
		return err
	}
	fmt.Printf("Device %s registered successfully (rows affected: %d)\n", info.DeviceID, rowsAffected)
	return nil
}

func UpdateDevice(ctx context.Context, info DeviceInfo) error {
	log.Printf("Updating device: %s", info.DeviceID)
	updateSQL := `UPDATE device SET manufacturer = ?, oui = ?, product_class = ?, serial_number = ?, connection_request_url = ?, parameter_key = ?, provisioning_code = ?, updated_at = CURRENT_TIMESTAMP WHERE device_id = ?;`
	result, err := db.ExecContext(ctx, updateSQL, info.Manufacturer, info.OUI, info.ProductClass, info.SerialNumber, info.ConnectionRequestURL, nullableString(info.ParameterKey), nullableString(info.ProvisioningCode), info.DeviceID)
	if err != nil {
		log.Printf("Error updating device %s: %v", info.DeviceID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for device %s: %v", info.DeviceID, err)
		return err
	}
	fmt.Printf("Device %s updated successfully (rows affected: %d)\n", info.DeviceID, rowsAffected)
	return nil
}

func GetDeviceByID(ctx context.Context, deviceID string) (info *DeviceInfo, err error) {
	log.Printf("Retrieving device: %s", deviceID)
	querySQL := `SELECT manufacturer, oui, product_class, serial_number, connection_request_url, parameter_key, provisioning_code FROM device WHERE device_id = ?;`
	row := db.QueryRowContext(ctx, querySQL, deviceID)

	var manufacturer, oui, productClass, serialNumber, connectionRequestURL string
	var parameterKey, provisioningCode sql.NullString
	if err := row.Scan(&manufacturer, &oui, &productClass, &serialNumber, &connectionRequestURL, &parameterKey, &provisioningCode); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Device %s not found", deviceID)
			return nil, ErrDeviceNotFound
		}
		log.Printf("Error retrieving device %s: %v", deviceID, err)
		return nil, err
	}

	return &DeviceInfo{
		DeviceID:             deviceID,
		Manufacturer:         manufacturer,
		OUI:                  oui,
		ProductClass:         productClass,
		SerialNumber:         serialNumber,
		ConnectionRequestURL: connectionRequestURL,
		ParameterKey:         parameterKey.String,
		ProvisioningCode:     provisioningCode.String,
	}, nil
}

func IsDeviceRegistered(ctx context.Context, deviceID string) (bool, error) {
	log.Printf("Checking if device is registered: %s", deviceID)
	querySQL := `SELECT COUNT(1) FROM device WHERE device_id = ?;`
	row := db.QueryRowContext(ctx, querySQL, deviceID)

	var count int
	if err := row.Scan(&count); err != nil {
		log.Printf("Error checking registration for device %s: %v", deviceID, err)
		return false, err
	}
	return count > 0, nil
}

func QueueRPC(ctx context.Context, deviceID, messageID, messageName string, parameters []byte) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO rpc_queue (message_id, device_id, message_name, parameters) VALUES (?, ?, ?, ?);`,
		messageID, deviceID, messageName, nullableString(string(parameters)),
	)
	if err != nil {
		log.Printf("Error enqueueing message %s for device %s: %v", messageName, deviceID, err)
	}
	return err
}

func FetchQueuedRPC(ctx context.Context, deviceID string) (*PendingMessage, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var rowID int64
	var messageID, messageName string
	var parameters sql.NullString
	var sendCount int

	row := tx.QueryRowContext(ctx,
		`SELECT id, message_id, message_name, parameters, send_count FROM rpc_queue WHERE device_id = ? ORDER BY id ASC LIMIT 1;`,
		deviceID,
	)
	if err := row.Scan(&rowID, &messageID, &messageName, &parameters, &sendCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE rpc_queue SET send_count = send_count + 1 WHERE id = ?;`, rowID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &PendingMessage{
		MessageID:   messageID,
		MessageName: messageName,
		Parameters:  []byte(parameters.String),
		SendCount:   sendCount,
	}, nil
}

func DeleteQueuedRPC(ctx context.Context, messageID string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM rpc_queue WHERE message_id = ?;`, messageID)
	return err
}
