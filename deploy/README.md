# Deploy — reverse proxy (example)

`Caddyfile` is an example Caddy configuration that fronts the site with HTTPS.
It is not a live deploy — copy it to your server. The live domain is
**priyoshacharity.com**.

> **Setting up the server for the first time? Start with
> [SERVER_SETUP.md](SERVER_SETUP.md)** — step-by-step instructions for a
> strong Ubuntu 24.04 server (4 vCPU / 8 GB RAM): SSH key only, firewall,
> Docker, fail2ban, automatic updates, `.env` secrets, Caddy HTTPS, DNS,
> optional Cloudflare, backups and updates. The short version below is
> only a summary.

## DNS steps (do these first)

In your DNS provider add A records pointing at your server's public IP:

- `priyoshacharity.com` → `<server IP>`
- `www.priyoshacharity.com` → `<server IP>` (redirects to the bare domain)
- `api.priyoshacharity.com` → `<server IP>`

Wait for DNS to propagate (usually minutes, sometimes up to a few hours).
Check with `dig +short priyoshacharity.com` — it should return your server IP.

## Server steps (summary — see SERVER_SETUP.md for the full guide)

1. Copy this repo to the server and copy `.env.example` to `.env`, filling in
   only the secrets: `ADMIN_TOKEN` (generate with `openssl rand -hex 32`)
   and `ENV=production`. The domain values are already set to
   priyoshacharity.com (`DOMAIN`, `FRONTEND_URL`,
   `NEXT_PUBLIC_API_URL`, `ALLOWED_ORIGIN`). Never commit `.env`.
2. `docker compose up --build -d`
3. Run Caddy with `deploy/Caddyfile` (see the comments in that file and
   Step 8 in SERVER_SETUP.md — Caddy needs `DOMAIN=priyoshacharity.com`
   in its environment). Caddy fetches Let's Encrypt certificates
   automatically.
4. Verify: `https://priyoshacharity.com` loads the site,
   `https://www.priyoshacharity.com` redirects to it, and
   `https://api.priyoshacharity.com/health` returns `{"status":"ok","service":"api-gateway"}`.

## Security

See [../SECURITY.md](../SECURITY.md): security headers, admin-token
lists, rate limiting, and the pre-launch security checklist. Only ports
80/443 are public — the website (3000) and API (8080–8083) are bound to
`127.0.0.1` in `docker-compose.yml`.
