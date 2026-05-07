# mf935-telemetry

A Go daemon that replaces the MF935 router's native polling model with a single batched poller and a WebSocket event stream. Instead of 10–30 HTTP requests per second from multiple clients, one goroutine holds the device connection and fans out only what changed.

Includes a web dashboard accessible from any device on the same network.

---

## Architecture

```
MF935 device (192.168.0.1)
      │
      │  1 batched GET every 2s
      ▼
┌─────────────────────────────────┐
│         Go Middleware           │
│                                 │
│  ┌──────────┐  ┌─────────────┐ │
│  │  Poller  │→ │ Diff engine │ │
│  └──────────┘  └──────┬──────┘ │
│                        │ events │
│                 ┌──────▼──────┐ │
│                 │  WS Server  │ │
│                 └─────────────┘ │
└───────────────────┬─────────────┘
                    │
          ┌─────────┴──────────┐
          │                    │
   ws://host:9000/stream   http://host:9000
   (React Native app)      (Web dashboard)
```

### Package layout

```
mf935-telemetry/
├── cmd/
│   └── server/
│       ├── main.go          ← entry point, wiring
│       ├── ui.go            ← static file server + /api/login handler
│       └── config.go        ← config.json reader
├── internal/
│   ├── auth/
│   │   └── session.go       ← login, AD token computation
│   ├── client/
│   │   └── router.go        ← HTTP client for goform endpoints
│   ├── diff/
│   │   └── diff.go          ← state comparison → events
│   ├── events/
│   │   └── events.go        ← event type definitions and constants
│   ├── poller/
│   │   └── poller.go        ← single poll goroutine
│   ├── state/
│   │   └── model.go         ← DeviceState struct
│   └── ws/
│       └── server.go        ← WebSocket hub and broadcaster
├── static/
│   ├── index.html           ← dashboard UI
│   ├── css/
│   │   └── style.css        ← MTN-themed styles
│   └── js/
│       ├── app.js           ← UI updates, event wiring
│       ├── auth.js          ← login modal, cookie, re-auth
│       ├── events.js        ← event constants and payload types
│       └── socket.js        ← WebSocket connection manager
├── config.json              ← optional, gitignored
├── go.mod
├── go.sum
└── README.md
```

---

## Requirements

- Go 1.21+
- Network access to the MF935 device at `192.168.0.1`
- Router admin password

---

## Setup

```bash
git clone https://github.com/invalid8/mf935-telemetry
cd mf935-telemetry
go mod tidy
go build ./cmd/server/
```

---

## Running

```bash
./server
```

On startup the server prints:

```
server: dashboard → http://192.168.x.x:9000
server: ws stream → ws://192.168.x.x:9000/stream
```

Open the dashboard URL from any device on your network.

---

## Authentication

### Option A — Config file (quick test / dev)

Create `config.json` in the project root:

```json
{
  "password": "your_router_password"
}
```

The server reads this on startup, authenticates immediately, and starts polling. No modal shown.

`config.json` is gitignored — never committed.

### Option B — Web UI login

Without `config.json`, the dashboard loads with a login modal. Enter your router password, click connect.

- Password is sent plaintext over the local network to `POST /api/login`
- Go hashes it (`SHA256(SHA256(password) + LD).toUpperCase()`) before it touches the device
- On success, a cookie is set — subsequent page loads re-authenticate silently without showing the modal
- Cookie expires after 30 days

---

## WebSocket API

Connect to:

```
ws://localhost:9000/stream
```

On connect, the server immediately replays the latest `snapshot` event with full device state. After that, only changed fields are emitted as typed events.

All messages use this envelope:

```json
{
  "event": "<event_type>",
  "payload": { ... },
  "at": "2026-05-05T20:39:36Z"
}
```

---

### Event Reference

#### `snapshot`

Sent immediately on client connect. Full device state.

```json
{
  "event": "snapshot",
  "payload": {
    "ppp_status": "ppp_connected",
    "network_type": "LTE",
    "network_provider": "MTN Nigeria",
    "wan_ipaddr": "10.100.237.95",
    "simcard_roam": "Home",
    "dial_mode": "auto_dial",
    "modem_main_state": "modem_init_complete",
    "signalbar": "3",
    "rssi": "-106",
    "lte_rsrp": "-106",
    "rscp": "",
    "battery_vol_percent": "79",
    "battery_charging": "1",
    "battery_pers": "3",
    "SSID1": "My WiFi",
    "wifi_cur_state": "1",
    "sta_count": "2",
    "station_mac": "34:cf:f6:98:e3:1a;4e:1a:05:7c:d0:cc",
    "realtime_rx_thrpt": "59704",
    "realtime_tx_thrpt": "2596",
    "realtime_rx_bytes": "1074694135",
    "realtime_tx_bytes": "195213116",
    "realtime_time": "12628",
    "monthly_rx_bytes": "24835509607",
    "monthly_tx_bytes": "2210054890",
    "sms_unread_num": "9",
    "data_volume_limit_switch": "0",
    "wan_lte_ca": "",
    "new_version_state": "version_idle",
    "current_upgrade_state": "fota_idle",
    "loginfo": "ok",
    "last_updated": "2026-05-05T20:39:36Z"
  },
  "at": "2026-05-05T20:39:36Z"
}
```

---

#### `signal_changed`

Fired when signal bars, RSSI, LTE RSRP, or network type changes.

```json
{
  "event": "signal_changed",
  "payload": {
    "signalbar": "3",
    "rssi": "-106",
    "lte_rsrp": "-106",
    "network_type": "LTE"
  },
  "at": "..."
}
```

---

#### `connection_gained`

Fired when `ppp_status` transitions to `ppp_connected`.

```json
{
  "event": "connection_gained",
  "payload": {
    "ppp_status": "ppp_connected",
    "network_provider": "MTN Nigeria",
    "wan_ipaddr": "10.100.237.95",
    "network_type": "LTE"
  },
  "at": "..."
}
```

---

#### `connection_lost`

Fired when `ppp_status` transitions away from `ppp_connected`.

```json
{
  "event": "connection_lost",
  "payload": {
    "ppp_status": "ppp_disconnected",
    "network_provider": "MTN Nigeria",
    "wan_ipaddr": "",
    "network_type": "LTE"
  },
  "at": "..."
}
```

---

#### `network_changed`

Fired when provider or WAN IP changes without a full connect/disconnect.

```json
{
  "event": "network_changed",
  "payload": {
    "ppp_status": "ppp_connected",
    "network_provider": "MTN Nigeria",
    "wan_ipaddr": "10.100.237.95",
    "network_type": "LTE"
  },
  "at": "..."
}
```

---

#### `battery_changed`

Fired when battery percentage, charging state, or bars change.

```json
{
  "event": "battery_changed",
  "payload": {
    "battery_vol_percent": "79",
    "battery_charging": "1",
    "battery_pers": "3",
    "battery_value": ""
  },
  "at": "..."
}
```

---

#### `clients_changed`

Fired when connected WiFi client count or MACs change.

```json
{
  "event": "clients_changed",
  "payload": {
    "sta_count": "2",
    "m_sta_count": "0",
    "station_mac": "34:cf:f6:98:e3:1a;4e:1a:05:7c:d0:cc"
  },
  "at": "..."
}
```

`station_mac` is semicolon-separated. Split on `";"` and discard the trailing empty string.

---

#### `sms_received`

Fired when `sms_unread_num` increases.

```json
{
  "event": "sms_received",
  "payload": { "sms_unread_num": "9" },
  "at": "..."
}
```

---

#### `sms_db_changed`

Fired when the SMS database changes — new, deleted, or read. Use this as the trigger to refetch the inbox.

```json
{
  "event": "sms_db_changed",
  "payload": { "sms_db_change": "1" },
  "at": "..."
}
```

---

#### `throughput_updated`

Fired every poll cycle when rx or tx throughput changes. Expect this frequently.

```json
{
  "event": "throughput_updated",
  "payload": {
    "realtime_rx_thrpt": "59704",
    "realtime_tx_thrpt": "2596",
    "realtime_rx_bytes": "1074694135",
    "realtime_tx_bytes": "195213116",
    "realtime_time": "12628"
  },
  "at": "..."
}
```

Throughput values are in **bytes/s**. Session bytes and time reset when the connection drops and reconnects.

---

#### `data_limit_changed`

Fired when data volume limit settings change.

```json
{
  "event": "data_limit_changed",
  "payload": {
    "data_volume_limit_switch": "1",
    "data_volume_limit_size": "50",
    "data_volume_limit_unit": "GB",
    "data_volume_alert_percent": "80"
  },
  "at": "..."
}
```

---

#### `ota_available`

Fired when `new_version_state` changes.

```json
{
  "event": "ota_available",
  "payload": {
    "new_version_state": "version_has_new_software",
    "current_upgrade_state": "fota_idle"
  },
  "at": "..."
}
```

---

## Testing the Stream

```bash
# Install websocat
curl -Lo websocat https://github.com/vi/websocat/releases/download/v1.13.0/websocat.x86_64-unknown-linux-musl
chmod +x websocat

# Connect
./websocat ws://localhost:9000/stream
```

---

## Known Device Behaviour

| Behaviour | Detail |
|---|---|
| `network_provider = "62130"` | Raw PLMN — map to operator name in app |
| WiFi passwords | Returned as Base64 (`WPAPSK1_encode`) — decode before displaying |
| `sms_unread_num` | Use directly |
| SMS `tag = "1"` | Means unread |
| `station_mac` | Semicolon-delimited, trailing semicolon produces empty last element |
| `crypto.subtle` | Unavailable on plain HTTP non-localhost — all hashing is done server-side |