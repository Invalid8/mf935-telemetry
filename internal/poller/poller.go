package poller

import (
	"context"
	"log"
	"strings"
	"time"

	"mf935-telemetry/internal/auth"
	"mf935-telemetry/internal/client"
	"mf935-telemetry/internal/diff"
	"mf935-telemetry/internal/events"
	"mf935-telemetry/internal/state"
	"mf935-telemetry/internal/ws"
)

const (
	pollInterval  = 2 * time.Second
	retryInterval = 5 * time.Second
)

type deviceStatus int

const (
	statusOK deviceStatus = iota
	statusUnreachable
	statusMismatch
)

var pollFields = []string{
	"modem_main_state", "pin_status", "opms_wan_mode", "loginfo",
	"new_version_state", "current_upgrade_state", "is_mandatory",
	"wifi_dfs_status", "battery_value", "ppp_dial_conn_fail_counter",
	"signalbar", "network_type", "network_provider", "ppp_status",
	"wifi_cur_state", "SSID1", "sta_count", "m_sta_count",
	"simcard_roam", "station_mac", "battery_charging",
	"battery_vol_percent", "battery_pers", "realtime_tx_bytes",
	"realtime_rx_bytes", "realtime_time", "realtime_tx_thrpt",
	"realtime_rx_thrpt", "monthly_rx_bytes", "monthly_tx_bytes",
	"date_month", "data_volume_limit_switch", "data_volume_limit_size",
	"data_volume_alert_percent", "data_volume_limit_unit",
	"dial_mode", "wan_lte_ca", "wan_ipaddr", "rssi", "lte_rsrp", "rscp",
	"sms_received_flag", "sms_unread_num", "sms_db_change",
}

type Poller struct {
	client  *client.RouterClient
	session *auth.Session
	hub     *ws.Hub
}

func New(c *client.RouterClient, s *auth.Session, h *ws.Hub) *Poller {
	return &Poller{client: c, session: s, hub: h}
}

func (p *Poller) Run(ctx context.Context) {
	var prev state.DeviceState
	first := true
	status := statusOK

	log.Println("poller: started")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("poller: stopped")
			return

		case <-ticker.C:
			next, err := p.fetch()
			if err != nil {
				if status == statusOK {
					status = statusUnreachable
					log.Printf("poller: device unreachable: %v", err)
					p.hub.Broadcast(events.Event{
						Type:    events.EventDeviceUnreachable,
						Payload: map[string]string{"reason": err.Error()},
						At:      time.Now(),
					})
				}
				ticker.Reset(retryInterval)
				continue
			}

			if status == statusUnreachable {
				if err := p.session.ValidateMF935(); err != nil {
					if status != statusMismatch {
						status = statusMismatch
						log.Printf("poller: device mismatch: %v", err)
						p.hub.Broadcast(events.Event{
							Type:    events.EventDeviceMismatch,
							Payload: map[string]string{"reason": sanitizeMismatchReason(err.Error())},
							At:      time.Now(),
						})
					}
					continue
				}

				log.Println("poller: device recovered")
				status = statusOK
				first = true
				ticker.Reset(pollInterval)
				p.hub.Broadcast(events.Event{
					Type:    events.EventDeviceRecovered,
					Payload: map[string]string{},
					At:      time.Now(),
				})
			}

			if status == statusMismatch {
				if err := p.session.ValidateMF935(); err != nil {
					continue
				}
				log.Println("poller: mismatch resolved")
				status = statusOK
				first = true
				ticker.Reset(pollInterval)
				p.hub.Broadcast(events.Event{
					Type:    events.EventDeviceRecovered,
					Payload: map[string]string{},
					At:      time.Now(),
				})
			}

			if first {
				p.hub.StoreSnapshot(events.Event{
					Type:    events.EventSnapshot,
					Payload: next,
					At:      time.Now(),
				})
				first = false
			} else {
				for _, ev := range diff.Compute(prev, next) {
					p.hub.Broadcast(ev)
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
		PPPStatus:           data["ppp_status"],
		NetworkType:         data["network_type"],
		NetworkProvider:     data["network_provider"],
		WanIP:               data["wan_ipaddr"],
		SimcardRoam:         data["simcard_roam"],
		DialMode:            data["dial_mode"],
		ModemMainState:      data["modem_main_state"],
		SignalBars:          data["signalbar"],
		RSSI:                data["rssi"],
		LTERsrp:             data["lte_rsrp"],
		RSCP:                data["rscp"],
		BatteryPct:          data["battery_vol_percent"],
		BatteryCharging:     data["battery_charging"],
		BatteryPers:         data["battery_pers"],
		BatteryValue:        data["battery_value"],
		SSID1:               data["SSID1"],
		WifiState:           data["wifi_cur_state"],
		StaCount:            data["sta_count"],
		MStaCount:           data["m_sta_count"],
		StationMAC:          data["station_mac"],
		RxThrpt:             data["realtime_rx_thrpt"],
		TxThrpt:             data["realtime_tx_thrpt"],
		RxBytes:             data["realtime_rx_bytes"],
		TxBytes:             data["realtime_tx_bytes"],
		RealtimeTime:        data["realtime_time"],
		MonthlyRxBytes:      data["monthly_rx_bytes"],
		MonthlyTxBytes:      data["monthly_tx_bytes"],
		DateMonth:           data["date_month"],
		SmsUnreadNum:        data["sms_unread_num"],
		SmsDbChange:         data["sms_db_change"],
		SmsReceivedFlag:     data["sms_received_flag"],
		DataLimitSwitch:     data["data_volume_limit_switch"],
		DataLimitSize:       data["data_volume_limit_size"],
		DataLimitUnit:       data["data_volume_limit_unit"],
		DataAlertPct:        data["data_volume_alert_percent"],
		WanLteCa:            data["wan_lte_ca"],
		NewVersionState:     data["new_version_state"],
		CurrentUpgradeState: data["current_upgrade_state"],
		Loginfo:             data["loginfo"],
		LastUpdated:         time.Now(),
	}, nil
}

func sanitizeMismatchReason(err string) string {
	if strings.Contains(err, "wa_inner_version") {
		return "device at 192.168.0.1 is not an MTN MF935"
	}
	return "unknown device"
}
