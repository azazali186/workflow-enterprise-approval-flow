# Approval Flow — Admin Console (Frontend)

A production-grade SaaS admin dashboard for the Approval Flow enterprise workflow
backend. React 18 + Vite + TypeScript + Redux Toolkit + React Query + Tailwind CSS v4.

## Stack

| Layer      | Choice                                                        |
| ---------- | ------------------------------------------------------------- |
| UI         | React 18, Vite 8, TypeScript (strict)                         |
| State      | Redux Toolkit (auth / toasts / UI) + React Query (server state) |
| Styling    | Tailwind CSS v4, design tokens via `@theme` in `src/index.css` |
| Forms      | React Hook Form + Zod                                          |
| Animation  | Framer Motion (transitions, modals, micro-interactions)        |
| Charts     | Pure SVG (donut chart, stat sparklines) — no heavy chart dep   |
| Routing    | React Router v6, lazy-loaded pages, route guards               |

## Getting started

```bash
npm install
cp .env.example .env   # then set VITE_API_BASE_URL if the backend isn't at :8080
npm run dev            # http://localhost:5173 (proxies /api -> http://localhost:8080)
```

Production build + preview:

```bash
npm run build
npm run preview
```

### Environment

- `VITE_API_BASE_URL` — default `http://localhost:8080`. Leave unset when using the
  Vite dev proxy (recommended — avoids CORS entirely).

## Backend contract

- **Base path** `/api/v1` — every request goes through the centralized client in
  `src/services/api/client.ts`.
- **Methods** — the backend is POST-based (POST for list/get/create/update/delete).
  The client exposes `post`, `postList`, `patch`, `del` helpers that map to the
  backend routes.
- **Response shape** — all responses are normalized by `src/services/api/normalize.ts`:

  ```ts
  // list endpoints
  { data: T[], pagination: { total, limit, offset }, summary: {...} }
  ```

  The normalizer unwraps the nested `{ code, message, data }` envelope, tolerates
  `data` being an object or an array, and surfaces `pagination` + `summary` when
  present. Missing fields are handled safely everywhere.
- **Errors** — `src/services/api/errors.ts` maps HTTP codes to friendly messages
  (401 → re-auth flow, 429 → rate-limit hint, 400 → first validation message, …).
- **Auth** — access token in `localStorage`; a 401-triggered refresh loop re-issues
  tokens via `/auth/refresh`; on failure the user is signed out. The auth slice in
  Redux is the single source of truth for the session.

## Architecture

Feature-based, one folder per domain:

```
src/
  services/          # API layer (client + one service module per feature)
  features/<feature>/
    components/      # UI pieces (modals, tables, badges)
    pages/           # route-level pages
    hooks/           # data hooks (React Query) + logic
  components/        # shared UI kit (DataTable, Modal, Toast, …)
    data-table/      # server-driven table: pagination, sorting, filters
  layouts/           # dashboard shell, sidebar, header
  store/             # Redux slices + hooks
  hooks/             # shared hooks (debounce, media query, permission)
  utils/             # formatters, list helpers, storage
  routes/            # lazy routes + guards
```

Constraints: max ~300 lines per file, no business logic inside UI components,
no mock data — everything reads from the backend.

### Key components

- **`DataTable`** (`components/data-table/`) — server-side pagination/sorting/
  filtering with a `Showing 1–10 of 120 users` summary line, skeleton loaders,
  friendly empty states, and error states with retry.
- **Modal system** — accessible, animated, focus-trapped; used by all forms.
- **Toast system** — global, actioned via the `useToast` hook.
- **Status badges** — shared vocabulary mapped from backend statuses.
- **Sidebar** — collapsible groups, active states, drawer on mobile.
- **Permission-aware UI** — `usePermission` + route guards hide/disable actions the
  backend would reject.

## Scripts

| Command            | Purpose                              |
| ------------------ | ------------------------------------ |
| `npm run dev`      | Dev server with API proxy            |
| `npm run build`    | Type-check + production build (code-split) |
| `npm run preview`  | Serve the production build           |
| `npx tsc --noEmit` | Type-check only                      |

## Notes

- Every page is lazy-loaded; the vendor bundle is split in `vite.config.ts`.
- The dashboard query keys are stable (no `new Date()` inside keys) to avoid
  refetch loops that trip the backend rate limiter.
- The backend must be reachable for any real data; the login page works standalone
  and surfaces API errors gracefully when it isn't.
