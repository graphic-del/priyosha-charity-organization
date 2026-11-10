# Push this code to GitHub

No GitHub repo has been created yet — no credentials or repo URL were
provided, and nothing has been pushed. The local repo is ready: it is
initialised, committed, and clean. Pushing is one step once you create the
repo.

GitHub user: **graphic-del**
Repo name: **priyosha-charity-organization**
Remote: `https://github.com/graphic-del/priyosha-charity-organization.git`

If you create the repo under a different name, only change that one word
(`priyosha-charity-organization`) in the remote URL below.

## 1. Create the repo on GitHub

1. Go to https://github.com/new (sign in as **graphic-del** first).
2. **Repository name:** `priyosha-charity-organization`
3. **Description (optional):** `Website for Priyosha Charity Organization — Next.js frontend, Go microservices backend.`
4. Leave it **Public** or choose **Private** — either works.
5. Do **NOT** tick “Add a README”, “Add .gitignore”, or “Choose a license”
   — this folder already has a README and .gitignore, and adding GitHub's
   would conflict with the first push.
6. Click **Create repository**.

## 2. Push the local code (run in this folder)

```bash
cd priyosha-charity-code

# Point the local repo at your new GitHub repo (once only).
git remote add origin https://github.com/graphic-del/priyosha-charity-organization.git

# Make sure the branch is called main, then push.
git branch -M main
git push -u origin main
```

You will be asked to sign in — use a GitHub personal access token as the
password when prompted (GitHub no longer accepts account passwords for
pushing). Create one at https://github.com/settings/tokens with the `repo`
scope, and store it in a password manager, never in a file in this repo.

Verify on github.com that the files are there, and that `node_modules/`,
`.next/`, `.env` and `data/` are **not** listed (they are gitignored).

## 3. Clone / update later

```bash
# On another machine:
git clone https://github.com/graphic-del/priyosha-charity-organization.git
cd priyosha-charity-organization
cp .env.example .env   # then fill in your real values (ADMIN_TOKEN, ENV)

# After making changes locally:
git add -A
git commit -m "Describe what changed"
git push
```

## 4. Deploy from GitHub (optional)

On your server, `git clone` the repo instead of copying the folder, then
follow `README.md` → Deploy. To update the live site later:

```bash
git pull
docker compose up --build -d
```
