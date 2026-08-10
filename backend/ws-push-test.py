"""End-to-end WebSocket push test against the running stack (port 3000).

Connects with a standards-compliant client, then creates a workflow +
template and submits an application, expecting the saga to push
`application_submitted` over the socket.

Pins the RFC 6455 accept-key computation explicitly so the check never depends
on quirks of the local websockets copy. Works against the docker-compose stack
(maps 3000 -> nginx) or any deployment behind a /api + /ws proxy.

Usage: python ws-push-test.py <email> <password>
"""
import asyncio
import base64
import hashlib
import json
import sys
import time
import urllib.request

import websockets.client

# RFC 6455 §1.3: the accept key is SHA1(key + GUID) base64-encoded. Pin it so
# the check is independent of any environment's websockets implementation.
def good_accept(key: str) -> str:
    GUID = "258EAFA5-E914-47DA-95CA-5AB9A8095E44"
    return base64.b64encode(hashlib.sha1((key + GUID).encode()).digest()).decode()

websockets.client.accept_key = good_accept

BASE = "http://localhost:3000/api/v1"


def api(path, token, payload):
    req = urllib.request.Request(
        BASE + path,
        data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"},
        method="POST",
    )
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())


async def main():
    email, password = sys.argv[1], sys.argv[2]
    login = api("/auth/login", None, {"email": email, "password": password})
    token = login["data"]["access_token"]
    user_id = login["data"]["user"]["id"]
    print("logged in as", user_id)

    # Unique names so repeated runs never collide on unique constraints.
    suffix = str(int(time.time()))
    wf = api("/workflows/create", token, {
        "name": f"WS Push Workflow {suffix}",
        "category": "smoke",
        "steps": [{"name": "Review"}],
    })
    workflow_id = wf["data"]["workflow"]["id"] if "workflow" in wf.get("data", {}) else wf["data"]["id"]
    tmpl = api("/templates/create", token, {"name": f"WS Push Template {suffix}", "category": "smoke"})
    template_id = tmpl["data"]["template"]["id"] if "template" in tmpl.get("data", {}) else tmpl["data"]["id"]
    print("workflow", workflow_id, "template", template_id)

    async with websockets.connect(f"ws://localhost:3000/ws?token={token}") as ws:
        print("WS connected")

        sub = api("/applications/submit", token, {
            "applicant_id": user_id,
            "workflow_id": workflow_id,
            "template_id": template_id,
            "title": f"WS push test application {suffix}",
            "priority": "low",
        })
        print("submit status:", sub.get("code"))

        # The saga broadcasts application_submitted to all connected clients.
        try:
            msg = await asyncio.wait_for(ws.recv(), timeout=5)
            parsed = json.loads(msg)
            print("PUSH RECEIVED:", parsed.get("event"))
            print("PASS" if parsed.get("event") == "application_submitted" else "CHECK EVENT")
        except asyncio.TimeoutError:
            print("NO PUSH within 5s")
        await ws.close()


asyncio.run(main())
