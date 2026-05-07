package poller

import (
	"context"
	"log"
	"time"

	"mf935-telemetry/internal/client"
	"mf935-telemetry/internal/diff"
	"mf935-telemetry/internal/events"
	"mf935-telemetry/internal/state"
	"mf935-telemetry/internal/ws"
)

const pollInterval = 2 * time.Second

var pollFields = []string{
	// aj — always polled
	"modem_main_state", "pin_status", "opms_wan_mode", "loginfo",
	"new_version_state", "current_upgrade_state", "is_mandatory",
	"wifi_dfs_status", "battery_value", "ppp_dial_conn_fail_counter",
	// ew — status bar fields
	"signalbar", "network_type", "network_provider", "ppp_status",
	"wifi_cur_state", "SSID1", "sta_count", "m_sta_count",
	"simcard_roam", "station_mac", "battery_charging",
	"battery_vol_percent", "battery_pers", "realtime_tx_bytes",
	"realtime_rx_bytes", "realtime_time", "realtime_tx_thrpt",
	"realtime_rx_thrpt", "monthly_rx_bytes", "monthly_tx_bytes",
	"date_month", "data_volume_limit_switch", "data_volume_limit_size",
	"data_volume_alert_percent", "data_volume_limit_unit",
	"dial_mode", "wan_lte_ca", "wan_ipaddr", "rssi", "lte_rsrp", "rscp",
	// SMS
	"sms_received_flag", "sms_unread_num", "sms_db_change",
}

type Poller struct {
	client *client.RouterClient
	hub    *ws.Hub
}

func New(c *client.RouterClient, h *ws.Hub) *Poller {
	return &Poller{client: c, hub: h}
}

func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var prev state.DeviceState
	first := true

	log.Println("poller: started")

	for {
		select {
		case <-ctx.Done():
			log.Println("poller: stopped")
			return

		case <-ticker.C:
			next, err := p.fetch()
			if err != nil {
				log.Printf("poller: fetch error: %v", err)
				continue
			}

			if first {
				p.hub.StoreSnapshot(events.Event{
					Type:    events.EventSnapshot,
					Payload: next,
					At:      time.Now(),
				})
				first = false
			} else {
				for _, event := range diff.Compute(prev, next) {
					p.hub.Broadcast(event)
				}
			}

			prev = next
		}
	}
}

func (p *Poller) fetch() (state.DeviceState, error) {
	data, err := p.client.GetCmds(pollFields)
	if err != nil {
		return state.DeviceState{}, err
	}

	return state.DeviceState{
		// Network
		PPPStatus:       data["ppp_status"],
		NetworkType:     data["network_type"],
		NetworkProvider: data["network_provider"],
		WanIP:           data["wan_ipaddr"],
		SimcardRoam:     data["simcard_roam"],
		DialMode:        data["dial_mode"],
		ModemMainState:  data["modem_main_state"],
		// Signal
		SignalBars: data["signalbar"],
		RSSI:       data["rssi"],
		LTERsrp:    data["lte_rsrp"],
		RSCP:       data["rscp"],
		// Battery
		BatteryPct:      data["battery_vol_percent"],
		BatteryCharging: data["battery_charging"],
		BatteryPers:     data["battery_pers"],
		BatteryValue:    data["battery_value"],
		// WiFi
		SSID1:      data["SSID1"],
		WifiState:  data["wifi_cur_state"],
		StaCount:   data["sta_count"],
		MStaCount:  data["m_sta_count"],
		StationMAC: data["station_mac"],
		// Throughput
		RxThrpt:      data["realtime_rx_thrpt"],
		TxThrpt:      data["realtime_tx_thrpt"],
		RxBytes:      data["realtime_rx_bytes"],
		TxBytes:      data["realtime_tx_bytes"],
		RealtimeTime: data["realtime_time"],
		// Monthly
		MonthlyRxBytes: data["monthly_rx_bytes"],
		MonthlyTxBytes: data["monthly_tx_bytes"],
		DateMonth:      data["date_month"],
		// SMS
		SmsUnreadNum:    data["sms_unread_num"],
		SmsDbChange:     data["sms_db_change"],
		SmsReceivedFlag: data["sms_received_flag"],
		// Data limit
		DataLimitSwitch: data["data_volume_limit_switch"],
		DataLimitSize:   data["data_volume_limit_size"],
		DataLimitUnit:   data["data_volume_limit_unit"],
		DataAlertPct:    data["data_volume_alert_percent"],
		// Carrier aggregation
		WanLteCa: data["wan_lte_ca"],
		// OTA
		NewVersionState:     data["new_version_state"],
		CurrentUpgradeState: data["current_upgrade_state"],
		// Session
		Loginfo:     data["loginfo"],
		LastUpdated: time.Now(),
	}, nil
}
