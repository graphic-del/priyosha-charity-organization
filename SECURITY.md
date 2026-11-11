# Security — Priyosha Charity Organization

> **Honest note first:** no website is 100% unhackable — anyone who promises
> that is not telling the truth. What this project does is make Priyosha
> *as hard as reasonably possible* to break into, with several layers, so
> that one mistake does not hand an attacker the site. Keep the software
> updated, keep secrets secret, and keep backups — that is most of security.

## What is already protected

**The site itself**
- **HTTPS only** in production (Caddy + Let's Encrypt, see
  `deploy/Caddyfile`). Browsers are told, via `Strict-Transport-Security`,
  to only ever use HTTPS for this domain for a year.
- **Security headers on every response** — from the Next.js frontend
  (`frontend/next.config.mjs`) and from every Go service:
  - `X-Content-Type-Options: nosniff` — browsers must not guess file types.
  - `X-Frame-Options: DENY` + `frame-ancestors 'none'` — the site cannot
    be loaded inside another page (clickjacking).
  - `Referrer-Policy: strict-origin-when-cross-origin` — browsers send
    as little referrer information as possible.
  - `Permissions-Policy: camera=(), microphone=(), geolocation=()` — the
    site never asks for camera, microphone or location.
  - `Content-Security-Policy` — the frontend only loads scripts, styles,
    images and API calls from itself and `https://api.priyoshacharity.com`;
    the Go API answers allow nothing to be loaded at all.
- **No card numbers or passwords are ever stored.** A pledge is a record
  of intent, not a charge — real payments are not wired up yet, and when
  they are, card entry must happen on the payment provider's own page.

**The API**
- **Admin lists are locked.** `GET /api/pledges` and `GET /api/messages`
  need the `X-Admin-Token` header matching the `ADMIN_TOKEN` env var.
  Without it — or if `ADMIN_TOKEN` is unset — they return **403**. The
  check runs in the gateway *and* in the donation/contact services.
  Donor and message names and emails are never public.
- **CORS is locked to one origin.** Only `https://priyoshacharity.com`
  may call the API from a browser. In production the gateway **refuses to
  start** if `ALLOWED_ORIGIN` is empty or `*`.
- **Rate limiting** (gateway, per IP): `POST /api/pledges` and
  `POST /api/messages` — 10 per minute; `GET /api/*` — 120 per minute.
  Over the limit returns **429** with a `Retry-After` header. This stops
  one person flooding the forms with fake pledges or messages.
- **Input checks.** Pledges: amount $1–$100,000, frequency
  `one-time`/`monthly`, name 1–100 characters, email a valid address
  (max 254 characters). Messages: name 1–100, email valid, interest
  `donate`/`volunteer`/`other`, message 1–2000 characters. Anything with
  `<` or `>` in a name, email or message is rejected (400) — no HTML or
  scripts can be stored. Request bodies are capped at 64 KB, and every Go
  server has read/write timeouts so a slow connection cannot hang it.
- **Secrets live in `.env`, never in Git.** `.env` is git-ignored. The
  repo only ever contains `.env.example` with placeholders.

**The server** (see `deploy/SERVER_SETUP.md`)
- Only ports 80 (HTTP, redirects to HTTPS) and 443 (HTTPS) are open to
  the internet. The website (3000) and API (8080–8083) are bound to
  `127.0.0.1` — localhost only — and are reached only through Caddy.
- SSH by key only; password login and root login are disabled.
- Firewall (UFW): allow OpenSSH, 80, 443 only; deny everything else.
- `fail2ban` blocks repeated failed logins; `unattended-upgrades` keeps
  security patches installing automatically.
- Services run in Docker containers, each doing one job, and restart
  automatically unless stopped.

## What is NOT protected yet — do these before launch

- [ ] **Set a real `ADMIN_TOKEN`.** Generate with
  `openssl rand -hex 32`, put it in `.env` on the server only, share it
  only with the site owner. Never put it in Git, chat, or email.
- [ ] **Set a real contact email.** The code still says
  `donate@priyosha-charity.example` (a placeholder — that address does
  not exist). Replace it with the organization's real address in
  `frontend/app/page.tsx` and the README *after* the mailbox exists.
  Until then, security reports should go to the founder directly.
- [ ] **Point DNS at the server** — `priyoshacharity.com`,
  `www.priyoshacharity.com`, `api.priyoshacharity.com` (see
  `deploy/SERVER_SETUP.md`).
- [ ] **Real payments / real email are not wired up.** Do not collect card
  details on this site until a payment provider (e.g. Stripe) is
  integrated on the provider's hosted page. Do not paste email passwords
  into the code — use a provider API key in `.env` when that work is done.

## How to report a security issue

If you find a security problem in this site, **do not post it publicly**
(for example, do not open a public GitHub issue about it).

Until a real contact email is published, report it to the founder,
**Md Shadman Mahmood**, directly — and say in the message that it is a
security report for Priyosha Charity Organization, with what you found
and how to reproduce it, and nothing else. The founder will fix it and
then tell people what changed. Do not test the live site with attacks
that could delete or expose other people's pledges or messages.

## Update policy

- **Update the server every month, at least:** `sudo apt update &&
  sudo apt upgrade -y` (unattended-upgrades covers security patches
  automatically, but check it is running — see
  `deploy/SERVER_SETUP.md`).
- **Update the code's dependencies when a security fix appears:**
  `cd frontend && npm audit` and update Next.js/React when a fix is
  released; `go` has no external dependencies here, so keep the Go
  toolchain itself updated on the build machine.
- **Rebuild and redeploy after any dependency update:**
  `git pull && docker compose up --build -d`, then check
  `https://priyoshacharity.com` and
  `https://api.priyoshacharity.com/health` load.
- **Rotate `ADMIN_TOKEN`** if it was ever shared in the wrong place, or
  if someone who had it leaves the organization: generate a new one,
  put it in `.env`, restart the services, and tell only the people who
  need it.

## Backup policy

- **What to back up:** the Docker volumes holding
  `data/pledges.json` and `data/messages.json` (the recorded pledges and
  messages), plus `.env` (kept somewhere safe, *not* in Git) and the
  GitHub repo itself.
- **How often:** at least weekly while the site is small; daily once
  real donations start. Copy the two JSON files off the server
  (see `deploy/SERVER_SETUP.md` for the commands) and store the copy
  somewhere only the founder can reach.
- **Test a restore** before you need one: copy the files back and check
  the admin lists still show the records.
- **Move to a real database** (e.g. Postgres) before volume grows — JSON
  files are fine to start, not a long-term store for donor records.

## Pre-launch security checklist

- [ ] Server created with the steps in `deploy/SERVER_SETUP.md`
  (key-only SSH, UFW, fail2ban, auto-updates).
- [ ] `ENV=production` in `.env`; gateway starts (it refuses to start
  with `ALLOWED_ORIGIN` empty or `*`).
- [ ] `ALLOWED_ORIGIN=https://priyoshacharity.com`,
  `NEXT_PUBLIC_API_URL=https://api.priyoshacharity.com`.
- [ ] `ADMIN_TOKEN` is a long random value, in `.env` only, and
  `GET /api/pledges` without it returns 403.
- [ ] `https://priyoshacharity.com` loads with a padlock; `http://`
  redirects to `https://`; `curl -I` shows the security headers
  (`Strict-Transport-Security`, `X-Frame-Options: DENY`,
  `Content-Security-Policy`).
- [ ] Sending 12 rapid pledges returns 429 for the later ones.
- [ ] Invalid pledge (amount 0, bad email, `<script>` in a name)
  returns 400.
- [ ] Real contact email replaces the placeholder; no real token,
  password or email credential is in Git.
- [ ] A backup of `data/pledges.json` and `data/messages.json` exists
  off the server.
- [ ] Only ports 80/443 are open — check with the firewall status
  command in `deploy/SERVER_SETUP.md`.
