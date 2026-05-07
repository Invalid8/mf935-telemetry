package state

import "time"

type DeviceState struct {
	// Network
	PPPStatus       string `json:"ppp_status"`
	NetworkType     string `json:"network_type"`
	NetworkProvider string `json:"network_provider"`
	WanIP           string `json:"wan_ipaddr"`
	SimcardRoam     string `json:"simcard_roam"`
	DialMode        string `json:"dial_mode"`
	ModemMainState  string `json:"modem_main_state"`

	// Signal
	SignalBars string `json:"signalbar"`
	RSSI       string `json:"rssi"`
	LTERsrp    string `json:"lte_rsrp"`
	RSCP       string `json:"rscp"`

	// Battery
	BatteryPct      string `json:"battery_vol_percent"`
	BatteryCharging string `json:"battery_charging"`
	BatteryPers     string `json:"battery_pers"`
	BatteryValue    string `json:"battery_value"`

	// WiFi
	SSID1      string `json:"SSID1"`
	WifiState  string `json:"wifi_cur_state"`
	StaCount   string `json:"sta_count"`
	MStaCount  string `json:"m_sta_count"`
	StationMAC string `json:"station_mac"`

	// Throughput
	RxThrpt      string `json:"realtime_rx_thrpt"`
	TxThrpt      string `json:"realtime_tx_thrpt"`
	RxBytes      string `json:"realtime_rx_bytes"`
	TxBytes      string `json:"realtime_tx_bytes"`
	RealtimeTime string `json:"realtime_time"`

	// Monthly
	MonthlyRxBytes string `json:"monthly_rx_bytes"`
	MonthlyTxBytes string `json:"monthly_tx_bytes"`
	DateMonth      string `json:"date_month"`

	// SMS
	SmsUnreadNum    string `json:"sms_unread_num"`
	SmsDbChange     string `json:"sms_db_change"`
	SmsReceivedFlag string `json:"sms_received_flag"`

	// Data limit
	DataLimitSwitch string `json:"data_volume_limit_switch"`
	DataLimitSize   string `json:"data_volume_limit_size"`
	DataLimitUnit   string `json:"data_volume_limit_unit"`
	DataAlertPct    string `json:"data_volume_alert_percent"`

	// Carrier aggregation
	WanLteCa string `json:"wan_lte_ca"`

	// OTA
	NewVersionState     string `json:"new_version_state"`
	CurrentUpgradeState string `json:"current_upgrade_state"`

	// Session
	Loginfo string `json:"loginfo"`

	// Set by the poller after each successful fetch — not from the device
	LastUpdated time.Time `json:"last_updated"`
}
