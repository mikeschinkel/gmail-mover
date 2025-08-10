# Gmail “Inbox Dashboard” — Proof of Concept (PoC)

**Scope:** This document captures **temporary, PoC-only** choices and step-by-step instructions to stand up a working Gmail dashboard add-on quickly. Long-term/committed decisions live in ADR-008.

* **PoC runtime host:** Cloudflare Worker (✅ for PoC; **TBD** for long-term)
* **UI contract:** Return **Google Card Framework JSON** from HTTP endpoint (✅ PoC and **likely final**)
* **Gmail scopes:** Minimal — `gmail.addons.execute` + `userinfo.email` (✅ PoC)
* **Token verification:** **Disabled** for PoC (flip on for prod)

---

## 0) What we’re building

A Gmail Add-on (HTTP runtime) whose **Homepage** renders a dashboard of widgets. The add-on calls a **Cloudflare Worker** at `POST /homepage`. The Worker returns **Card JSON** and (optionally) calls our backend API for counts.

---

## 1) Prereqs

* Node 18+
* Cloudflare account + `wrangler`
* A Google Cloud project where you can enable **Google Workspace Marketplace SDK** (for HTTP deployments)

---

## 2) Create the Worker (TypeScript, “Worker only”)

```bash
npm create cloudflare@latest gmail-dashboard-worker
# Choose: Worker only → TypeScript
cd gmail-dashboard-worker
npm i jose
```

**Project layout**

```
./
  wrangler.toml
  package.json
  src/index.ts    # we’ll replace with the PoC handler below
```

### 2.1 `wrangler.toml`

> Replace `<your-subdomain>` if different. You already have: `gmail-dashboard-worker.mike-7ae.workers.dev`.

```toml
name = "gmail-dashboard-worker"
main = "src/index.ts"
compatibility_date = "2025-08-08"

[vars]
SELF_URL = "https://gmail-dashboard-worker.mike-7ae.workers.dev"
# For PoC we’ll host a tiny HTML page at GET /
DASHBOARD_APP_URL = "https://gmail-dashboard-worker.mike-7ae.workers.dev/"
# Leave blank to use stub counts; set later to your backend like https://api.example.com
DASHBOARD_API = ""
# Filled in later from the Marketplace HTTP deployment page (not required while VERIFY_TOKENS=false)
OAUTH_CLIENT_ID = ""
# PoC: no verification; set to "true" in prod
VERIFY_TOKENS = "false"
# Optional: audience override if you verify the system token later
SYSTEM_TOKEN_AUD = ""
```

### 2.2 `package.json` (excerpt)

```json
{
  "name": "gmail-dashboard-worker",
  "type": "module",
  "scripts": { "dev": "wrangler dev", "deploy": "wrangler deploy" },
  "dependencies": { "jose": "^5.9.2" },
  "devDependencies": { "wrangler": "^3.75.0", "typescript": "^5.5.4" }
}
```

### 2.3 `src/index.ts`

Paste this entire file:

```ts
import { createRemoteJWKSet, jwtVerify, JWTPayload } from "jose";

export interface Env {
  SELF_URL: string;
  DASHBOARD_APP_URL: string;
  DASHBOARD_API: string;
  OAUTH_CLIENT_ID: string;
  VERIFY_TOKENS: string; // "true" to enforce token checks
  SYSTEM_TOKEN_AUD?: string;
}

const GOOGLE_JWKS = createRemoteJWKSet(new URL("https://www.googleapis.com/oauth2/v3/certs"));

async function verifyIdToken(token: string, audience: string | string[]): Promise<JWTPayload> {
  const { payload } = await jwtVerify(token, GOOGLE_JWKS, {
    audience,
    issuer: ["https://accounts.google.com", "accounts.google.com"],
  });
  return payload;
}

function cardResponse(counts: Record<string, number>, appUrl: string) {
  return {
    renderActions: {
      action: {
        navigations: [
          {
            pushCard: {
              header: { title: "Inbox Dashboard" },
              sections: [
                { widgets: [{ decoratedText: { topLabel: "Time-sensitive", text: String(counts.time_sensitive ?? 0) } }] },
                { widgets: [{ decoratedText: { topLabel: "Deals",          text: String(counts.deals ?? 0) } }] },
                { widgets: [{ decoratedText: { topLabel: "Events",         text: String(counts.events ?? 0) } }] },
                { widgets: [{ decoratedText: { topLabel: "Newsletters",    text: String(counts.newsletters ?? 0) } }] },
                {
                  widgets: [
                    {
                      buttonList: {
                        buttons: [
                          { text: "Open full dashboard", onClick: { openLink: { url: appUrl } } }
                        ]
                      }
                    }
                  ]
                }
              ]
            }
          }
        ]
      }
    }
  } as const;
}

function htmlPlaceholder(selfUrl: string): Response {
  const html = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Inbox Dashboard (PoC)</title><style>body{font:16px system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,Cantarell,Noto Sans,sans-serif;margin:0;padding:2rem;line-height:1.45}.card{max-width:680px;margin:auto;padding:1.25rem;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 4px 20px rgba(0,0,0,.06)}h1{font-size:1.4rem;margin:.2rem 0 1rem}code{background:#f8fafc;border:1px solid #e5e7eb;border-radius:6px;padding:.15rem .35rem}.row{display:flex;gap:1rem;flex-wrap:wrap}.pill{padding:.4rem .7rem;border-radius:999px;background:#f1f5f9;border:1px solid #e2e8f0}</style><div class="card"><h1>Inbox Dashboard (PoC)</h1><p>Gmail add-on calls <code>POST /homepage</code> to render cards. You can host the full dashboard UI here later.</p><div class="row"><span class="pill">Worker: <code>${selfUrl}</code></span><span class="pill">Endpoint: <code>/homepage</code></span></div><p>Test:</p><pre><code>curl -X POST ${selfUrl}/homepage -H 'content-type: application/json' -d '{}'</code></pre></div></html>`;
  return new Response(html, { headers: { "content-type": "text/html; charset=utf-8" } });
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const method = request.method.toUpperCase();

    if (method === "GET" && (url.pathname === "/" || url.pathname === "/index.html")) {
      return htmlPlaceholder(env.SELF_URL || `${url.origin}`);
    }

    if (method === "POST" && url.pathname === "/homepage") {
      let event: any = {};
      try { event = await request.json(); } catch {}

      const verify = (env.VERIFY_TOKENS || "false").toLowerCase() === "true";
      const authObj = event?.authorizationEventObject || {};

      if (verify) {
        const authz = request.headers.get("authorization") || "";
        const systemToken = authz.startsWith("Bearer ") ? authz.slice(7) : "";
        if (!systemToken) return new Response("Missing bearer token", { status: 401 });
        try {
          const expectedAud = env.SYSTEM_TOKEN_AUD || `${url.origin}${url.pathname}`;
          await verifyIdToken(systemToken, expectedAud);
        } catch {
          return new Response("Invalid system token", { status: 401 });
        }
      }

      let userEmail = "unknown";
      if (verify && authObj.userIdToken && env.OAUTH_CLIENT_ID) {
        try {
          const u = await verifyIdToken(authObj.userIdToken, env.OAUTH_CLIENT_ID);
          userEmail = (u.email as string) || (u.sub as string) || "unknown";
        } catch {}
      }

      let counts = { time_sensitive: 0, deals: 0, events: 0, newsletters: 0 };
      if (env.DASHBOARD_API) {
        try {
          const r = await fetch(`${env.DASHBOARD_API.replace(/\/$/, "")}/snapshot`, {
            method: "POST",
            headers: { "content-type": "application/json" },
            body: JSON.stringify({ user: userEmail })
          });
          if (r.ok) {
            const data = await r.json();
            if (data?.counts && typeof data.counts === "object") {
              counts = { ...counts, ...data.counts };
            }
          }
        } catch {}
      }

      return new Response(JSON.stringify(cardResponse(counts, env.DASHBOARD_APP_URL || `${url.origin}/`)), {
        headers: { "content-type": "application/json; charset=utf-8" }
      });
    }

    return new Response("Not found", { status: 404 });
  }
};
```

### 2.4 Run & Deploy

```bash
npm run dev
# in another terminal
curl -sS -X POST http://127.0.0.1:8787/homepage -H 'content-type: application/json' -d '{}' | jq

npm run deploy
# Test the deployed endpoint
curl -sS -X POST https://gmail-dashboard-worker.mike-7ae.workers.dev/homepage \
  -H 'content-type: application/json' -d '{}' | jq
```

---

## 3) Create the Gmail Add-on (HTTP deployment)

In **Google Cloud Console → Google Workspace Marketplace SDK → HTTP Deployments → Create deployment** and use:

```json
{
  "oauthScopes": [
    "https://www.googleapis.com/auth/gmail.addons.execute",
    "https://www.googleapis.com/auth/userinfo.email"
  ],
  "addOns": {
    "common": {
      "name": "Inbox Dashboard (PoC)",
      "logoUrl": "https://gmail-dashboard-worker.mike-7ae.workers.dev/icon.png",
      "homepageTrigger": {
        "runFunction": "https://gmail-dashboard-worker.mike-7ae.workers.dev/homepage"
      }
    },
    "gmail": {},
    "httpOptions": { "granularOauthPermissionSupport": "OPT_IN" }
  }
}
```

**Install** the deployment to your account (developer install). Open Gmail and click your add-on’s icon in the right panel; the Homepage should render.

---

## 4) Getting the `OAUTH_CLIENT_ID` (for later token verification)

* On the **HTTP Deployment** detail page (Marketplace SDK), copy the **OAuth client ID** listed under **Authorization**.
* Paste that into `wrangler.toml` → `OAUTH_CLIENT_ID`.
* End users **do not** create client IDs. We (publisher) own this configuration. For environments, we may use separate deployments (and client IDs) for dev/test/prod.

> PoC runs fine without this because `VERIFY_TOKENS=false`.

---

## 5) Hooking real data (optional in PoC)

* Set `DASHBOARD_API` to your backend: e.g., `https://api.example.com`.
* Implement `POST /snapshot` on the backend to return:

```json
{ "counts": { "time_sensitive": 3, "deals": 5, "events": 2, "newsletters": 8 } }
```

---

## 6) What we are **not** doing in the PoC

* No restricted Gmail scopes.
* No token verification (yet).
* No compose/context actions.
* No storage (KV/D1) — we assume the backend provides data.

---

## 7) Turning on verification (post-PoC)

* Set `VERIFY_TOKENS=true` and ensure:

	* `OAUTH_CLIENT_ID` is set (verify `userIdToken`).
	* Optionally set `SYSTEM_TOKEN_AUD` to the endpoint URL used by Gmail for the system token audience.

---

## 8) Quick Troubleshooting

* **Gmail shows blank sidebar:** Check Worker logs; ensure `POST /homepage` is reachable and returns JSON.
* **CORS:** Not applicable — Google calls server-to-server.
* **403/401 with verification on:** Audience/issuer mismatches; print the decoded header/payload in logs during setup.

---

## 9) Curl snippets

```bash
# Local dev
curl -X POST http://127.0.0.1:8787/homepage -H 'content-type: application/json' -d '{}'

# Prod URL
curl -X POST https://gmail-dashboard-worker.mike-7ae.workers.dev/homepage \
  -H 'content-type: application/json' -d '{}'
```

---

**End — PoC doc (to be discarded/revised once we lock long-term runtime)**
