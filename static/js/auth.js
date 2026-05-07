/**
 * auth.js
 * Handles password hashing, cookie storage, and the login modal flow.
 * Must be initialised before the socket connects.
 */

const COOKIE_NAME = "mf_pw";
const COOKIE_DAYS = 30;

// ---------------------------------------------------------------------------
// SHA-256 — browser native
// ---------------------------------------------------------------------------

async function sha256Upper(str) {
  const buf = await crypto.subtle.digest(
    "SHA-256",
    new TextEncoder().encode(str),
  );
  const hex = Array.from(new Uint8Array(buf))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
  return hex.toUpperCase();
}

// ---------------------------------------------------------------------------
// Cookie helpers
// ---------------------------------------------------------------------------

export function getPasswordCookie() {
  const match = document.cookie
    .split("; ")
    .find((r) => r.startsWith(COOKIE_NAME + "="));
  return match ? decodeURIComponent(match.split("=")[1]) : null;
}

function setPasswordCookie(value) {
  const expires = new Date(Date.now() + COOKIE_DAYS * 864e5).toUTCString();
  document.cookie = `${COOKIE_NAME}=${encodeURIComponent(value)}; expires=${expires}; path=/; SameSite=Strict`;
}

export function clearPasswordCookie() {
  document.cookie = `${COOKIE_NAME}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/`;
}

// ---------------------------------------------------------------------------
// Login API
// ---------------------------------------------------------------------------

/**
 * Sends a pre-hashed password to /api/login.
 * @param {string} preHashed - SHA256(plaintext).toUpperCase()
 * @returns {Promise<{ok: boolean, message?: string}>}
 */
async function postLogin(password) {
  try {
    const res = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password }),
    });
    return await res.json();
  } catch {
    return { ok: false, message: "network error — is the server running?" };
  }
}

/**
 * Hashes plaintext then logs in. Stores cookie on success.
 * @param {string} plaintext
 */

export async function attemptLogin(plaintext) {
  const result = await postLogin(plaintext);
  if (result.ok) setPasswordCookie(plaintext);
  return result;
}

/**
 * Re-authenticates using the stored cookie (already pre-hashed).
 * @param {string} preHashed
 */
export async function reAuthWithCookie(plaintext) {
  return postLogin(plaintext);
}

// ---------------------------------------------------------------------------
// Modal
// ---------------------------------------------------------------------------

const overlay = () => document.getElementById("login-overlay");
const input = () => document.getElementById("login-password");
const btn = () => document.getElementById("login-submit");
const errEl = () => document.getElementById("login-error");

function setLoading(loading) {
  btn().disabled = loading;
  btn().textContent = loading ? "authenticating..." : "connect";
}

function setError(msg) {
  errEl().textContent = msg;
}

function resetModal() {
  setLoading(false);
  setError("");
  input().value = "";
}

/**
 * Shows the login modal. Resolves only when login succeeds.
 * @returns {Promise<void>}
 */
export function showLoginModal() {
  return new Promise((resolve) => {
    overlay().classList.add("visible");

    // Autofocus after transition
    setTimeout(() => input().focus(), 50);

    async function submit() {
      const val = input().value.trim();
      if (!val) {
        setError("password cannot be empty");
        return;
      }

      setLoading(true);
      setError("");

      let result;
      try {
        result = await attemptLogin(val);
      } catch {
        result = { ok: false, message: "unexpected error" };
      }

      if (result.ok) {
        overlay().classList.remove("visible");
        resolve();
        return;
      }

      // Failed — always reset so user can try again
      resetModal();
      setError(result.message ?? "incorrect password");
      input().focus();
    }

    btn().addEventListener("click", submit);
    input().addEventListener("keydown", (e) => {
      if (e.key === "Enter") submit();
    });
  });
}
