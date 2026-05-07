package diff

import (
	"time"

	"mf935-telemetry/internal/events"
	"mf935-telemetry/internal/state"
)

func Compute(prev, next state.DeviceState) []events.Event {
	var out []events.Event
	now := time.Now()

	if prev.SignalBars != next.SignalBars ||
		prev.RSSI != next.RSSI ||
		prev.LTERsrp != next.LTERsrp ||
		prev.NetworkType != next.NetworkType {
		out = append(out, events.Event{
			Type: events.EventSignalChanged,
			Payload: map[string]string{
				"signalbar":    next.SignalBars,
				"rssi":         next.RSSI,
				"lte_rsrp":     next.LTERsrp,
				"network_type": next.NetworkType,
			},
			At: now,
		})
	}

	if prev.PPPStatus != next.PPPStatus ||
		prev.NetworkProvider != next.NetworkProvider ||
		prev.WanIP != next.WanIP {
		eventType := events.EventNetworkChanged

		if next.PPPStatus == "ppp_connected" {
			eventType = events.EventConnectionGained
		} else if prev.PPPStatus == "ppp_connected" {
			eventType = events.EventConnectionLost
		}

		out = append(out, events.Event{
			Type: eventType,
			Payload: map[string]string{
				"ppp_status":       next.PPPStatus,
				"network_provider": next.NetworkProvider,
				"wan_ipaddr":       next.WanIP,
				"network_type":     next.NetworkType,
			},
			At: now,
		})
	}

	if prev.BatteryPct != next.BatteryPct ||
		prev.BatteryCharging != next.BatteryCharging ||
		prev.BatteryPers != next.BatteryPers {
		out = append(out, events.Event{
			Type: events.EventBatteryChanged,
			Payload: map[string]string{
				"battery_vol_percent": next.BatteryPct,
				"battery_charging":    next.BatteryCharging,
				"battery_pers":        next.BatteryPers,
				"battery_value":       next.BatteryValue,
			},
			At: now,
		})
	}

	if prev.StaCount != next.StaCount ||
		prev.MStaCount != next.MStaCount ||
		prev.StationMAC != next.StationMAC {
		out = append(out, events.Event{
			Type: events.EventClientsChanged,
			Payload: map[string]string{
				"sta_count":   next.StaCount,
				"m_sta_count": next.MStaCount,
				"station_mac": next.StationMAC,
			},
			At: now,
		})
	}

	if prev.SmsUnreadNum != next.SmsUnreadNum {
		out = append(out, events.Event{
			Type: events.EventSmsReceived,
			Payload: map[string]string{
				"sms_unread_num": next.SmsUnreadNum,
			},
			At: now,
		})
	}

	if prev.SmsDbChange != next.SmsDbChange {
		out = append(out, events.Event{
			Type: events.EventSmsDbChanged,
			Payload: map[string]string{
				"sms_db_change": next.SmsDbChange,
			},
			At: now,
		})
	}

	if prev.RxThrpt != next.RxThrpt || prev.TxThrpt != next.TxThrpt {
		out = append(out, events.Event{
			Type: events.EventThroughput,
			Payload: map[string]string{
				"realtime_rx_thrpt": next.RxThrpt,
				"realtime_tx_thrpt": next.TxThrpt,
				"realtime_rx_bytes": next.RxBytes,
				"realtime_tx_bytes": next.TxBytes,
				"realtime_time":     next.RealtimeTime,
			},
			At: now,
		})
	}

	if prev.DataLimitSwitch != next.DataLimitSwitch ||
		prev.DataLimitSize != next.DataLimitSize ||
		prev.DataAlertPct != next.DataAlertPct {
		out = append(out, events.Event{
			Type: events.EventDataLimitChanged,
			Payload: map[string]string{
				"data_volume_limit_switch":  next.DataLimitSwitch,
				"data_volume_limit_size":    next.DataLimitSize,
				"data_volume_limit_unit":    next.DataLimitUnit,
				"data_volume_alert_percent": next.DataAlertPct,
			},
			At: now,
		})
	}

	if prev.NewVersionState != next.NewVersionState {
		out = append(out, events.Event{
			Type: events.EventOTAAvailable,
			Payload: map[string]string{
				"new_version_state":     next.NewVersionState,
				"current_upgrade_state": next.CurrentUpgradeState,
			},
			At: now,
		})
	}

	return out
}
