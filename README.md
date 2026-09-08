# GoFurry Platform

A discovery and exchange platform for the furry ecosystem, centered on resource
knowledge. P0-0 establishes the engineering foundation; product domains and
authentication are outside this phase.

## Development

Requires Go 1.27.1+, Node 24 LTS and pnpm 10.11.0. Docker is required for disposable
infrastructure and image acceptance. The repository uses one Go module and a pnpm
workspace, with no global Go generator installation required.

```text
pnpm install --frozen-lockfile
pnpm generate
pnpm check
```

Prepare private `server/env/*.local` inputs as described in
[development.md](docs/development.md). Existing real files must be preserved. In
separate terminals, `pnpm dev:api`, `pnpm dev:admin-api`, `pnpm dev:worker`,
`pnpm dev:web` and `pnpm dev:admin` start the applications.

## Repository map

| Area | Responsibility |
| --- | --- |
| `apps/web` | Public Astro Node SSR with React islands |
| `apps/admin` | React/Vite SPA with TanStack Router/Query |
| `packages/api-client`, `packages/design` | Generated API facades and shared SCSS |
| `server` | API/Admin/Worker, drivers, jobs, migrations and generators |
| `contracts` | OpenAPI and durable engineering constraints |
| `deploy` | Server/Web/Admin/PostgreSQL image definitions |
| `docs` | Product, architecture, implementation and development guidance |

`dev` is the current integration branch; `main` is a stable release snapshot.
Publishing, merging to `main`, tagging and releasing require explicit instruction.

Start with [AGENTS.md](AGENTS.md), the [P0-0 specification](docs/implementation/p0-0-repository-bootstrap.md),
[product overview](docs/product/PRODUCT.md) and [architecture](docs/architecture/ARCHITECTURE.md).
See [CHANGELOG.md](CHANGELOG.md) and [deployment artifacts](deploy/README.md).

License: **AGPL-3.0-only**. See [LICENSE](LICENSE).
