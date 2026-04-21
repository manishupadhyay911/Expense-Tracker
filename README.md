# Expense Tracker

Monorepo for a production-ready expense tracker with:

- `backend/`: Go API
- `frontend/`: React + Vite UI

Development and implementation notes live in [`plan.md`](./plan.md).

## Repository Layout

- [`backend/`](./backend): Go API and tests
- [`frontend/`](./frontend): React + Vite app
- [`plan.md`](./plan.md): living implementation plan

## Local Development

### Backend

```sh
cd backend
GOCACHE=/tmp/expense-go-cache GOMODCACHE=/tmp/expense-go-modcache go test ./...
PORT=8080 go run ./cmd/api
```

The backend listens on `http://localhost:8080` by default.

### Frontend

```sh
cd frontend
npm install
npm run dev
```

The frontend listens on `http://localhost:5173` by default.

## Environment Variables

- `PORT`: backend port, defaults to `8080`
- `VITE_API_BASE_URL`: frontend API base URL, defaults to `/api`

For local development, the Vite dev server proxies `/api` to `http://localhost:8080`.
For deployment, the frontend is intended to be deployed separately from the backend, and `VITE_API_BASE_URL` should point at the deployed backend URL.
The frontend includes [`frontend/.env.example`](./frontend/.env.example) as a template for that value.

## Deployment

The frontend and backend are deployed as separate Vercel projects:

- Frontend project root: `frontend/`
- Backend project root: `backend/`

The frontend uses `frontend/vercel.json` to support SPA routing, and the backend uses `backend/api/index.go` as the Go function entrypoint.
In Vercel, set `VITE_API_BASE_URL` on the frontend project to the backend project URL plus `/api`.

Live deployments:

- Frontend alias: [https://frontend-kappa-three-94.vercel.app](https://frontend-kappa-three-94.vercel.app)
- Backend alias: [https://backend-sigma-ten-86.vercel.app](https://backend-sigma-ten-86.vercel.app)

The alias URLs are stable across deploys, while the underlying production deployment URLs may change when the project is redeployed.

## API Contract

### `GET /health`

Returns:

```json
{"status":"ok"}
```

### `POST /expenses`

Headers:

- `Content-Type: application/json`
- `Idempotency-Key: <unique key>`

Request body:

```json
{
  "amount": "12.50",
  "category": "Food",
  "description": "Lunch",
  "date": "2026-04-21"
}
```

Returns an expense object with:

- `id`
- `amount`
- `category`
- `description`
- `date`
- `created_at`

### `GET /expenses`

Optional query parameters:

- `category`
- `sort=date_desc`

Returns a JSON array of expenses sorted newest-first by date when requested.
Use `sort=date_asc` for oldest-first ordering.

## Notes

- The backend currently uses in-memory persistence.
- Idempotency is enforced with the `Idempotency-Key` header on create requests.
- The frontend stores pending submissions locally so retries can safely reuse the same key.

## Design Notes

- Key design decisions: a Go API plus React + Vite frontend, money stored as currency-safe minor units, idempotent create requests to make retries safe, and separate frontend/backend deployments on Vercel.
- Timebox trade-offs: in-memory persistence was chosen to keep the first version small and fast to ship, and the UI stays intentionally simple rather than adding a heavier state or styling system.
- Intentionally not done: serving the frontend from the backend, persistent storage, authentication, multi-user support, background jobs, advanced analytics, a full deployment pipeline beyond the Vercel-ready structure, separate feature branches, branch protection rules, and CODEOWNERS-based approval flow.
- Future scope: support multiple currencies, paginated expense responses, richer filters, and export/reporting features once the core flow is stable.
