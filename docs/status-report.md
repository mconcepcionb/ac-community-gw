# Informe de estado — `ac-community-gw`

**Fecha:** 2026-09-20 · **Repo:** `ac-community-gw` · **Rama:** `main` (sin remoto configurado) · **Tipo:** informe puntual (point-in-time), no documento estable.

---

## 1. Resumen ejecutivo

`ac-community-gw` es una **API REST (Go 1.27)** que actúa de pasarela entre aplicaciones de comunidad y **AzerothCore**, ocultando su interfaz administrativa SOAP y sus comandos CLI. Se complementa con una **SPA (React 19 + TypeScript + Vite)** como primer consumidor del contrato.

Arquitectura de **monolito modular**: `core` + plugins compilados en un único binario, con fronteras de importación **verificadas por tests**. El backend expone **65 rutas / 77 operaciones** documentadas en OpenAPI 3.1, con cliente TypeScript generado. La SPA tiene **dos superficies** (portal de jugador y consola de staff) con 35 rutas y 16 módulos de features.

**Estado de verificación (ejecutado durante la elaboración del informe):**

| Comprobación | Resultado |
| --- | --- |
| `go test ./...` | ✅ Todo pasa (exit 0) |
| `go vet ./...` | ✅ Limpio |
| Cobertura unitaria Go | ⚠️ 36.6% global (80–100% en paquetes core; 0% en `repository/generated`) |
| `pnpm test` (SPA) | ✅ 33 archivos / 83 tests |
| `pnpm lint` (Biome) | ✅ Limpio (194 archivos) |
| `pnpm typecheck` (tsc) | ✅ Limpio |
| `task openapi:check` | Pipeline de drift configurado (no re-ejecutado por requerir toolchain completa) |
| Tests de integración | ⏸️ Requieren PostgreSQL + MariaDB (no ejecutados) |

---

## 2. Arquitectura

```
Consumidores (web, bots)
        │ REST/HTTPS
        ▼
  ac-community-gw  (core + plugins)
        │
        ▼
AzerothCore SOAP · Discord · PostgreSQL
```

- **`cmd/server`** — único punto que importa `core` + plugins concretos (wiring).
- **`internal/core`** — infraestructura transversal sin conocimiento de dominio (config, HTTP+middleware, RBAC, registries de comandos/servicios/eventos, auditoría, readiness).
- **`internal/plugins`** — 11 módulos de dominio.
- **`internal/adapters`** — PostgreSQL, AzerothCore SOAP, MySQL de Azeroth, Discord.
- **Reglas impuestas por `internal/architecture/architecture_test.go`:** `core` no importa plugins; plugins no importan otros plugins ni adapters. Comunicación entre plugins solo vía registries (Query/Command/Event/Permission).

**Decisiones documentadas:** 14 ADRs (modular monolith, `net/http` puro, PostgreSQL+sqlc+goose, sesiones server-side, permisos por módulo, SPA desacoplada, namespaces `gw`/`azeroth`, etc.).

---

## 3. Funcionalidades

### API (77 operaciones, 65 paths)

**Identidad y autenticación**
- OAuth Discord (`/auth/discord/login|callback`), sesiones server-side con cookie `HttpOnly`, `/logout`, `GET /api/v1/me` (perfil + roles + permisos efectivos).
- Sincronización de roles Discord, mappings rol→permisos, provisioning de usuarios.

**AzerothCore — portal de jugador**
- `me/characters`, visibilidad pública por personaje (opt-in), `me/mail` (auto-envío), `me/account` y **claim de cuenta existente con código in-game**.
- Leaderboards con opt-in público, estado de servidor.

**AzerothCore — administración / game master**
- Cuentas: crear, password, email, ban/unban, gmlevel, links de cuenta.
- Personajes: búsqueda global, detalle con equipo, ban/unban, correo de staff.
- Moderación en vivo: online, kick, mute/unmute, anuncios.
- Items con tooltips al estilo in-game, cola de claims.

**Gateway / operaciones**
- Store: catálogo de productos, wallets, órdenes, compra, grant de puntos, resolución (refund/retry).
- Reports de jugadores, vistas de moderación unificada.
- Auditoría persistente + visor, anotaciones de staff, catálogo de permisos.
- API keys / service accounts con scopes.
- **User 360** (agregado de usuario de comunidad).
- Endpoints públicos anónimos: `public/status`, `public/leaderboards/{board}`.
- Infra: `/healthz`, `/readyz`, `/metrics` (opcional, protegido).

### SPA (35 rutas, 16 features)

- **Portal (`/`):** dashboard, perfil, onboarding, personajes + detalle, correo, tienda + detalle, wallet, status, leaderboards, reportes.
- **Consola (`/admin/*`):** overview, usuarios + 360, cuentas, personajes, items, online, store (catálogo, wallets, órdenes), moderación, roles, auditoría, API clients, anotaciones.
- UI permission-driven: landings distintas según permisos, `<Can>`/`<PermissionGate>`, `/forbidden`, nav separada por superficie.
- Stack: TanStack Router + Query + Table, react-hook-form + Zod, Tailwind v4 + shadcn/Radix, cliente `@hey-api/openapi-ts`, MSW para tests.

---

## 4. Calidad de código

**Puntos fuertes**
- **0 TODO/FIXME/HACK** en código de producción (verificado, case-sensitive).
- Fronteras arquitectónicas **con tests** (no solo convención); incluye chequeo de `gofmt`.
- Sin frameworks pesados por decisión: `net/http`, `database/sql`+sqlc, `log/slog`, sin ORM ni DI. SQL explícito.
- Seguridad cuidada y documentada: validación `Safe*` de valores antes de construir comandos, rechazo de saltos de línea, body limits, rate-limit por IP con trusted proxies, headers de Caddy (CSP/HSTS/nosniff), arranque que rechaza configs inseguras en producción.
- Contrato único: anotaciones Go → `swagger.yaml` → cliente TS + Zod, con **drift check** en CI.
- Naming y separación limpios: `domain/`, `repository/` (sqlc), `handlers`, `plugin.go`, `permissions.go` por plugin.

**Debilidades**
- **Cobertura global 36.6%.** Muchos paquetes con tests fuertes (`azerothcore` 93%, `events` 90%, `config` 85%, SOAP 80.6%, Discord 75.9%, httpapi 78.2%), pero **plugins clave bajos** (`reports` 24%, `azerothaccount` 36%, `store` 38%, `azerothcharacter` 36%) y **`repository`/`generated` a 0%** (arrastra la media).
- No hay `golangci-lint`; el linting Go se limita a `go vet` + `gofmt` (la SPA sí tiene Biome).
- Tests de integración con DB no ejecutados en esta revisión (hay que confiar en CI con servicios).

---

## 5. Metodología

Es lo más destacable del repo: **desarrollo guiado por planes con tickets atómicos y gates**.

- `docs/plan/` contiene **129 tickets** organizados en planes: `foundation`, `discord-auth`, `spa` (36 tickets), `portal` (20), `ux-improvements` (11), `gateway-admin` (9), y `review-remediation` (**42 tickets**).
- El plan de remediación define **principios** (ticket atómico, gates entre tickets, test-first para defectos, migraciones aditivas/reversibles, docs junto al comportamiento) y **fases con gates** (G0 baseline → G1 seguridad → G2 correctness → G3 hardening → G4 data → G5 infra → G6 frontend → G7 docs), con **matriz de trazabilidad**.
- Separación docs clara: `docs/*.md` = hoy, `docs/ADR/` = por qué, `docs/plan/` = qué, `docs/runbooks/` = operación (10 runbooks).
- **Conventional Commits** (`feat`, `fix`, `refactor`, `docs`, `build`) con referencia al ticket (`portal 004`, `gateway-admin 009`).
- CI (`ci.yml`): job Go (`fmt:check`, `vet`, `test`, `test:race`, `openapi:check`) y job web (`lint`, `typecheck`, `test`).
- `Taskfile` como única interfaz de desarrollo (60+ tareas).

---

## 6. Tiempo invertido y productividad

Todos los commits son del **2026-09-20**, autor único **Marce Concepción**. No hay historial previo ni ramas.

| Métrica | Valor |
| --- | --- |
| Commits | **59** (`feat` 38, `refactor` 9, `docs` 8, `fix` 2, `build` 2) |
| Ventana de actividad | **09:42 → 22:48** (13 h 06 min de reloj de pared) |
| Ventanas activas | 09:42–14:07 (4h25m) + 19:22–22:48 (3h26m) = **≈7 h 51 min** |
| Mayor ráfaga | 22:00–22:48 → **14 commits en 48 min** (split gateway-admin) |
| Archivos versionados | **635** |
| Líneas | **+79.680 / −4.753** (neto 74.927 en el árbol) |

**Desglose de las ~75k líneas:**

| Categoría | LOC aproximadas |
| --- | --- |
| Go escrito a mano (excl. generated) | 21.368 (de los cuales **5.915 son tests**, 270 funciones) |
| Go generado (sqlc) | 3.237 |
| TS/TSX SPA a mano (excl. generated) | 10.641 |
| TS generado (cliente OpenAPI + Zod) | 8.909 |
| `swagger.yaml` | 4.901 |
| Docs `docs/` (173 `.md`, incl. 129 tickets + 14 ADR) | 8.413 |
| Migraciones SQL (14) | 319 |
| `pnpm-lock.yaml` | 5.851 |

**Productividad aproximada**
- Bruta (todo lo commiteado): **≈9.500 LOC/h** sobre 7,85 h activas.
- Código escrito a mano (Go+web+migraciones, incl. tests): **≈4.200 LOC/h**.
- Ritmo de commits: **≈7,5 commits/h** activa; la primera hora (09:42–09:43) ya introduce el 68% de las líneas (bootstrap inicial de 43k líneas en 4 commits).

⚠️ **Nota de interpretación:** estas cifras no son comparables con desarrollo manual tradicional. El volumen está dominado por *scaffolding generado* (sqlc, cliente OpenAPI, shadcn/ui, lockfile) y por documentación/planes extensos, lo que es coherente con un flujo asistido por IA + plantillas. La métrica significativa es **~32k LOC de código propio funcional en una jornada**, con tests, CI y docs completos.

---

## 7. Riesgos y recomendaciones

1. **Subir cobertura en plugins de dominio** (`reports`, `azerothaccount`, `azerothcharacter`, `store`) e ignorar `repository/generated` en el cálculo de cobertura para que refleje realidad.
2. **Ejecutar la suite de integración** (Postgres + MariaDB) con regularidad; hay muchos tests `_integration_test.go` sin ejecutar en la verificación local.
3. **Añadir `golangci-lint`** al gate Go (hoy solo vet + gofmt).
4. **Configurar remoto y CI real** (el repo no tiene `origin`); verificar que `task openapi:check`, `fmt:check` y `test:race` pasan en verde antes de confiar en el gate.
5. **Pendientes operativos declarados en el propio repo:** rotar el `Discord client secret` y el password SOAP que estuvieron en `.env` local; entregar `POSTGRES_PASSWORD`/`MARIADB_*` a Compose; pinning de digest de imágenes base.
6. **Riesgo de mantenibilidad** por concentración en un solo autor/jornada: el valor del repo reside en su documentación y tests; conviene asegurar que otra persona pueda correr `task check` end-to-end.

---

## 8. Veredicto

Proyecto **maduro en diseño y disciplina** (arquitectura verificada, contrato generado, seguridad tratada, metodología por tickets con gates) y **funcionalmente amplio** (77 operaciones, dos superficies SPA completas). Su principal debilidad medible es la **cobertura de tests del backend (36.6%)** y la **falta de ejecución de integración**, más aspectos operativos pendientes. El esfuerzo concentrado sugiere un MVP muy completo generado en una sola jornada intensiva, listo para endurecer con más tests de integración y CI estable.
