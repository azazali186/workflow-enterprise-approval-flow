import { ensureFreshAccessToken } from '@/services/api/client'
import { storage } from '@/utils/storage'

/**
 * WebSocket event envelope sent by the backend hub (see
 * backend/internal/pkg/websocket/hub.go — SendToUser / SendToAll).
 */
export interface RealtimeEvent<T = Record<string, unknown>> {
  event: string
  payload: T
}

export type RealtimeHandler = (event: RealtimeEvent) => void

const INITIAL_BACKOFF_MS = 1_000
const MAX_BACKOFF_MS = 30_000

function wsUrl(token: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  // The backend accepts the token as ?token= because browsers cannot set the
  // Authorization header on a WebSocket handshake.
  return `${protocol}//${window.location.host}/ws?token=${encodeURIComponent(token)}`
}

/**
 * Authenticated WebSocket client for the backend's /ws endpoint.
 *
 * - Connects with the current access token (browser-safe ?token= handshake).
 * - Auto-reconnects with exponential backoff (1s → 2s → … capped at 30s) so
 *   brief network blips, backend restarts, or LB idle timeouts self-heal.
 * - Single-flight connect: repeated connect() calls while connected are no-ops.
 * - A tab becoming visible again retries immediately instead of waiting out
 *   the backoff, which is what makes background-tab sessions feel live.
 */
export class RealtimeClient {
  private ws: WebSocket | null = null
  private handlers = new Set<RealtimeHandler>()
  private retryTimer: ReturnType<typeof setTimeout> | null = null
  private backoffMs = INITIAL_BACKOFF_MS
  private manuallyClosed = false
  // Set while a connection is being established (token refresh is async), so
  // rapid connect() calls can't open two sockets.
  private connecting = false

  /** Connect (or reconnect) with the stored token. */
  connect(): void {
    if (this.connecting) return
    if (this.ws && this.ws.readyState <= WebSocket.OPEN) return
    // connect() is an explicit intent to (re)establish a connection — e.g.
    // after logout → login. Clear the latch a prior disconnect() set and drop
    // any pending backoff timer (a fresh attempt supersedes it).
    this.manuallyClosed = false
    if (this.retryTimer) {
      clearTimeout(this.retryTimer)
      this.retryTimer = null
    }

    this.connecting = true
    // Reconnect attempts must never spin against an expired access token:
    // refresh it first so the handshake (which carries it as ?token=) succeeds.
    void ensureFreshAccessToken()
      .then((token) => {
        this.connecting = false
        // Without a token there is nothing to connect with — and a failed
        // refresh has already cleared the session. Never schedule retries here
        // (that would spin forever); the hook reconnects when auth changes.
        if (!token) return
        // A disconnect() landed while the refresh was in flight: honor it.
        if (this.manuallyClosed) return

        const ws = new WebSocket(wsUrl(token))
        this.ws = ws

        ws.onopen = () => {
          this.backoffMs = INITIAL_BACKOFF_MS
        }

        ws.onmessage = (msg: MessageEvent) => {
          try {
            const data = JSON.parse(String(msg.data)) as RealtimeEvent
            if (data && typeof data.event === 'string') {
              this.handlers.forEach((handler) => handler(data))
            }
          } catch {
            // Ignore malformed frames; the connection stays up.
          }
        }

        ws.onclose = () => {
          if (this.ws === ws) this.ws = null
          if (!this.manuallyClosed) this.scheduleRetry()
        }

        ws.onerror = () => {
          // onclose fires right after; nothing extra to do.
        }
      })
      // ensureFreshAccessToken is not expected to reject, but if storage access
      // throws (private browsing, storage disabled) the flag must not stick.
      .catch(() => {
        this.connecting = false
      })
  }

  /** Subscribe to all realtime events. Returns an unsubscribe function. */
  subscribe(handler: RealtimeHandler): () => void {
    this.handlers.add(handler)
    return () => {
      this.handlers.delete(handler)
    }
  }

  /** Close the socket and stop reconnecting (logout / unmount). */
  disconnect(): void {
    this.manuallyClosed = true
    if (this.retryTimer) {
      clearTimeout(this.retryTimer)
      this.retryTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.handlers.clear()
  }

  /** Try to reconnect immediately (e.g. when the tab becomes visible). */
  reconnectNow(): void {
    if (this.retryTimer) {
      clearTimeout(this.retryTimer)
      this.retryTimer = null
    }
    if (this.manuallyClosed) return
    this.connect()
  }

  private scheduleRetry(): void {
    if (this.manuallyClosed || this.retryTimer) return
    this.retryTimer = setTimeout(() => {
      this.retryTimer = null
      this.connect()
    }, this.backoffMs)
    this.backoffMs = Math.min(this.backoffMs * 2, MAX_BACKOFF_MS)
  }
}

/** App-wide singleton so every page shares one socket. */
export const realtime = new RealtimeClient()
