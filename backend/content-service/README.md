# content-service

Serves the public facts and programme data for Priyosha Charity Organization as JSON, on port **8083**. Stateless — the data is compiled in.

## Endpoints

- `GET /health` — `{"status":"ok","service":"content-service"}`
- `GET /stats` — the need in Bangladesh, plus the warning that the figures cannot be added together
- `GET /programs` — the charity's programmes

## The figures (as provided by the founder)

- **18.1 million / 1.81 crore** people projected to face crisis-level food insecurity or worse during September–December 2026, within the areas assessed — Integrated Food Security Phase Classification (IPC)
- **22,185** people counted sleeping in public places in the 2022 census — narrow category (streets, stations, similar), does not capture everyone without secure housing, older count
- **About 16.7 lakh dogs** estimated in a 2024 study — estimates the dog population; does not establish how many are homeless or hungry
- **Hungry/homeless animals overall, including cats:** no reliable nationwide total found — research identifies gaps in Bangladesh's stray-animal population data; giving an exact count would be guesswork

Warning shown with them: *these figures cannot be added together — hunger and homelessness can affect the same people, and animal population estimates do not measure suffering.*

To update a figure, edit the `stats` slice in `main.go` and redeploy this service only — the frontend reads from the API, so no frontend change is needed.

## Env vars

| Var | Default | Meaning |
|---|---|---|
| `PORT` | `8083` | Listen port |

## Run locally

```bash
go run .
```

Stdlib only — no external dependencies.
