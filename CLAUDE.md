# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Backend (Go)

```bash
go build -o http2smtp ./        # Build
go build ./...                  # Compile-check everything
go vet ./...
go test ./...                   # Run all tests
go test ./util -run TestParserCron   # Run a single test
```

The go.mod carries `replace github.com/chuccp/go-web-frame => ../go-web-frame`, so the
sibling checkout is what actually gets compiled — editing it takes effect immediately,
and a version bump in go.mod alone changes nothing.

Run the server:

```bash
./http2smtp                                    # Defaults: 12566 (mgmt) + 12567 (API)
./http2smtp -web_port 12566 -api_port 12567    # Custom ports
./http2smtp -storage_root /path/to/storage     # Custom storage
```

The port flags also mark the instance as running in Docker (`core.isdocker`), which the
setup wizard uses to hide the port fields.

Stop a running instance:

```bash
taskkill /F /IM http2smtp.exe                   # Windows
lsof -ti :12566 | xargs -I {} kill -9 {}        # macOS / Linux
```

Cross-compile and Docker:

```bash
make release-linux      # → build_assets/http2smtp
make release-osx        # → build_assets/http2smtp
cd docker && docker compose up -d
docker build -t http2smtp .
```

### Frontend (`view/`, Vue 3 + Vite)

```bash
cd view && npm install
cd view && npm run dev          # Dev server on :3000, proxies /api → localhost:12566
cd view && npm run type-check   # vue-tsc --noEmit
cd view && npm run build        # vue-tsc && vite build — must stay green
cd view && npm run build:force  # vite build only
```

## Architecture

HTTP2SMTP is an HTTP-to-SMTP gateway: configure SMTP servers through the web UI, then send
mail over HTTP from your own code.

### Dual server

- **Management server** (`manage.port`, default 12566, context path `/api`) — web UI plus
  admin APIs. Authenticated by a session cookie.
- **Public API server** (`api.port`, default 12567, no context path) — only `/sendMail`.
  Authenticated by a token from the `t_token` table, passed as a query/body param. No
  session cookie is involved, so the two auth paths never mix.

`rest/API` is the only REST type on the public server; everything else lives on the
management server. Keeping them separate is what lets you expose mail-sending publicly
while the admin surface stays private.

### Web UI hosting

The management server serves the built frontend itself, so no nginx or reverse proxy is
needed. `manage.webPath` (default `web`) points at the built bundle and `main.go` hands it
to the server group as `web.ServerConfig{Locations, Page404: "index.html"}`. API routes
stay under `/api`; everything else falls through to the static handler, which serves the
bundle from the site root and returns `index.html` for unmatched HTML requests so
client-side routes survive a refresh or deep link.

The static handler runs as a no-route fallback, so it sits outside the auth filter — the
bundle (login page included) must load unauthenticated. Every `/api/*` route is still
authorised normally.

### Backend layout

| Path | Role |
|---|---|
| `main.go` | Wires both rest groups, models, services and the cron runner |
| `rest/` | HTTP handlers. `Init(context)` registers routes; protect with `.WithMeta(auth2.WithLogin())` |
| `service/` | Business logic: `TokenService`, `ScheduleService`, `LogService`, `SmtpService`, `UserService` |
| `model/` | GORM models plus the `Config` struct mapped to `config.ini` |
| `entity/` | Request/response DTOs and the log status constants |
| `auth/` | Session cookie (`authentication.go`) and the AES-CBC helpers (`token.go`) |
| `smtp/` | Delivery via `wneessen/go-mail`; `*2` suffixed functions are the live ones |
| `runner/` | `ScheduleRunner` — reloads scheduled tasks into cron every minute |
| `db/` | Builds the GORM connection from config (sqlite or mysql) |
| `util/` | String, time, template and HTTP helpers |

Tables: `t_user`, `t_SMTP`, `t_mail`, `t_token`, `t_schedule`, `t_log`.

### Sessions and authorisation

The session is an AES-CBC-encrypted `LoginUser` carried in an **HttpOnly cookie named
`user_token`**, validated against `t_user.salt`. `Salt` is a session *epoch*, not part of
password hashing (bcrypt carries its own salt): the token embeds the salt it was minted
with, and `auth.User` rejects it once the stored value no longer matches.

Salt is therefore rotated whenever sessions must end — **sign-out, password change, and
user disable**. That means signing out ends every session for that user, on every device.
`auth.User` also accepts the token via an `Authorization` header, which the browser never
uses but external callers can.

Ownership: non-admin users only see their own rows, filtered by `user_id`. Admin-only
endpoints re-check `IsAdmin` in the handler. `PUT` handlers verify ownership and preserve
the original owner, so an admin editing someone else's row does not take it over.

### Configuration

Precedence is defaults → `config.ini` → CLI flags → environment variables.

> **INI values load as strings, and the decoder only weak-converts when the target is a
> struct.** Two consequences that have bitten this codebase:
>
> - Read config through `model.GetConfig()` / `model.CoreState()`. `IConfig.GetBoolOrDefault("core.init")`
>   only accepts a real bool, so against an INI-sourced `init = true` it silently returns the
>   default — which once left the setup guards dead and stopped the database from attaching
>   at startup. Pass the `*Config` itself, never `&cfg`: a `**Config` makes the decoder fall
>   back to `encoding/json`, which cannot coerce `"true"` → bool or `"12566"` → int.
> - A `[]string` field cannot be set from an INI file at all (`cannot convert string to
>   []string`), and a bad value crashes startup rather than reporting a config error.

## Frontend (`view/`)

Vue 3 + Element Plus + Pinia + vue-i18n, four locales (`zh-cn`, `en`, `zh-tw`, `ja`) that
must stay in key-for-key sync.

- `src/api/request.ts` — axios on `/api`, `withCredentials: true`, no credentials are
  attached by hand: the session cookie rides along automatically. Non-`0`/`200` codes and
  HTTP errors both reject, and 401 clears the local state and redirects to `/login`.
- `src/store/auth.ts` — holds **display state only** (username, `isAdmin`, a `loggedIn`
  flag). The session credential is HttpOnly and deliberately unreadable from script, so
  this flag drives navigation only; every request is authorised server-side.
  `views/Index.vue` hydrates it from `GET /api/set`, which reports `hasLogin`,
  `username` and `isAdmin`.
- `src/router/index.ts` — guard on `meta.requiresAuth` / `meta.requiresAdmin`.
- `src/views/settings/index.vue` — admin-only; ports and database config.
- `vite.config.ts` — dev proxy for `/api`. Keep `VITE_API_BASE_URL` relative so requests
  stay same-origin; pointing it at the backend directly is cross-origin and the session
  cookie will be dropped.

### Release packaging

`view/` is the frontend that ships. The release workflow runs `npm ci` and `npm run build`
in `view/`, then packages `view/dist` as `web/` in the release tarball. Because
`npm run build` runs `vue-tsc` first, a type error fails the release rather than shipping.
(Older releases pulled a prebuilt frontend from `chuccp/d-mail-view`; that dependency is
gone, so `d-mail-view` is no longer part of the build.)

## Known gaps

- **Signing out ends every session for that user**, on all devices — it rotates the session
  salt (see above). Per-device revocation would need a server-side session store.
- **The public `/sendMail` endpoint has no rate limiting.** The framework ships a
  `ratelimit` component; it is not wired up here.
- **INI cannot express a `[]string` field.** `web.server.locations` therefore has to be set
  in code (see `main.go`); putting it in `config.ini` makes startup fail with
  `cannot convert string to []string`.
