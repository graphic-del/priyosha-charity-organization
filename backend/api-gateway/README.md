# api-gateway

Single entry point for the Priyosha backend, on port **8080**.

## Routes

| Route | Forwards to |
|---|---|
| `GET /health` | itself |
| `GET /api/stats` | content-service `/stats` |
| `GET /api/programs` | content-service `/programs` |
| `GET /api/pledges` | donation-service `/pledges` — **admin only** (`X-Admin-Token`, 403 without it) |
| `POST /api/pledges` | donation-service `/pledges` (public) |
| `GET /api/messages` | contact-service `/messages` — **admin only** (`X-Admin-Token`, 403 without it) |
| `POST /api/messages` | contact-service `/messages` (public) |

It also adds CORS headers so the Next.js frontend can call it from the browser.
CORS allows exactly one origin: `ALLOWED_ORIGIN`. Local default is
`http://localhost:3000`. In production (`ENV=production` or
`NODE_ENV=production`) the gateway refuses to start if `ALLOWED_ORIGIN` is
empty or `*`.

Admin lists are enforced here **and** in the donation/contact services, all
using the same `ADMIN_TOKEN` env var. The gateway forwards the
`X-Admin-Token` header downstream. If `ADMIN_TOKEN` is unset, admin GETs
return 403 everywhere.

## Env vars

| Var | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Listen port |
| `ENV` / `NODE_ENV` | `development` | `production` enables the CORS fail-fast check |
| `CONTENT_SERVICE_URL` | `http://content-service:8083` | Content service base URL |
| `DONATION_SERVICE_URL` | `http://donation-service:8081` | Donation service base URL |
| `CONTACT_SERVICE_URL` | `http://contact-service:8082` | Contact service base URL |
| `ALLOWED_ORIGIN` | `http://localhost:3000` | CORS origin — exactly your frontend URL. Refuses to start in production if empty or `*` |
| `ADMIN_TOKEN` | *(empty)* | Shared secret for admin GETs. If empty, admin GETs return 403 |

## Run locally

```bash
CONTENT_SERVICE_URL=http://localhost:8083 \
DONATION_SERVICE_URL=http://localhost:8081 \
CONTACT_SERVICE_URL=http://localhost:8082 \
go run .
```

Stdlib only — no external dependencies.
