# contact-service

Records messages from the website's contact / volunteer form, on port **8082**.

## Endpoints

- `GET /health` — `{"status":"ok","service":"contact-service"}`
- `GET /messages` — **admin only**: requires `X-Admin-Token: <ADMIN_TOKEN>`,
  returns 403 without a valid token, and 403 for every request while
  `ADMIN_TOKEN` is unset. Lists messages, newest first.
- `POST /messages` — public. Body `{"name": "...", "email": "...", "interest": "donate"|"volunteer"|"other", "message": "..."}` → `201` with the recorded message

## Storage

In-memory, persisted to a JSON file (`DATA_FILE` env var, default `./data/messages.json`) so messages survive a restart.

## ⚠️ No email is sent yet

A message is only stored. To notify the team, send an email via your provider (SendGrid, AWS SES, …) in `main.go` — see the `TODO (real email)` comment there — or read messages with `GET /messages`.

## Env vars

| Var | Default | Meaning |
|---|---|---|
| `PORT` | `8082` | Listen port |
| `DATA_FILE` | `./data/messages.json` | Where messages are persisted |
| `ADMIN_TOKEN` | *(empty)* | Required to read messages (`GET /messages`). If empty, `GET /messages` always returns 403 |

## Run locally

```bash
go run .
```

Stdlib only — no external dependencies.
