package events

import "time"

type Event struct {
	Type    string      `json:"event"`
	Payload interface{} `json:"payload"`
	At      time.Time   `json:"at"`
}

const (
	EventSignalChanged    = "signal_changed"
	EventNetworkChanged   = "network_changed"
	EventConnectionGained = "connection_gained"
	EventConnectionLost   = "connection_lost"
	EventBatteryChanged   = "battery_changed"
	EventClientsChanged   = "clients_changed"
	EventSmsReceived      = "sms_received"
	EventSmsDbChanged     = "sms_db_changed"
	EventThroughput       = "throughput_updated"
	EventDataLimitChanged = "data_limit_changed"
	EventOTAAvailable     = "ota_available"
	EventSnapshot         = "snapshot"
)
