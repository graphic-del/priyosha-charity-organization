# Strong server setup — Priyosha Charity Organization

> **Who this is for:** the founder, setting up the live site for the first
> time. You need a terminal (the black window where you type commands) for
> the server part — there is no button for this on the server. Every step
> below gives you the exact commands to copy, one step at a time, and tells
> you what success looks like. Do one step, check it worked, then move on.
> If a step fails, stop there and read the error — do not skip ahead.

> **Honest note:** no server is 100% unhackable. This guide makes the
> server as hard as reasonably possible to break into: key-only login,
> a firewall, automatic security updates, and a blocker for repeated
> login attempts. See `../SECURITY.md` for the full security model.

## What you are building

- **Server:** Ubuntu 24.04, **1 vCPU / 2 GB RAM / 70 GB SSD** — the
  DigitalOcean Droplet you already created, named `priyosha`, in
  New York (NYC3). It costs $16/month for the Droplet, **$19.20/month
  total with automated backups**. That is enough to launch Priyosha:
  the site is light and will run well on it. If lots of people start
  visiting later, you can make it bigger in one click in DigitalOcean
  (Resize) — you do not need to buy a bigger server now. Size is about
  speed, not security: the security in this guide stays tight either way.
  Because 2 GB RAM is tight for building the website, Step 4 adds a
  2 GB swap file first — do not skip it, or the first build can run
  out of memory and fail.
- **On the server:** Docker + Docker Compose run the site (Next.js) and
  the four Go services; Caddy gives you HTTPS automatically.
- **Domain:** `priyoshacharity.com` (site), `www.priyoshacharity.com`
  (redirects to the site), `api.priyoshacharity.com` (the API).

## Before you start

You need:
- [x] Your DigitalOcean Droplet `priyosha` — already created and Active
      (NYC3, Ubuntu 24.04, 1 vCPU / 2 GB RAM). Do not create another one.
- [x] Your Droplet's public IP address — **`159.65.241.137`**. This
      guide already uses that IP everywhere below, so you can copy the
      commands as they are. (Private IP `10.108.0.2` is only for inside
      DigitalOcean — do not use it for DNS or SSH.)
- [ ] The domain `priyoshacharity.com` and access to its DNS settings
      (where you bought the domain).
- [ ] A computer with a terminal. On Windows use **Windows Terminal** or
      **Git Bash**; on Mac open **Terminal** (Spotlight → Terminal).
- [ ] About an hour, unhurried. Do not do this at 2 a.m. the night
      before launch.

## Step 1 — Create the server ✅ Already done — skip this step

Your server is already created — do not create another one:

- **Name:** `priyosha`
- **Provider / region:** DigitalOcean, New York (NYC3)
- **Image / OS:** Ubuntu 24.04 LTS x64
- **Size:** 1 vCPU / 2 GB RAM / 70 GB SSD — $16/month, $19.20/month
  total with automated backups
- **Authentication:** your SSH key (ed25519, made on your Windows PC
  and added in DigitalOcean as `Priyosha`)
- **Public IP:** **`159.65.241.137`** — this guide uses it everywhere
  below, so copy the commands as they are.
- **Status:** Active

Go straight to Step 2.

## Step 2 — Log in and make your own user (do this first)

Log in from your computer's terminal (replace the IP):

```bash
ssh root@159.65.241.137
```

The first time it asks "Are you sure you want to continue connecting?"
— type `yes` and press Enter.

Now create a non-root user (here called `priyosha` — use your own name
if you prefer) and give it sudo (admin) rights. Run these on the server,
one at a time:

```bash
adduser priyosha
usermod -aG sudo priyosha
mkdir -p /home/priyosha/.ssh
cp ~/.ssh/authorized_keys /home/priyosha/.ssh/authorized_keys
chown -R priyosha:priyosha /home/priyosha/.ssh
chmod 700 /home/priyosha/.ssh
chmod 600 /home/priyosha/.ssh/authorized_keys
```

**Success looks like:** `adduser` asks you for a password for the new
user and some details (you can leave the details blank). Then log out
and back in as the new user, and use that from now on:

```bash
exit
ssh priyosha@159.65.241.137
```

## Step 3 — Lock the front door (SSH, firewall)

Run these on the server, as `priyosha`. Each command asks for your
password the first time (that is `sudo` — normal).

**3a. Turn off password login and root login.** Open the SSH settings:

```bash
sudo nano /etc/ssh/sshd_config
```

Find these lines (use Ctrl+W to search) and set them exactly like this
(remove the `#` at the start if there is one):

```
PasswordAuthentication no
PermitRootLogin no
PubkeyAuthentication yes
```

Save with Ctrl+O, Enter, then Ctrl+X to exit. Then restart SSH and test
**in a second terminal window first** — keep your current window open:

```bash
sudo systemctl restart ssh
```

In the new window, try `ssh priyosha@159.65.241.137` — it should log you
in with your key and **not** ask for a password. Only if that works,
close the old window. If it does not work, fix it from the still-open
window before logging out — otherwise you can lock yourself out.

**3b. Firewall — only three things allowed in.** UFW is Ubuntu's
firewall:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable
sudo ufw status verbose
```

**Success looks like:** the status shows `80/tcp ALLOW`, `443/tcp
ALLOW`, `OpenSSH ALLOW`, and `Default: deny (incoming)`. Ports 3000 and
8080–8083 are **not** in the list — they are bound to localhost only
(`127.0.0.1` in `docker-compose.yml`), so nobody on the internet can
reach the website or API directly. Only Caddy (80/443) is public.

## Step 4 — Add swap, then install Docker, fail2ban, automatic updates

Still on the server, as `priyosha`:

**4a. Add a 2 GB swap file first — do not skip this.** Your Droplet
has only 2 GB RAM, and building the Next.js website needs more than
that for a few minutes. Swap gives it that extra room (slower, but it
lets the build finish instead of being killed for out-of-memory).

```bash
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
free -h
```

**Success looks like:** `free -h` shows `Swap: 2.0Gi` (and `Mem:
1.9Gi` total). If `fallocate` says it is not supported, use
`sudo dd if=/dev/zero of=/swapfile bs=1M count=2048` instead, then
run the `chmod` / `mkswap` / `swapon` lines again.

**4b. Install Docker and the security tools:**

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io docker-compose-plugin fail2ban unattended-upgrades
sudo usermod -aG docker priyosha
```

Log out and back in once so the Docker group takes effect:

```bash
exit
ssh priyosha@159.65.241.137
docker --version
docker compose version
```

**Success looks like:** both commands print a version number.
`fail2ban` starts automatically and blocks IPs that fail logins
repeatedly. `unattended-upgrades` installs security patches
automatically — check it is enabled:

```bash
sudo systemctl status fail2ban --no-pager | head -5
sudo systemctl status unattended-upgrades --no-pager | head -5
```

Both should say `active (running)`.

## Step 5 — Get the code onto the server

```bash
cd ~
git clone https://github.com/graphic-del/priyosha-charity-organization.git
cd priyosha-charity-organization
```

**Success looks like:** `git clone` prints "Cloning into..." and the
folder contains `README.md`, `frontend/`, `backend/`, `deploy/` and
`docker-compose.yml`. Check with `ls`.

> ⚠️ **Check this before Step 6:** the clone must include the security
> hardening — run `ls deploy/` and `cat SECURITY.md | head -5`. If
> `SECURITY.md` or `deploy/SERVER_SETUP.md` is missing, GitHub does
> not have the latest code yet. Stop here and upload the hardened code
> to GitHub first (the same web-upload way as before, unzipped — do
> not upload the zip itself), then delete the cloned folder and run
> `git clone` again. Do not launch with the old code: it has no rate
> limits, no security headers, and its pledge/message lists are open.

## Step 6 — Secrets: make your `.env` (do this carefully)

```bash
cp .env.example .env
openssl rand -hex 32
```

The last command prints a long random string — that is your
`ADMIN_TOKEN`. Copy it, then open `.env`:

```bash
nano .env
```

Set exactly these lines (the domain lines are already correct in
`.env.example` — check them, do not change them):

```
ENV=production
DOMAIN=priyoshacharity.com
FRONTEND_URL=https://priyoshacharity.com
NEXT_PUBLIC_API_URL=https://api.priyoshacharity.com
NEXT_PUBLIC_SITE_URL=https://priyoshacharity.com
ALLOWED_ORIGIN=https://priyoshacharity.com
ADMIN_TOKEN=<paste the long random string from openssl here>
```

Save with Ctrl+O, Enter, Ctrl+X. **Never commit `.env`, never paste
the token into GitHub, chat, or email.** Share it only with the site
owner, privately.

**Success looks like:** `cat .env` shows `ENV=production` and a long
`ADMIN_TOKEN` value. (Do not screenshot `.env` and send it to anyone.)

## Step 7 — Start the site

```bash
docker compose up --build -d
docker compose ps
```

The first build takes a few minutes — on this 2 GB Droplet allow
5–10 minutes, and do not interrupt it. **Success looks like:** `docker
compose ps` shows five containers — frontend, gateway, content,
donation, contact — all with status `Up`. If one says `Exited`, read
its logs before going on. If the frontend build is killed or you see
"out of memory", check swap is on with `free -h` (Step 4a) and try
the build again — do not raise the server size just for the build:

```bash
docker compose logs --tail 50 <container-name>
# e.g. docker compose logs --tail 50 priyosha-gateway
```

Check the site works on the server itself (before DNS):

```bash
curl -s http://localhost:3000/ | head -c 200; echo
curl -s http://localhost:8080/health; echo
curl -s http://localhost:8080/api/stats | head -c 200; echo
```

**Success looks like:** the first command shows HTML, the second shows
`{"status":"ok","service":"api-gateway"}`, the third shows the figures.

## Step 8 — Caddy for HTTPS

Install Caddy and run it with the included config:

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install -y caddy
sudo cp ~/priyosha-charity-organization/deploy/Caddyfile /etc/caddy/Caddyfile
```

Caddy reads the domain from the environment, so start it with the
domain set. The simplest reliable way on a first install is to edit
the Caddy systemd override once:

```bash
sudo systemctl edit caddy
```

Add these lines in the editor, then save:

```
[Service]
Environment=DOMAIN=priyoshacharity.com
```

Then:

```bash
sudo systemctl restart caddy
sudo systemctl status caddy --no-pager | head -8
```

**Success looks like:** Caddy is `active (running)`. It will fetch
Let's Encrypt certificates automatically once DNS points at the server
(next step) — until then it may show certificate errors, which is
normal.

## Step 9 — DNS: point the domain at the server

In your domain/DNS provider (where you bought `priyoshacharity.com`),
add three **A records**, all pointing at your server's public IP
`159.65.241.137`:

| Name / Host | Type | Value |
|---|---|---|
| `@` (or blank — the bare domain) | A | `159.65.241.137` |
| `www` | A | `159.65.241.137` |
| `api` | A | `159.65.241.137` |

DNS takes minutes to a few hours to spread. Check from your computer:

```bash
dig +short priyoshacharity.com
dig +short api.priyoshacharity.com
```

**Success looks like:** both commands return your server IP. Then wait
a few minutes and open in a browser:

- `https://priyoshacharity.com` — the site, with a padlock.
- `https://www.priyoshacharity.com` — redirects to the bare domain.
- `https://api.priyoshacharity.com/health` — shows
  `{"status":"ok","service":"api-gateway"}`.

Check the security headers are there:

```bash
curl -sI https://priyoshacharity.com | grep -i -E 'strict-transport|x-frame|x-content|content-security'
curl -sI https://api.priyoshacharity.com/health | grep -i -E 'strict-transport|x-frame|x-content|content-security'
```

**Success looks like:** both show `Strict-Transport-Security`,
`X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff` and a
`Content-Security-Policy` line.

## Step 10 — Optional: Cloudflare in front (free, extra protection)

This is optional, but it is the easiest extra shield against floods
and bots. Only do it after Steps 1–9 work without it.

1. Create a free Cloudflare account, add `priyoshacharity.com`.
2. Cloudflare gives you two nameservers — change them at your domain
   registrar (where you bought the domain) to Cloudflare's.
3. In Cloudflare → DNS, make sure the three A records above exist and
   the cloud icon is **orange (Proxied)** for all three.
4. In Cloudflare → SSL/TLS, set the mode to **Full (strict)**.
5. Wait for DNS to spread, then check the site again as in Step 9.

If anything breaks, turn the orange clouds grey (DNS only) in
Cloudflare and the site goes back to working directly — that is your
undo button.

## Keeping it safe after launch

- **Updates (monthly, at least):**
  ```bash
  ssh priyosha@159.65.241.137
  sudo apt update && sudo apt upgrade -y
  ```
  Security patches install automatically, but check every month that
  `fail2ban` and `unattended-upgrades` are still running (Step 4).
- **Back up the pledges and messages (weekly at least; daily once real
  donations start):**
  ```bash
  docker compose exec donation-service cat /data/pledges.json > ~/backup-pledges-$(date +%F).json
  docker compose exec contact-service cat /data/messages.json > ~/backup-messages-$(date +%F).json
  ```
  Then copy those files off the server (e.g. `scp`) to somewhere only
  the founder can reach. Do not store backups in Git.
- **Update the code:**
  ```bash
  cd ~/priyosha-charity-organization
  git pull
  docker compose up --build -d
  ```
  Then check the site and `https://api.priyoshacharity.com/health`.
- **If the admin token ever leaks:** generate a new one
  (`openssl rand -hex 32`), put it in `.env`, run
  `docker compose up -d` to restart, and tell only the people who
  need it.
- **Check the logs sometimes:**
  ```bash
  docker compose logs --tail 100
  sudo fail2ban-client status sshd
  sudo ufw status verbose
  ```
  The firewall should still show only OpenSSH, 80 and 443 allowed.

## If something goes wrong

- **Site does not load:** check Caddy (`sudo systemctl status caddy`),
  check the containers (`docker compose ps`), check DNS
  (`dig +short priyoshacharity.com`).
- **API returns 403 for pledges/messages:** that is correct without a
  token. With the token: `curl -H 'X-Admin-Token: <your token>'
  https://api.priyoshacharity.com/api/pledges`.
- **Gateway refuses to start:** read the log — in production it refuses
  to start if `ALLOWED_ORIGIN` is empty or `*`
  (`docker compose logs priyosha-gateway`).
- **429 errors:** the rate limit is working (10 pledges/messages per
  minute per IP). Wait a minute and try again — do not raise the limit
  just because one test hit it.
