import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { RealtimeClient } from './realtime'
import { storage } from '@/utils/storage'

// The client refreshes the token before every handshake; make it resolve
// synchronously with whatever token is in storage so tests stay deterministic.
vi.mock('@/services/api/client', () => ({
  ensureFreshAccessToken: () => Promise.resolve(storage.getAccessToken()),
}))

/** Minimal fake WebSocket that records constructor args and lets tests drive it. */
class FakeWebSocket {
  static instances: FakeWebSocket[] = []
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  url: string
  readyState = FakeWebSocket.CONNECTING
  onopen: (() => void) | null = null
  onmessage: ((msg: { data: string }) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null
  closeCalls = 0

  constructor(url: string) {
    this.url = url
    FakeWebSocket.instances.push(this)
  }

  close(): void {
    this.closeCalls++
    this.readyState = FakeWebSocket.CLOSED
  }

  open() {
    this.readyState = FakeWebSocket.OPEN
    this.onopen?.()
  }

  receive(data: unknown) {
    this.onmessage?.({ data: JSON.stringify(data) })
  }

  fail() {
    this.onerror?.()
    this.onclose?.()
  }
}

describe('RealtimeClient', () => {
  let client: RealtimeClient

  beforeEach(() => {
    vi.useFakeTimers()
    FakeWebSocket.instances = []
    vi.stubGlobal('WebSocket', FakeWebSocket)
    storage.clearAuth()
    client = new RealtimeClient()
  })

  /** The token refresh is async, so flush microtasks after connect(). */
  async function connectAndSettle() {
    client.connect()
    await vi.advanceTimersByTimeAsync(0)
  }

  afterEach(() => {
    client.disconnect()
    vi.unstubAllGlobals()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('connects with the stored token in the URL', async () => {
    storage.setAccessToken('tok-123')
    await connectAndSettle()

    expect(FakeWebSocket.instances).toHaveLength(1)
    expect(FakeWebSocket.instances[0].url).toContain('/ws?token=tok-123')
  })

  it('does not connect or retry when there is no session token', async () => {
    await connectAndSettle()
    expect(FakeWebSocket.instances).toHaveLength(0)
    // No retry loop: with no token there is nothing to reconnect with — the
    // hook reconnects when auth state changes instead.
    await vi.advanceTimersByTimeAsync(60_000)
    expect(FakeWebSocket.instances).toHaveLength(0)
  })

  it('is a no-op when already connected', async () => {
    storage.setAccessToken('tok-1')
    await connectAndSettle()
    FakeWebSocket.instances[0].open()

    await connectAndSettle()
    expect(FakeWebSocket.instances).toHaveLength(1)
  })

  it('dispatches parsed events to subscribers and unsubscribes', async () => {
    storage.setAccessToken('tok-1')
    const handler = vi.fn()
    const unsubscribe = client.subscribe(handler)
    await connectAndSettle()

    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ event: 'approval_needed', payload: { approval_id: 'a1' } })
    expect(handler).toHaveBeenCalledWith({
      event: 'approval_needed',
      payload: { approval_id: 'a1' },
    })

    unsubscribe()
    ws.receive({ event: 'notification', payload: {} })
    expect(handler).toHaveBeenCalledTimes(1)
  })

  it('ignores malformed frames without dropping the connection', async () => {
    storage.setAccessToken('tok-1')
    const handler = vi.fn()
    client.subscribe(handler)
    await connectAndSettle()

    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.onmessage?.({ data: 'not json' })
    expect(handler).not.toHaveBeenCalled()
  })

  it('reconnects with exponential backoff after a drop', async () => {
    storage.setAccessToken('tok-1')
    await connectAndSettle()
    const ws1 = FakeWebSocket.instances[0]
    ws1.open()

    // A successful connection resets the backoff to 1s: after a drop, the
    // next reconnect attempt happens after 1s.
    ws1.fail()
    expect(FakeWebSocket.instances).toHaveLength(1)
    await vi.advanceTimersByTimeAsync(1_000)
    expect(FakeWebSocket.instances).toHaveLength(2)

    // The new attempt fails while still CONNECTING (no successful open yet),
    // so the backoff doubles for the next attempt — 2s this time.
    FakeWebSocket.instances[1].fail()
    await vi.advanceTimersByTimeAsync(1_000)
    expect(FakeWebSocket.instances).toHaveLength(2)
    await vi.advanceTimersByTimeAsync(1_000)
    expect(FakeWebSocket.instances).toHaveLength(3)
  })

  it('caps backoff at 30s', async () => {
    storage.setAccessToken('tok-1')
    await connectAndSettle()

    // Cycle through drops until backoff is saturated.
    let drops = 0
    while (drops < 10) {
      const ws = FakeWebSocket.instances[FakeWebSocket.instances.length - 1]
      if (ws.readyState === FakeWebSocket.CONNECTING) {
        ws.open()
      }
      ws.fail()
      drops++
      await vi.advanceTimersByTimeAsync(35_000)
    }
    // Backoff never grows beyond 30s: after saturating, a drop still
    // reconnects within 30s (not 60s+).
    const before = FakeWebSocket.instances.length
    const ws = FakeWebSocket.instances[FakeWebSocket.instances.length - 1]
    if (ws.readyState === FakeWebSocket.CONNECTING) ws.open()
    ws.fail()
    await vi.advanceTimersByTimeAsync(30_000)
    expect(FakeWebSocket.instances.length).toBe(before + 1)
  })

  it('does not reconnect after explicit disconnect', async () => {
    storage.setAccessToken('tok-1')
    await connectAndSettle()
    const ws = FakeWebSocket.instances[0]
    ws.open()

    client.disconnect()
    expect(ws.closeCalls).toBe(1)

    ws.fail()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(FakeWebSocket.instances).toHaveLength(1)
  })

  it('reconnects after disconnect once a token is available again (logout → login)', async () => {
    storage.setAccessToken('tok-1')
    await connectAndSettle()
    FakeWebSocket.instances[0].open()

    client.disconnect()
    expect(FakeWebSocket.instances).toHaveLength(1)

    // A fresh login makes the hook call connect() again; the disconnect latch
    // must not leave the socket dead forever.
    client.connect()
    await vi.advanceTimersByTimeAsync(0)
    expect(FakeWebSocket.instances).toHaveLength(2)
    expect(FakeWebSocket.instances[1].url).toContain('/ws?token=tok-1')
  })

  it('honors a disconnect that lands while the token refresh is in flight', async () => {
    storage.setAccessToken('tok-1')
    client.connect() // refresh pending, socket not yet created
    client.disconnect()
    await vi.advanceTimersByTimeAsync(0)

    expect(FakeWebSocket.instances).toHaveLength(0)
    await vi.advanceTimersByTimeAsync(60_000)
    expect(FakeWebSocket.instances).toHaveLength(0)
  })

  it('reconnectNow retries immediately', async () => {
    storage.setAccessToken('tok-1')
    await connectAndSettle()
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.fail()

    client.reconnectNow()
    await vi.advanceTimersByTimeAsync(0)
    expect(FakeWebSocket.instances).toHaveLength(2)
  })
})
