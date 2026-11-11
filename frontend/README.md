# Priyosha Charity Organization — Frontend

Next.js (App Router, TypeScript) website. Plain CSS — no CSS framework — so the build stays small and reliable.

## Pages / sections

One page (`app/page.tsx`): header with logo, hero, mission, The Need (stats fetched from the Go content-service), Programs, interactive pledge/donation calculator, contact/volunteer form, founder section, footer.

## How it talks to the backend

All forms call the Go **api-gateway** — they do not fake-submit locally:

- `GET {NEXT_PUBLIC_API_URL}/api/stats` and `/api/programs` — content
- `POST {NEXT_PUBLIC_API_URL}/api/pledges` — pledge calculator
- `POST {NEXT_PUBLIC_API_URL}/api/messages` — contact form

If the API cannot be reached, the page falls back to a copy of the same figures compiled into `page.tsx`, so visitors still see the correct numbers.

## Env vars

| Var | Default | Meaning |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | URL of the Go api-gateway |
| `NEXT_PUBLIC_SITE_URL` / `FRONTEND_URL` | `https://priyoshacharity.com` | Canonical site URL for metadata, `/sitemap.xml` and `/robots.txt` |

Copy `.env.example` to `.env.local` and adjust for your environment. In production `NEXT_PUBLIC_API_URL` must be your deployed gateway URL and `NEXT_PUBLIC_SITE_URL` your real domain.

## Run locally

```bash
npm install
npm run dev   # http://localhost:3000
```

## Build / start (production)

```bash
npm install
npm run build
node .next/standalone/server.js   # http://localhost:3000
```

(The build uses `output: "standalone"`, so `npm start` is not the production
entry point — the standalone server is. For local development use
`npm run dev`.)

## Logo

`public/logo.png` — the organization's logo (person embracing a dog forming a heart, deep maroon on white). Brand colour: `#7a2436`.
