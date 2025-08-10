import icon from "./icon.png";
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
		header: { title: "GMover Inbox Dashboard" },
		sections: [
			{
				"collapsible": true,
				"uncollapsibleWidgetsCount": 1,
				"widgets": [
					{
						"decoratedText": {
							"startIcon": { "iconUrl": "https://raw.githubusercontent.com/SimonScholz/google-chat-action/refs/heads/main/assets/cancelled-128.png" },
							"text": "<b>5 Time-sensitive Emails</b>",
							"topLabel": "Needs attention"
						} },
					{ "decoratedText": { "text": "Email 1 subject" } },
					{ "decoratedText": { "text": "Email 2 subject" } },
					{ "decoratedText": { "text": "Email 3 subject" } },
					{ "decoratedText": { "text": "Email 4 subject" } },
					{ "decoratedText": { "text": "Email 5 subject" } },
					{ "decoratedText": { "text": "Email 6 subject" } },
					{ "decoratedText": { "text": "Email 7 subject" } }
				]
			},
			{ widgets: [ { decoratedText: { topLabel: "Deals", text: "0" } } ] },
			{ widgets: [ { decoratedText: { topLabel: "Events", text: "0" } } ] },
			{ widgets: [ { decoratedText: { topLabel: "Newsletters", text: "0" } } ] },
			{
				widgets: [
					{
						buttonList: {
							buttons: [
								{
									text: "Open full dashboard",
									onClick: { openLink: { url: "https://gmail-dashboard-worker.mike-7ae.workers.dev/" } }
								}
							]
						}
					}
				]
			}
		]
	};
}

function htmlPlaceholder(selfUrl: string): Response {
	const html = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Inbox Dashboard (PoC)</title><style>body{font:16px system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,Cantarell,Noto Sans,sans-serif;margin:0;padding:2rem;line-height:1.45}.card{max-width:680px;margin:auto;padding:1.25rem;border:1px solid #e5e7eb;border-radius:12px;box-shadow:0 4px 20px rgba(0,0,0,.06)}h1{font-size:1.4rem;margin:.2rem 0 1rem}code{background:#f8fafc;border:1px solid #e5e7eb;border-radius:6px;padding:.15rem .35rem}.row{display:flex;gap:1rem;flex-wrap:wrap}.pill{padding:.4rem .7rem;border-radius:999px;background:#f1f5f9;border:1px solid #e2e8f0}</style><div class="card"><h1>Inbox Dashboard (PoC)</h1><p>Gmail add-on calls <code>POST /homepage</code> to render cards. You can host the full dashboard UI here later.</p><div class="row"><span class="pill">Worker: <code>${selfUrl}</code></span><span class="pill">Endpoint: <code>/homepage</code></span></div><p>Test:</p><pre><code>curl -X POST ${selfUrl}/homepage -H 'content-type: application/json' -d '{}'</code></pre></div></html>`;
	return new Response(html, { headers: { "content-type": "text/html; charset=utf-8" } });
}

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		const url = new URL(request.url);
		const method = request.method.toUpperCase();

		if (method === "GET" && (url.pathname === "/icon.png")) {
			return new Response(icon, {
				headers: { "content-type": "image/png" }
			});
		}
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
						const data: { counts?: Record<string, number> } = await r.json();
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
