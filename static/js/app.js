/**
 * app.js
 * Main entry point — wires socket events to DOM updates.
 * Handles auth check before connecting the socket.
 * Wrapper is hidden until auth succeeds — no flash.
 */

import {
  EVENTS,
  fmtBytes,
  fmtDuration,
  fmtSignal,
  resolveProvider,
} from "./events.js";
import { TelemetrySocket } from "./socket.js";
import { getPasswordCookie, reAuthWithCookie, showLoginModal } from "./auth.js";

// ---------------------------------------------------------------------------
// DOM refs
// ---------------------------------------------------------------------------

const $ = (id) => document.getElementById(id);

const els = {
  wrapper: document.querySelector(".wrapper"),
  statusDot: $("status-dot"),
  statusLabel: $("status-label"),
  pppStatus: $("ppp_status"),
  networkType: $("network_type"),
  networkProvider: $("network_provider"),
  wanIp: $("wan_ipaddr"),
  roaming: $("simcard_roam"),
  dialMode: $("dial_mode"),
  lteCa: $("wan_lte_ca"),
  signalbar: $("signalbar"),
  rssi: $("rssi"),
  lteRsrp: $("lte_rsrp"),
  batteryPct: $("battery_vol_percent"),
  batteryCharging: $("battery_charging"),
  batteryPers: $("battery_pers"),
  batteryBar: $("battery-bar"),
  ssid: $("SSID1"),
  wifiState: $("wifi_cur_state"),
  staCount: $("sta_count"),
  rxThrpt: $("realtime_rx_thrpt"),
  txThrpt: $("realtime_tx_thrpt"),
  rxBytes: $("realtime_rx_bytes"),
  txBytes: $("realtime_tx_bytes"),
  sessTime: $("realtime_time"),
  monthlyRx: $("monthly_rx_bytes"),
  monthlyTx: $("monthly_tx_bytes"),
  smsUnread: $("sms_unread_num"),
  logList: $("log-list"),
};

// ---------------------------------------------------------------------------
// Status
// ---------------------------------------------------------------------------

function setStatus(status) {
  els.statusDot.className = "status-dot " + status;
  els.statusLabel.textContent = status;
}

// ---------------------------------------------------------------------------
// Renderers
// ---------------------------------------------------------------------------

function renderNetwork(p) {
  if (p.ppp_status !== undefined) {
    const connected = p.ppp_status === "ppp_connected";
    els.pppStatus.textContent = p.ppp_status
      .replace("ppp_", "")
      .replace(/_/g, " ");
    els.pppStatus.className =
      "val " +
      (connected ? "green" : p.ppp_status.includes("ing") ? "yellow" : "red");
  }
  if (p.network_type !== undefined)
    els.networkType.textContent = p.network_type || "—";
  if (p.network_provider !== undefined)
    els.networkProvider.textContent =
      resolveProvider(p.network_provider) || "—";
  if (p.wan_ipaddr !== undefined) els.wanIp.textContent = p.wan_ipaddr || "—";
  if (p.simcard_roam !== undefined)
    els.roaming.textContent = p.simcard_roam || "—";
  if (p.dial_mode !== undefined)
    els.dialMode.textContent = p.dial_mode?.replace(/_/g, " ") || "—";
  if (p.wan_lte_ca !== undefined)
    els.lteCa.textContent =
      p.wan_lte_ca === "ca_activated" ? "activated" : p.wan_lte_ca || "—";
}

function renderSignal(p) {
  if (p.signalbar !== undefined) {
    els.signalbar.textContent =
      fmtSignal(p.signalbar) + " (" + p.signalbar + "/5)";
    updateSignalBars(p.signalbar);
  }
  if (p.rssi !== undefined)
    els.rssi.textContent = p.rssi ? p.rssi + " dBm" : "—";
  if (p.lte_rsrp !== undefined)
    els.lteRsrp.textContent = p.lte_rsrp ? p.lte_rsrp + " dBm" : "—";
}

function renderBattery(p) {
  if (p.battery_vol_percent !== undefined) {
    const pct = parseInt(p.battery_vol_percent, 10);
    els.batteryPct.textContent = isNaN(pct) ? "—" : pct + "%";
    els.batteryBar.style.width = (isNaN(pct) ? 0 : pct) + "%";
    els.batteryBar.className =
      "battery-fill " + (pct > 50 ? "green" : pct > 20 ? "yellow" : "red");
  }
  if (p.battery_charging !== undefined) {
    els.batteryCharging.textContent =
      p.battery_charging === "1" ? "charging ⚡" : "not charging";
    els.batteryCharging.className =
      "val " + (p.battery_charging === "1" ? "green" : "");
  }
  if (p.battery_pers !== undefined)
    els.batteryPers.textContent = p.battery_pers + "/4 bars";
}

function renderWifi(p) {
  if (p.SSID1 !== undefined) els.ssid.textContent = p.SSID1 || "—";
  if (p.wifi_cur_state !== undefined) {
    els.wifiState.textContent = p.wifi_cur_state === "1" ? "on" : "off";
    els.wifiState.className =
      "val " + (p.wifi_cur_state === "1" ? "green" : "red");
  }
  if (p.sta_count !== undefined)
    els.staCount.textContent = p.sta_count + " connected";
}

function renderThroughput(p) {
  if (p.realtime_rx_thrpt !== undefined)
    els.rxThrpt.textContent = fmtBytes(p.realtime_rx_thrpt) + "/s";
  if (p.realtime_tx_thrpt !== undefined)
    els.txThrpt.textContent = fmtBytes(p.realtime_tx_thrpt) + "/s";
  if (p.realtime_rx_bytes !== undefined)
    els.rxBytes.textContent = fmtBytes(p.realtime_rx_bytes);
  if (p.realtime_tx_bytes !== undefined)
    els.txBytes.textContent = fmtBytes(p.realtime_tx_bytes);
  if (p.realtime_time !== undefined)
    els.sessTime.textContent = fmtDuration(p.realtime_time);
}

function renderMonthly(p) {
  if (p.monthly_rx_bytes !== undefined)
    els.monthlyRx.textContent = fmtBytes(p.monthly_rx_bytes);
  if (p.monthly_tx_bytes !== undefined)
    els.monthlyTx.textContent = fmtBytes(p.monthly_tx_bytes);
}

function renderSms(p) {
  if (p.sms_unread_num !== undefined) {
    const count = parseInt(p.sms_unread_num, 10);
    els.smsUnread.textContent = isNaN(count) ? "—" : count;
    els.smsUnread.className = "val " + (count > 0 ? "yellow" : "");
  }
}

// ---------------------------------------------------------------------------
// Signal bars
// ---------------------------------------------------------------------------

function updateSignalBars(bars) {
  const n = parseInt(bars, 10);
  document.querySelectorAll(".signal-bar").forEach((bar, i) => {
    bar.classList.toggle("active", i < n);
  });
}

// ---------------------------------------------------------------------------
// Event log
// ---------------------------------------------------------------------------

function logEvent(type, payload, at) {
  const entry = document.createElement("div");
  entry.className = "log-entry";

  const time = new Date(at).toLocaleTimeString();
  const summary = Object.entries(payload)
    .slice(0, 3)
    .map(
      ([k, v]) =>
        `<span class="log-key">${k}</span>=<span class="log-val">${v}</span>`,
    )
    .join(" ");

  entry.innerHTML = `
    <span class="log-time">${time}</span>
    <span class="log-event">${type}</span>
    <span class="log-summary">${summary}</span>
  `;

  els.logList.prepend(entry);
  while (els.logList.children.length > 60) {
    els.logList.removeChild(els.logList.lastChild);
  }
}

// ---------------------------------------------------------------------------
// Snapshot
// ---------------------------------------------------------------------------

function applySnapshot(p) {
  renderNetwork(p);
  renderSignal(p);
  renderBattery(p);
  renderWifi(p);
  renderThroughput(p);
  renderMonthly(p);
  renderSms(p);
}

// ---------------------------------------------------------------------------
// Socket
// ---------------------------------------------------------------------------

function startSocket() {
  // Reveal dashboard only after auth clears — no flash
  els.wrapper.style.display = "block";

  const socket = new TelemetrySocket(setStatus);

  socket
    .on(EVENTS.SNAPSHOT, (p, at) => {
      applySnapshot(p);
      logEvent(
        EVENTS.SNAPSHOT,
        { fields: Object.keys(p).length + " fields" },
        at,
      );
    })
    .on(EVENTS.SIGNAL_CHANGED, (p, at) => {
      renderSignal(p);
      logEvent(EVENTS.SIGNAL_CHANGED, p, at);
    })
    .on(EVENTS.NETWORK_CHANGED, (p, at) => {
      renderNetwork(p);
      logEvent(EVENTS.NETWORK_CHANGED, p, at);
    })
    .on(EVENTS.CONNECTION_GAINED, (p, at) => {
      renderNetwork(p);
      logEvent(EVENTS.CONNECTION_GAINED, p, at);
    })
    .on(EVENTS.CONNECTION_LOST, (p, at) => {
      renderNetwork(p);
      logEvent(EVENTS.CONNECTION_LOST, p, at);
    })
    .on(EVENTS.BATTERY_CHANGED, (p, at) => {
      renderBattery(p);
      logEvent(EVENTS.BATTERY_CHANGED, p, at);
    })
    .on(EVENTS.CLIENTS_CHANGED, (p, at) => {
      renderWifi(p);
      logEvent(EVENTS.CLIENTS_CHANGED, p, at);
    })
    .on(EVENTS.SMS_RECEIVED, (p, at) => {
      renderSms(p);
      logEvent(EVENTS.SMS_RECEIVED, p, at);
    })
    .on(EVENTS.SMS_DB_CHANGED, (p, at) => {
      logEvent(EVENTS.SMS_DB_CHANGED, p, at);
    })
    .on(EVENTS.THROUGHPUT, (p, at) => {
      renderThroughput(p);
      logEvent(EVENTS.THROUGHPUT, p, at);
    })
    .on(EVENTS.DATA_LIMIT_CHANGED, (p, at) => {
      logEvent(EVENTS.DATA_LIMIT_CHANGED, p, at);
    })
    .on(EVENTS.OTA_AVAILABLE, (p, at) => {
      logEvent(EVENTS.OTA_AVAILABLE, p, at);
    });

  socket.connect();
}

// ---------------------------------------------------------------------------
// Boot — auth before socket
// ---------------------------------------------------------------------------

async function boot() {
  const cookie = getPasswordCookie();

  if (cookie) {
    const result = await reAuthWithCookie(cookie);
    if (result.ok) {
      startSocket();
      return;
    }
    // Stale cookie — fall through to modal
  }

  await showLoginModal();
  startSocket();
}

boot();
