# Priyosha Charity Organization — Deployable Website

Source code for the website of **Priyosha Charity Organization**, a US-based
nonprofit founded by **Md Shadman Mahmood**. It raises charity from around the
world to reduce problems in Bangladesh, aiming for better futures for humans
and animals — especially cats and dogs who suffer the most.

- **Frontend:** Next.js (App Router, TypeScript)
- **Backend:** Go (Golang), standard library only
- **Architecture:** Microservices behind an API gateway

## Architecture

```
                    ┌──────────────────────┐
  Visitor ─────────►│  Frontend (Next.js)  │  :3000
                    │  public website      │
                    └──────────┬───────────┘
                               │  NEXT_PUBLIC_API_URL
                               ▼
                    ┌──────────────────────┐
                    │  api-gateway (Go)    │  :8080  single entry point
                    └──┬───────┬───────┬───┘
                       │       │       │
              ┌────────▼┐ ┌────▼────┐ ┌▼──────────────┐
              │ content │ │ donation│ │ contact       │
              │ service │ │ service │ │ service       │
              │ :8083   │ │ :8081   │ │ :8082         │
              └─────────┘ └─────────┘ └───────────────┘
              stats +       pledges       messages
              programs      (JSON file)   (JSON file)
```

| Service | Port | Job |
|---|---|---|
| `frontend/` | 3000 | Next.js site: hero, mission, The Need, programs, pledge calculator, contact/volunteer form, founder |
| `backend/api-gateway/` | 8080 | Single entry point; routes `/api/*` to the other services; adds CORS (exact origin only) |
| `backend/content-service/` | 8083 | `GET /stats`, `GET /programs`, `GET /health` — the figures and programme data |
| `backend/donation-service/` | 8081 | `POST /pledges` (public), `GET /pledges` (**admin only** — `X-Admin-Token`), `GET /health` — records pledges |
| `backend/contact-service/` | 8082 | `POST /messages` (public), `GET /messages` (**admin only** — `X-Admin-Token`), `GET /health` — records contact/volunteer messages |

Each Go service is independent: its own `go.mod`, `main.go` (stdlib only),
`Dockerfile` and `README.md`. The frontend only ever talks to the gateway.

## The figures on the site

The Need section uses exactly these figures, with sources, plus the warning
that they cannot be added together:

- **18.1 million / 1.81 crore** people projected to face crisis-level food
  insecurity or worse during September–December 2026, within the areas
  assessed — Integrated Food Security Phase Classification (IPC)
- **22,185** people counted sleeping in public places in the 2022 census —
  narrow category (streets, stations, similar); does not capture everyone
  without secure housing; older count — Bangladesh Bureau of Statistics
  (nsds.bbs.gov.bd)
- **About 16.7 lakh dogs** estimated in a 2024 study — estimates the dog
  population; does not establish how many are homeless or hungry (PMC)
- **Hungry/homeless animals overall, including cats:** no reliable nationwide
  total found — research identifies gaps in Bangladesh's stray-animal
  population data; giving an exact count would be guesswork
  (pmc.ncbi.nlm.nih.gov)

> These figures cannot be added together: hunger and homelessness can affect
> the same people, and animal population estimates do not measure suffering.

To update a figure, edit the `stats` slice in
`backend/content-service/main.go` and redeploy that service only — the
frontend fetches the figures from the API.

## Run locally (without Docker)

You need Go 1.23+ and Node.js 18.17+.

```bash
# 1. Backend — one terminal per service, or run them in the background.
cd backend/content-service  && PORT=8083 go run .
cd backend/donation-service && PORT=8081 go run .
cd backend/contact-service  && PORT=8082 go run .
cd backend/api-gateway && \
  CONTENT_SERVICE_URL=http://localhost:8083 \
  DONATION_SERVICE_URL=http://localhost:8081 \
  CONTACT_SERVICE_URL=http://localhost:8082 \
  PORT=8080 go run .

# 2. Frontend
cd frontend
cp .env.example .env.local   # NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev                  # http://localhost:3000 (development)
# Production build: npm run build, then node .next/standalone/server.js
```

Check the backend:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/stats
curl -X POST http://localhost:8080/api/pledges \
  -H 'Content-Type: application/json' \
  -d '{"amount":50,"frequency":"monthly","name":"Test Donor","email":"test@example.org"}'
curl -X POST http://localhost:8080/api/messages \
  -H 'Content-Type: application/json' \
  -d '{"name":"Test Volunteer","email":"test@example.org","interest":"volunteer","message":"I would like to help."}'

# Admin lists — need a token. Start the services with the same ADMIN_TOKEN,
# e.g. ADMIN_TOKEN=test-admin-token, then:
curl http://localhost:8080/api/pledges
# -> 403 (no token)
curl -H 'X-Admin-Token: test-admin-token' http://localhost:8080/api/pledges
# -> 200, the recorded pledges
curl -H 'X-Admin-Token: test-admin-token' http://localhost:8080/api/messages
# -> 200, the recorded messages
```

## Run locally (with Docker)

Docker is the easiest way to run everything together. Docker was not
available in the build environment, so `docker compose up` was not run
there — the compose file was written to match the services and verified by
running the services directly (see below).

```bash
cp .env.example .env   # then fill in your real values for production;
                       # local dev works with the defaults as-is
docker compose up --build
# Frontend: http://localhost:3000
# Gateway:  http://localhost:8080/health
```

To point the frontend at a deployed gateway, set it in `.env`:

```bash
NEXT_PUBLIC_API_URL=https://api.priyoshacharity.com docker compose up --build
```

## GitHub

See [GITHUB_SETUP.md](GITHUB_SETUP.md) — the repo is initialised and committed
locally, nothing has been pushed. Create a repo named
`priyosha-charity-organization` under GitHub user `graphic-del`, then:

```bash
git remote add origin https://github.com/graphic-del/priyosha-charity-organization.git
git branch -M main
git push -u origin main
```

## Deploy — priyoshacharity.com

The live domain is **priyoshacharity.com** (primary site
`https://priyoshacharity.com`, API `https://api.priyoshacharity.com`).
`www.priyoshacharity.com` redirects to the bare domain (see
[deploy/Caddyfile](deploy/Caddyfile)).

One straightforward option (VPS / any host with Docker):

1. Add DNS A records pointing at your server's IP —
   see [deploy/README.md](deploy/README.md) for the exact records
   (`priyoshacharity.com`, `www.priyoshacharity.com` and
   `api.priyoshacharity.com`).
2. Copy this whole folder to the server (or `git clone` it from your repo).
3. Copy `.env.example` to `.env` and fill in only the secrets:
   `ADMIN_TOKEN` (a long random value — generate with
   `openssl rand -hex 32`), and `ENV=production`. The domain values are
   already set to priyoshacharity.com in `.env.example`. Never commit
   `.env` or a real token.
4. Run `docker compose up --build -d`.
5. Put Caddy in front of ports 3000 (website) and 8080 (API) using
   [deploy/Caddyfile](deploy/Caddyfile) — it gets HTTPS certificates
   automatically. Keep only ports 80/443 open to the internet.

Alternatives: deploy the frontend to Vercel (`frontend/`, set
`NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_SITE_URL` in the project settings)
and the Go services to any host that runs Docker or plain binaries
(`go build` in each `backend/*/`, then run the binary with the env vars in
its README).

## Environment variables

Copy `.env.example` to `.env` (never commit `.env`). All values are also
documented there.

| Var | Where | Default | Meaning |
|---|---|---|---|
| `DOMAIN` | compose / Caddy | `priyoshacharity.com` | Your domain, no `https://` prefix |
| `FRONTEND_URL` | compose / frontend | `https://priyoshacharity.com` | Full public URL of the website |
| `NEXT_PUBLIC_API_URL` | frontend | `http://localhost:8080` | Gateway URL the browser calls |
| `NEXT_PUBLIC_SITE_URL` | frontend | `FRONTEND_URL` | Canonical site URL for metadata / sitemap / robots |
| `ENV` | api-gateway | `development` | `development` or `production` |
| `PORT` | each Go service | 8080/8081/8082/8083 | Listen port |
| `CONTENT_SERVICE_URL` | api-gateway | `http://content-service:8083` | Content service base URL |
| `DONATION_SERVICE_URL` | api-gateway | `http://donation-service:8081` | Donation service base URL |
| `CONTACT_SERVICE_URL` | api-gateway | `http://contact-service:8082` | Contact service base URL |
| `ALLOWED_ORIGIN` | api-gateway | `http://localhost:3000` | CORS origin — exactly your frontend URL. In production the gateway **refuses to start** if this is empty or `*` |
| `ADMIN_TOKEN` | gateway, donation, contact | *(empty)* | Shared secret for `GET /pledges` and `GET /messages` (`X-Admin-Token` header). If empty, those GETs always return **403** |
| `DATA_FILE` | donation/contact services | `./data/pledges.json` / `./data/messages.json` | JSON file where records persist |

## What to replace before taking real donations

- **Admin lists are locked down, but need a token.** `GET /pledges` and
  `GET /messages` now require `X-Admin-Token: <ADMIN_TOKEN>` and return
  **403** without it (or when `ADMIN_TOKEN` is unset). Set the same long
  random `ADMIN_TOKEN` on the gateway, donation-service and
  contact-service (docker-compose.yml already passes it to all three) and
  share it only with the site owner. Read lists with `curl -H
  'X-Admin-Token: <token>' https://api.priyoshacharity.com/api/pledges`.
- **Real payments are NOT wired up.** A pledge is a record of intent, not a
  charge. To take payments, plug in Stripe (or another provider) in
  `backend/donation-service/main.go` — see the `TODO (real payments)` comment
  there. Never store card numbers in this service.
- **No email is sent yet.** Contact messages are only stored. Plug in an
  email provider (SendGrid, AWS SES, …) in
  `backend/contact-service/main.go` — see the `TODO (real email)` comment.
- **Contact details are placeholders.** No real email address or phone number
  was invented for the organization. Replace the placeholder
  `donate@priyosha-charity.example` in `frontend/app/page.tsx` (and this
  README) with the organization's real contact details before launch.
  A matching address such as `contact@priyoshacharity.com` could be used
  once email is set up for the domain, but that address is not claimed
  to exist yet — do not publish it until it is actually created.
- **Storage is a JSON file**, not a database. Fine for starting out; move to
  Postgres or similar before volume grows, and back up the `data/` files.
- **Set `ALLOWED_ORIGIN`** to your exact frontend origin in production
  (`https://priyoshacharity.com` — already the default in `.env.example`).
  The gateway now refuses to start in production if it is empty or `*` —
  do not work around that check.
- **Domain is set.** The live domain is `priyoshacharity.com` /
  `api.priyoshacharity.com`, configured in `.env.example`, the frontend
  metadata fallback, `docker-compose.yml` and `deploy/Caddyfile`.
  `www.priyoshacharity.com` redirects to `https://priyoshacharity.com`.

## Project layout

```
priyosha-charity-code/
├── README.md
├── GITHUB_SETUP.md        # how to create the GitHub repo and push
├── .env.example           # copy to .env — DOMAIN, URLs, ADMIN_TOKEN, ports
├── docker-compose.yml
├── .gitignore
├── deploy/
│   ├── Caddyfile          # example reverse proxy (site :3000, api :8080, HTTPS)
│   └── README.md          # DNS A-record + server steps
├── frontend/                 # Next.js (App Router, TypeScript), plain CSS
│   ├── app/
│   │   ├── layout.tsx        # metadataBase from NEXT_PUBLIC_SITE_URL / FRONTEND_URL
│   │   ├── page.tsx          # all sections; forms call the Go gateway
│   │   ├── sitemap.ts
│   │   ├── robots.ts
│   │   └── globals.css       # maroon #7a2436 brand styles
│   ├── public/logo.png       # Priyosha Charity logo
│   ├── Dockerfile
│   ├── package.json
│   └── README.md
└── backend/
    ├── api-gateway/          # Go, :8080 — routes /api/*, CORS, admin-token check on GET lists
    ├── donation-service/     # Go, :8081 — POST /pledges (public), GET /pledges (admin only)
    ├── contact-service/      # Go, :8082 — POST /messages (public), GET /messages (admin only)
    └── content-service/      # Go, :8083 — GET /stats, /programs
        # each: go.mod, main.go (stdlib only), Dockerfile, README.md
```

## Verification done during the build

- `go build ./...` and `go vet ./...` pass for all four Go services.
- `npm install` and `npm run build` pass for the frontend.
- All four Go services were started directly (Docker was unavailable in the
  build environment) and the gateway was exercised with `curl`: health
  endpoints, `GET /api/stats`, `GET /api/programs`, a `POST /api/pledges`
  and a `POST /api/messages`, and the records were confirmed persisted by
  `GET /api/pledges` and `GET /api/messages` (and in the JSON data files).
  An invalid pledge (`amount: 0`) was correctly rejected with HTTP 400.
- The production frontend was served both with `npm start` and with the
  standalone server (`node .next/standalone/server.js`); the homepage
  returned HTTP 200 and contained the org name, the 18.1 million / 22,185 /
  16.7 lakh figures, the founder section and the logo reference.
- Production hardening was then verified separately: `GET /api/pledges`
  and `GET /api/messages` without a token return 403, and with the correct
  `X-Admin-Token` return the lists; the gateway refuses to start in
  production (`ENV=production`) when `ALLOWED_ORIGIN` is empty or `*`.
  `docker compose up` itself was not run (Docker unavailable in the build
  environment) — its YAML was checked and the services were run directly.
