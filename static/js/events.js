/**
 * events.js
 * Event name constants and payload shape definitions for the MF935 WebSocket stream.
 * Every event type the server can emit is defined here.
 * Import this before socket.js and app.js.
 */

export const EVENTS = {
  SNAPSHOT: "snapshot",
  SIGNAL_CHANGED: "signal_changed",
  NETWORK_CHANGED: "network_changed",
  CONNECTION_GAINED: "connection_gained",
  CONNECTION_LOST: "connection_lost",
  BATTERY_CHANGED: "battery_changed",
  CLIENTS_CHANGED: "clients_changed",
  SMS_RECEIVED: "sms_received",
  SMS_DB_CHANGED: "sms_db_changed",
  THROUGHPUT: "throughput_updated",
  DATA_LIMIT_CHANGED: "data_limit_changed",
  OTA_AVAILABLE: "ota_available",

  DEVICE_UNREACHABLE: "device_unreachable",
  DEVICE_MISMATCH: "device_mismatch",
  DEVICE_RECOVERED: "device_recovered",
};

/**
 * @typedef {Object} DeviceUnreachablePayload
 * @property {string} reason - Human-readable description
 */

/**
 * @typedef {Object} DeviceMismatchPayload
 * @property {string} reason - e.g. "device at 192.168.0.1 is not an MTN MF935"
 */

/**
 * @typedef {Object} SnapshotPayload
 * @property {string} ppp_status
 * @property {string} network_type
 * @property {string} network_provider
 * @property {string} wan_ipaddr
 * @property {string} simcard_roam
 * @property {string} dial_mode
 * @property {string} modem_main_state
 * @property {string} signalbar
 * @property {string} rssi
 * @property {string} lte_rsrp
 * @property {string} rscp
 * @property {string} battery_vol_percent
 * @property {string} battery_charging
 * @property {string} battery_pers
 * @property {string} SSID1
 * @property {string} wifi_cur_state
 * @property {string} sta_count
 * @property {string} station_mac
 * @property {string} realtime_rx_thrpt
 * @property {string} realtime_tx_thrpt
 * @property {string} realtime_rx_bytes
 * @property {string} realtime_tx_bytes
 * @property {string} realtime_time
 * @property {string} monthly_rx_bytes
 * @property {string} monthly_tx_bytes
 * @property {string} sms_unread_num
 * @property {string} sms_db_change
 * @property {string} data_volume_limit_switch
 * @property {string} wan_lte_ca
 * @property {string} new_version_state
 * @property {string} last_updated
 */

/** Maps raw PLMN codes to readable operator names */
export const PROVIDER_MAP = {
  62130: "MTN Nigeria",
  62150: "Glo Nigeria",
  62120: "Airtel Nigeria",
  62160: "9mobile Nigeria",
};

export function resolveProvider(raw) {
  return PROVIDER_MAP[raw] ?? raw;
}

export function fmtBytes(n) {
  n = parseInt(n, 10);
  if (isNaN(n) || n === 0) return "0 B";
  if (n < 1024) return n + " B";
  if (n < 1048576) return (n / 1024).toFixed(1) + " KB";
  if (n < 1073741824) return (n / 1048576).toFixed(1) + " MB";
  return (n / 1073741824).toFixed(2) + " GB";
}

export function fmtDuration(s) {
  s = parseInt(s, 10);
  if (isNaN(s)) return "—";
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  return [h, m, sec].map((v) => String(v).padStart(2, "0")).join(":");
}

export function fmtSignal(bars) {
  const map = {
    0: "No signal",
    1: "Poor",
    2: "Weak",
    3: "Fair",
    4: "Good",
    5: "Excellent",
  };
  return map[bars] ?? bars;
}
