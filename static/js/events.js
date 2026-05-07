/**
 * events.js
 * Event name constants and payload shape definitions for the MF935 WebSocket stream.
 * Every event type the server can emit is defined here.
 * Import this before socket.js and app.js.
 */

// ---------------------------------------------------------------------------
// Event name constants
// Mirror of internal/events/events.go
// ---------------------------------------------------------------------------

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
};

// ---------------------------------------------------------------------------
// Payload shapes (JSDoc — for editor hints and self-documentation)
// ---------------------------------------------------------------------------

/**
 * @typedef {Object} SnapshotPayload
 * Full device state — sent immediately on connect.
 * @property {string} ppp_status          - "ppp_connected" | "ppp_disconnected" | "ppp_connecting"
 * @property {string} network_type        - "LTE" | "WCDMA" | "no_service"
 * @property {string} network_provider    - Raw PLMN e.g. "62130" → map via PROVIDER_MAP
 * @property {string} wan_ipaddr          - WAN IPv4
 * @property {string} simcard_roam        - "Home" | "International"
 * @property {string} dial_mode           - "auto_dial" | "manual_dial"
 * @property {string} modem_main_state    - "modem_init_complete" | "modem_waitpin" | ...
 * @property {string} signalbar           - "0"–"5"
 * @property {string} rssi                - dBm string
 * @property {string} lte_rsrp            - dBm string
 * @property {string} rscp                - dBm string (WCDMA only)
 * @property {string} battery_vol_percent - "0"–"100"
 * @property {string} battery_charging    - "1" = charging, "0" = not
 * @property {string} battery_pers        - "0"–"4" bars
 * @property {string} SSID1               - Primary WiFi network name
 * @property {string} wifi_cur_state      - "1" = on, "0" = off
 * @property {string} sta_count           - Connected client count (string)
 * @property {string} station_mac         - Semicolon-separated MAC list
 * @property {string} realtime_rx_thrpt   - bytes/s download
 * @property {string} realtime_tx_thrpt   - bytes/s upload
 * @property {string} realtime_rx_bytes   - Session download bytes
 * @property {string} realtime_tx_bytes   - Session upload bytes
 * @property {string} realtime_time       - Session duration in seconds
 * @property {string} monthly_rx_bytes    - Month download bytes
 * @property {string} monthly_tx_bytes    - Month upload bytes
 * @property {string} sms_unread_num      - Unread SMS count (string)
 * @property {string} sms_db_change       - Changes when SMS DB is modified
 * @property {string} data_volume_limit_switch - "0" = off, "1" = on
 * @property {string} wan_lte_ca          - "ca_activated" | "ca_deactivated"
 * @property {string} new_version_state   - "version_idle" | "version_has_new_software"
 * @property {string} last_updated        - ISO timestamp
 */

/**
 * @typedef {Object} SignalPayload
 * @property {string} signalbar
 * @property {string} rssi
 * @property {string} lte_rsrp
 * @property {string} network_type
 */

/**
 * @typedef {Object} ConnectionPayload
 * @property {string} ppp_status
 * @property {string} network_provider
 * @property {string} wan_ipaddr
 * @property {string} network_type
 */

/**
 * @typedef {Object} BatteryPayload
 * @property {string} battery_vol_percent
 * @property {string} battery_charging
 * @property {string} battery_pers
 * @property {string} battery_value
 */

/**
 * @typedef {Object} ClientsPayload
 * @property {string} sta_count
 * @property {string} m_sta_count
 * @property {string} station_mac   - Split on ";" — last element is always empty
 */

/**
 * @typedef {Object} ThroughputPayload
 * @property {string} realtime_rx_thrpt
 * @property {string} realtime_tx_thrpt
 * @property {string} realtime_rx_bytes
 * @property {string} realtime_tx_bytes
 * @property {string} realtime_time
 */

/**
 * @typedef {Object} SmsPayload
 * @property {string} sms_unread_num
 */

/**
 * @typedef {Object} SmsDbPayload
 * @property {string} sms_db_change
 */

/**
 * @typedef {Object} DataLimitPayload
 * @property {string} data_volume_limit_switch
 * @property {string} data_volume_limit_size
 * @property {string} data_volume_limit_unit
 * @property {string} data_volume_alert_percent
 */

/**
 * @typedef {Object} OTAPayload
 * @property {string} new_version_state
 * @property {string} current_upgrade_state
 */

// ---------------------------------------------------------------------------
// Display helpers
// ---------------------------------------------------------------------------

/** Maps raw PLMN codes to readable operator names */
export const PROVIDER_MAP = {
  62130: "MTN Nigeria",
  62150: "Glo Nigeria",
  62120: "Airtel Nigeria",
  62160: "9mobile Nigeria",
};

/** Resolve a raw provider string to a display name */
export function resolveProvider(raw) {
  return PROVIDER_MAP[raw] ?? raw;
}

/** Format bytes into a human-readable string */
export function fmtBytes(n) {
  n = parseInt(n, 10);
  if (isNaN(n) || n === 0) return "0 B";
  if (n < 1024) return n + " B";
  if (n < 1048576) return (n / 1024).toFixed(1) + " KB";
  if (n < 1073741824) return (n / 1048576).toFixed(1) + " MB";
  return (n / 1073741824).toFixed(2) + " GB";
}

/** Format seconds into HH:MM:SS */
export function fmtDuration(s) {
  s = parseInt(s, 10);
  if (isNaN(s)) return "—";
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  return [h, m, sec].map((v) => String(v).padStart(2, "0")).join(":");
}

/** Signal bar count to a label */
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
