# donation-service

Records pledges (promises to give) for Priyosha Charity Organization, on port **8081**.

## Endpoints

- `GET /health` — `{"status":"ok","service":"donation-service"}`
- `GET /pledges` — **admin only**: requires `X-Admin-Token: <ADMIN_TOKEN>`,
  returns 403 without a valid token, and 403 for every request while
  `ADMIN_TOKEN` is unset. Lists pledges, newest first.
- `POST /pledges` — public. Body `{"amount": 50, "frequency": "one-time"|"monthly", "name": "...", "email": "..."}` → `201` with the recorded pledge

## Storage

In-memory, persisted to a JSON file (`DATA_FILE` env var, default `./data/pledges.json`) so pledges survive a restart. Replace with a real database before handling real money.

## ⚠️ Payments are NOT wired up

A pledge here is a record of intent, **not a charge**. To take real payments, create a Stripe Checkout Session / PaymentIntent in `main.go` — see the `TODO (real payments)` comment there — and never store card numbers in this service. Card entry must happen on the payment provider's hosted page.

## Env vars

| Var | Default | Meaning |
|---|---|---|
| `PORT` | `8081` | Listen port |
| `DATA_FILE` | `./data/pledges.json` | Where pledges are persisted |
| `ADMIN_TOKEN` | *(empty)* | Required to read pledges (`GET /pledges`). If empty, `GET /pledges` always returns 403 |

## Run locally

```bash
go run .
```

Stdlib only — no external dependencies.
