# Expense Tracker Monorepo Plan

## Summary
Create a new monorepo called `Expense Tracker` with two separate applications:
- `backend/`: a Go API for expense storage and retrieval
- `frontend/`: a React + Vite UI for creating, filtering, sorting, and totaling expenses

The backend will use lightweight in-memory persistence for this iteration, with idempotency support so repeated submits do not create duplicate expenses during a running session.
The end goal is a production-ready deployment on Vercel, with local development mirroring the same API and frontend contract as closely as possible.

## Key Changes
- Workspace layout
  - Add a fresh monorepo structure with independent `backend/` and `frontend/` apps under one root.
  - Leave the current CRUD sample untouched.
- Backend
  - Implement `POST /expenses` and `GET /expenses` in Go.
  - Model expenses with `id`, `amount`, `category`, `description`, `date`, and `created_at`.
  - Store money using a currency-safe integer representation internally.
  - Add idempotency handling for create requests using a request key kept in memory.
  - Support `category` filtering and both `sort=date_desc` and `sort=date_asc` on list requests.
- Frontend
  - Scaffold a React + Vite app in TypeScript.
  - Build a simple expense form, category filter, newest-first/oldest-first sort control, expense list/table, and a visible total for the filtered list.
  - Reuse a stable client-generated idempotency key for each pending create attempt so retries do not duplicate data.
- Integration
  - Keep the backend and frontend as separate dev apps that talk over HTTP.
  - Deploy the frontend and backend as separate Vercel projects.
  - Make the frontend Vercel-ready and keep the backend compatible with Vercel deployment constraints.
  - Document local setup, environment variables, and the API contract in the README so both apps are easy to run together.

## Progress
- Done
  - Monorepo folder created in `Documents/Expense tracker`
  - GitHub repository connected and initial plan committed
  - Backend scaffolded and upgraded to a working expense API
  - Frontend scaffolded with Vite, React, and TypeScript
  - Frontend wired to the Go API with create/list/filter/sort flows and idempotency handling
  - Backend and frontend production builds verified locally
  - Local development and API contract documented in the README
  - Frontend summary view, stricter validation, and unit tests added
  - Backend tests and frontend tests passing
  - README design notes added for decisions, trade-offs, and intentional gaps
  - Separate Vercel deployment setup added for frontend and backend
- In progress
  - Production hardening and release checks

## Future Scope
- Support multiple currencies and currency-aware formatting.
- Add paginated expense responses for larger datasets.
- Introduce richer filtering, reporting, and export capabilities.
- Consider durable persistence once the core flow is stable.

## Production Readiness
- Deployment
  - Target Vercel as the primary hosting platform.
  - Ensure the frontend can be deployed directly to Vercel as a separate project.
  - Ensure the Go backend is structured in a Vercel-compatible way or a clearly documented Vercel-friendly deployment path as a separate project.
- Reliability
  - Add input validation, clear error responses, and deterministic request handling.
  - Preserve idempotency for create requests so retries and duplicate submits are safe.
  - Make the frontend resilient to refreshes, retries, and transient API failures.
- Operations
  - Add environment-based configuration for API URLs and runtime settings.
  - Include logging that is useful in production but not overly noisy.
  - Document the minimum operational assumptions in the README.
- Quality
  - Add automated tests for backend behavior and UI flows.
  - Prefer production-like code structure over one-off shortcuts.

## Test Plan
- Backend tests
  - Create expense success path.
  - Validation for required fields and bad amount/date input.
  - Idempotent retry behavior for repeated create requests.
  - Filtering and date-desc sorting.
  - CORS / API contract checks if the frontend is served from a different origin in development.
- Frontend checks
  - Form submission and error handling.
  - Filter/sort behavior.
  - Total updates when the visible list changes.
  - Retry behavior and idempotency key reuse across refreshes and repeated submits.
- End-to-end sanity
  - Run backend and frontend locally, create expenses, refresh the page, and confirm retries do not duplicate entries.
  - Verify the Vercel deployment path works in a staging-like environment before treating it as complete.

## Assumptions
- Use a monorepo rather than separate Git repositories.
- Keep persistence in-memory for now, so data does not survive process restarts.
- Avoid introducing a UI framework or state library unless it becomes necessary.
- Keep styling simple and functional rather than polished.
- Treat Vercel compatibility as a first-class requirement rather than an afterthought.
