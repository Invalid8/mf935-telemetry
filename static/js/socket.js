/**
 * socket.js
 * WebSocket connection manager for the MF935 telemetry stream.
 * Handles connect, reconnect, and dispatches typed events to registered handlers.
 */

import { EVENTS } from "./events.js";

const WS_URL = `ws://${location.host}/stream`;
const RECONNECT_DELAY_MS = 3000;

export class TelemetrySocket {
  /** @type {WebSocket|null} */
  #ws = null;

  /** @type {Map<string, Function[]>} */
  #handlers = new Map();

  /** @type {Function|null} */
  #onStatusChange = null;

  /**
   * @param {(status: "connecting"|"connected"|"disconnected") => void} onStatusChange
   */
  constructor(onStatusChange) {
    this.#onStatusChange = onStatusChange;
  }

  // ---------------------------------------------------------------------------
  // Public API
  // ---------------------------------------------------------------------------

  /** Start the connection. Safe to call multiple times. */
  connect() {
    if (this.#ws && this.#ws.readyState === WebSocket.OPEN) return;
    this.#setStatus("connecting");
    this.#open();
  }

  /**
   * Register a handler for a specific event type.
   * @param {string} eventType - Use EVENTS.* constants
   * @param {(payload: any, at: string) => void} handler
   */
  on(eventType, handler) {
    if (!this.#handlers.has(eventType)) {
      this.#handlers.set(eventType, []);
    }
    this.#handlers.get(eventType).push(handler);
    return this;
  }

  // ---------------------------------------------------------------------------
  // Internal
  // ---------------------------------------------------------------------------

  #open() {
    this.#ws = new WebSocket(WS_URL);

    this.#ws.onopen = () => {
      this.#setStatus("connected");
    };

    this.#ws.onmessage = (msg) => {
      let parsed;
      try {
        parsed = JSON.parse(msg.data);
      } catch (e) {
        console.error("socket: failed to parse message", e);
        return;
      }

      const { event, payload, at } = parsed;
      this.#dispatch(event, payload, at);
    };

    this.#ws.onclose = () => {
      this.#setStatus("disconnected");
      setTimeout(() => this.#open(), RECONNECT_DELAY_MS);
    };

    this.#ws.onerror = () => {
      this.#ws.close();
    };
  }

  /** Dispatch an event to all registered handlers */
  #dispatch(type, payload, at) {
    const handlers = this.#handlers.get(type) ?? [];
    for (const handler of handlers) {
      try {
        handler(payload, at);
      } catch (e) {
        console.error(`socket: handler error for "${type}"`, e);
      }
    }
  }

  #setStatus(status) {
    this.#onStatusChange?.(status);
  }
}
