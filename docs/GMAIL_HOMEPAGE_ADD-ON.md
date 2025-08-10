Perfect — I’ll put together **GMAIL\_HOMEPAGE\_ADD-ON.md** as a clean, distraction-free but complete implementation guide, including:

* **Overview & Scope**
* **Google Cloud Console Setup** for a Gmail Homepage add-on
* **Manifest configuration** (with working homepage + sidebar setup)
* **Cloudflare Workers deployment** steps (including gotchas we hit)
* **Card JSON example** (yours, formatted)
* **Troubleshooting notes** only where they help someone replicate success
* **POC-only caveats** clearly marked so they can be revisited later

I’ll also make sure it’s organized so you can hand it to someone else and they can follow it without the context of our chats.

---

Here’s the completed side document:

---

# GMAIL\_HOMEPAGE\_ADD-ON.md

**Gmail Homepage Add-on — Technical Implementation Guide**
*(Proof-of-Concept Deployment with Google Workspace & Cloudflare Workers)*

---

## 1. Overview

This guide documents the full implementation process for the Gmail Homepage add-on proof-of-concept (POC).
It covers both **Google Cloud Console configuration** and **Cloudflare Workers deployment**, along with the **manifest**, **card JSON**, and key troubleshooting notes that helped us get from zero to a functioning homepage add-on.

The intent is to make it possible to reproduce or extend the work without needing to refer to prior conversations.

---

## 2. Google Cloud Console Setup

### 2.1 Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/).
2. Click **Select a project → New Project**.
3. Name the project (e.g., `gmover-homepage-addon`).
4. Note the **Project ID** — you’ll need it later.

### 2.2 Enable APIs

Enable the following APIs in **APIs & Services → Library**:

* **Gmail API**
* **Google Workspace Add-ons API**

### 2.3 Configure OAuth Consent Screen

1. Go to **APIs & Services → OAuth consent screen**.
2. Set **User Type** to **Internal** (POC in Workspace environment) or **External** if distributing more widely.
3. Fill in:

    * **App name** (e.g., `GMover Inbox Dashboard`)
    * **User support email**
    * **Developer contact email**
4. Add **Scopes**:

    * `https://www.googleapis.com/auth/gmail.readonly` *(for reading message counts)*
    * `https://www.googleapis.com/auth/script.container.ui` *(for rendering cards in Gmail)*
5. Save.

### 2.4 Create OAuth Credentials

1. Go to **APIs & Services → Credentials**.
2. Click **Create Credentials → OAuth client ID**.
3. Choose **Web application**.
4. Add **Authorized redirect URIs** for deployment target — in our case, Cloudflare Workers URL:

   ```
   https://<your-worker-subdomain>.workers.dev
   ```
5. Save and download the `credentials.json`.

---

## 3. Add-on Manifest Configuration

Create a manifest JSON file describing the homepage add-on.

**manifest.json**

```json
{
  "oauthScopes": [
    "https://www.googleapis.com/auth/gmail.readonly",
    "https://www.googleapis.com/auth/script.container.ui"
  ],
  "addOns": {
    "common": {
      "name": "GMover Inbox Dashboard",
      "logoUrl": "https://example.com/logo.png"
    },
    "gmail": {
      "homepageTrigger": {
        "runFunction": "buildHomepageCard"
      },
      "contextualTriggers": [],
      "universalActions": []
    }
  }
}
```

**Key points**:

* `"homepageTrigger"` ensures the card appears on Gmail's **homepage tab**.
* No `contextualTriggers` are defined here — this POC is homepage-only.
* `runFunction` must correspond to an exported function in your backend handler.

---

## 4. Cloudflare Workers Setup

### 4.1 Create Worker Project

```bash
npm create cloudflare@latest gmail-homepage-addon
cd gmail-homepage-addon
```

### 4.2 Implement Worker Script

The Worker responds to Gmail add-on requests with card JSON.

Example:

```js
export default {
  async fetch(request, env) {
    return new Response(JSON.stringify(cardResponse()), {
      headers: { "Content-Type": "application/json" }
    });
  }
};
```

### 4.3 Deploy to Cloudflare

```bash
npx wrangler publish
```

Deployment will give you a URL like:

```
https://gmover-homepage-addon.yourname.workers.dev
```

---

## 5. Card JSON Example

From our working POC:

```js
function cardResponse(counts, appUrl) {
  return {
    header: { title: "GMover Inbox Dashboard" },
    sections: [
      {
        collapsible: true,
        uncollapsibleWidgetsCount: 1,
        widgets: [
          {
            decoratedText: {
              startIcon: {
                iconUrl: "https://raw.githubusercontent.com/SimonScholz/google-chat-action/refs/heads/main/assets/cancelled-128.png"
              },
              text: "<b>5 Time-sensitive Emails</b>",
              topLabel: "Unread & Urgent",
              bottomLabel: "Tap to review",
              onClick: {
                openLink: { url: appUrl + "/time-sensitive" }
              }
            }
          }
        ]
      }
    ]
  };
}
```

---

## 6. Troubleshooting Notes

### 6.1 Homepage Card Not Showing

* Ensure **homepageTrigger** exists in manifest under `"gmail"`.
* Verify your deployment URL is **publicly accessible** (Cloudflare Worker must be deployed).
* Check **OAuth consent screen** is **Published** (not in draft).

### 6.2 Sidebar Not Opening

* Gmail distinguishes between homepage cards and sidebar contextual add-ons.
* Sidebar display requires a `UniversalAction` or `contextualTrigger` to open it — not covered in this POC, but the backend must handle those requests.

### 6.3 Cloudflare CORS

* Gmail add-on requests come from Google servers, so CORS headers in Worker must allow Google origins:

```js
headers: {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "POST, GET, OPTIONS"
}
```

---

## 7. POC-Only Caveats

* All counts in `cardResponse` were **hardcoded** for demonstration.
* No OAuth token exchange implemented in Worker yet — API calls to Gmail will require service authentication.
* External image URLs used for icons; in production, host them reliably.

---

Do you want me to now **append a full sequence diagram** showing the Gmail client, Google servers, and Cloudflare Worker interactions for this homepage add-on? That would make the flow even clearer for future reference.
