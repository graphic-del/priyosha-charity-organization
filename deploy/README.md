# Deploy — reverse proxy (example)

`Caddyfile` is an example Caddy configuration that fronts the site with HTTPS.
It is not a live deploy — copy it to your server. The live domain is
**priyoshacharity.com**.

## DNS steps (do these first)

In your DNS provider add A records pointing at your server's public IP:

- `priyoshacharity.com` → `<server IP>`
- `www.priyoshacharity.com` → `<server IP>` (redirects to the bare domain)
- `api.priyoshacharity.com` → `<server IP>`

Wait for DNS to propagate (usually minutes, sometimes up to a few hours).
Check with `dig +short priyoshacharity.com` — it should return your server IP.

## Server steps

1. Copy this repo to the server and copy `.env.example` to `.env`, filling in
   only the secrets: `ADMIN_TOKEN` and `ENV=production`. The domain values
   are already set to priyoshacharity.com (`DOMAIN`, `FRONTEND_URL`,
   `NEXT_PUBLIC_API_URL`, `ALLOWED_ORIGIN`).
2. `docker compose up --build -d`
3. Run Caddy with `deploy/Caddyfile` (see the comments in that file).
   Caddy fetches Let's Encrypt certificates automatically.
4. Verify: `https://priyoshacharity.com` loads the site,
   `https://www.priyoshacharity.com` redirects to it, and
   `https://api.priyoshacharity.com/health` returns `{"status":"ok","service":"api-gateway"}`.
