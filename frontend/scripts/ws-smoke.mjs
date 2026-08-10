// Real-browser WebSocket smoke test against the running stack (port 3000).
// Usage: node scripts/ws-smoke.mjs <adminEmail> <adminPassword>
import { chromium } from 'playwright'

const [email, password] = process.argv.slice(2)
if (!email || !password) {
  console.error('usage: node scripts/ws-smoke.mjs <email> <password>')
  process.exit(1)
}

const suffix = Date.now()
const browser = await chromium.launch()
const page = await browser.newPage()
page.on('console', (m) => {
  if (m.type() === 'error') console.log('[browser console.error]', m.text().slice(0, 160))
})

await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' })

// Phase 0: login + open a browser WebSocket, exactly like the app does.
const boot = await page.evaluate(async ({ email, password }) => {
  const login = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  const body = await login.json()
  if (!body?.data?.access_token) return { loginStatus: login.status, body: JSON.stringify(body).slice(0, 200) }
  const token = body.data.access_token
  const userId = body.data.user.id

  const state = { events: [] }
  const ws = new WebSocket(`ws://${window.location.host}/ws?token=${encodeURIComponent(token)}`)
  window.__wsState = state
  ws.onopen = () => state.events.push('open')
  ws.onmessage = (m) => state.events.push(`msg:${String(m.data).slice(0, 160)}`)
  ws.onerror = () => state.events.push('error')
  ws.onclose = (e) => state.events.push(`close:${e.code}`)
  return { loginStatus: login.status, token, userId }
}, { email, password })
console.log('login:', boot.loginStatus, 'userId:', boot.userId)

if (!boot.token) process.exit(1)

// Phase 1: the socket must stay open (no premature close after the handshake).
await page.waitForTimeout(4000)
let state = await page.evaluate(() => window.__wsState.events)
console.log('after 4s idle:', JSON.stringify(state))
if (!state.includes('open') || state.some((e) => e.startsWith('close'))) {
  console.log('FAIL - socket did not stay open')
  await browser.close()
  process.exit(1)
}

// Phase 2: trigger a real push. Create a workflow + template, then submit an
// application as this user; the saga broadcasts application_submitted.
const push = await page.evaluate(async ({ email, password, userId, token, suffix }) => {
  const auth = { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }
  const j = async (r) => ({ status: r.status, body: await r.json() })

  const wf = await j(await fetch('/api/v1/workflows/create', {
    method: 'POST', headers: auth,
    body: JSON.stringify({ name: `WS Smoke Workflow ${suffix}`, category: 'smoke', steps: [{ name: 'Review' }] }),
  }))
  const workflowId = wf.body?.data?.workflow?.id ?? wf.body?.data?.id

  const tmpl = await j(await fetch('/api/v1/templates/create', {
    method: 'POST', headers: auth,
    body: JSON.stringify({ name: `WS Smoke Template ${suffix}`, category: 'smoke' }),
  }))
  const templateId = tmpl.body?.data?.template?.id ?? tmpl.body?.data?.id

  const sub = await j(await fetch('/api/v1/applications/submit', {
    method: 'POST', headers: auth,
    body: JSON.stringify({
      applicant_id: userId,
      workflow_id: workflowId,
      template_id: templateId,
      title: `WS smoke test application ${suffix}`,
      priority: 'low',
    }),
  }))
  return { wfStatus: wf.status, templateStatus: tmpl.status, subStatus: sub.status,
           workflowId, templateId, subBody: JSON.stringify(sub.body).slice(0, 160) }
}, { email, password, userId: boot.userId, token: boot.token, suffix })
console.log('push chain:', JSON.stringify(push))

await page.waitForTimeout(3000)
state = await page.evaluate(() => window.__wsState.events)
console.log('after submit:', JSON.stringify(state))

const receivedPush = state.some((e) => e.startsWith('msg:'))
const stillOpen = !state.some((e) => e.startsWith('close'))
console.log(stillOpen && receivedPush ? 'PASS - socket open + live push received' : 'CHECK - see events above')
await browser.close()
