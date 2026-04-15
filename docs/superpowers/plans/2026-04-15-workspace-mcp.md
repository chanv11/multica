# Workspace MCP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build workspace-shared Claude MCP management on web, with agent binding, `${VAR}` resolution from agent Environment, and secure `.mcp.json` generation in daemon.

**Architecture:** Introduce a workspace MCP registry plus an agent-to-MCP binding table, then switch daemon startup from `agent.runtime_config.mcp_servers` to binding-driven aggregation. Web gets a new MCP management page plus an agent MCP binding tab, while desktop stays unchanged.

**Tech Stack:** Go, Chi, sqlc, PostgreSQL JSONB, React, Next.js App Router, TanStack Query, Zustand, shared `packages/core` and `packages/views`

---

## File Map

### Backend

- Modify: `server/internal/daemon/execenv/context.go`
- Modify: `server/internal/daemon/daemon.go`
- Modify: `server/internal/daemon/types.go`
- Modify: `server/internal/handler/agent.go`
- Modify: `server/internal/handler/daemon.go`
- Modify: `server/cmd/server/router.go`
- Modify: `server/pkg/db/queries/agent.sql`
- Add: `server/pkg/db/queries/mcp.sql`
- Add: `server/internal/handler/mcp.go`
- Add: `server/migrations/043_workspace_mcp.up.sql`
- Add: `server/migrations/043_workspace_mcp.down.sql`
- Modify: `server/pkg/db/generated/*.go`

### Frontend shared

- Add: `packages/core/types/mcp.ts`
- Modify: `packages/core/types/index.ts`
- Modify: `packages/core/api/client.ts`
- Add: `packages/core/mcp/queries.ts`
- Add: `packages/views/mcp/components/mcp-page.tsx`
- Add: `packages/views/mcp/components/mcp-list.tsx`
- Add: `packages/views/mcp/components/mcp-detail.tsx`
- Add: `packages/views/mcp/index.ts`
- Add: `packages/views/agents/components/tabs/mcp-tab.tsx`
- Modify: `packages/views/agents/components/agent-detail.tsx`

### Web-only

- Add: `apps/web/app/(dashboard)/mcp/page.tsx`
- Modify: `apps/web/app/(dashboard)/layout.tsx` or the web dashboard nav source that adds sidebar links

### Tests

- Add: `server/internal/handler/mcp_test.go`
- Modify: `server/internal/handler/agent_test.go`
- Modify: `server/internal/daemon/execenv/execenv_test.go`
- Add: `server/internal/daemon/mcp_test.go` or extend daemon tests around claim/build flow
- Add: `packages/views/mcp/components/mcp-page.test.tsx`
- Add: `packages/views/agents/components/tabs/mcp-tab.test.tsx`

## Task 1: Secure `.mcp.json` File Permissions

**Files:**
- Modify: `server/internal/daemon/execenv/context.go`
- Modify: `server/internal/daemon/execenv/execenv_test.go`

- [ ] **Step 1: Write the failing test**

Add a test asserting `.mcp.json` is written with mode `0600`.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd server && go test ./internal/daemon/execenv -run TestWriteMCPConfigClaude -v`
Expected: FAIL because the file mode is currently `0644` or the test does not exist yet.

- [ ] **Step 3: Write minimal implementation**

Change `os.WriteFile(path, data, 0o644)` to `os.WriteFile(path, data, 0o600)` in `writeMCPConfig`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd server && go test ./internal/daemon/execenv -run 'TestWriteMCPConfigClaude|TestWriteMCPConfigFileMode' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/internal/daemon/execenv/context.go server/internal/daemon/execenv/execenv_test.go
git commit -m "fix(daemon): secure mcp config permissions"
```

## Task 2: Add Database Tables and Queries for Workspace MCP

**Files:**
- Add: `server/migrations/043_workspace_mcp.up.sql`
- Add: `server/migrations/043_workspace_mcp.down.sql`
- Add: `server/pkg/db/queries/mcp.sql`
- Modify: `server/pkg/db/generated/*.go`

- [ ] **Step 1: Write the failing schema/query tests or integration assertions**

Add or extend DB-backed handler/integration tests to expect workspace MCP CRUD and agent binding persistence.

- [ ] **Step 2: Run targeted tests to verify they fail**

Run: `cd server && go test ./internal/handler -run 'TestMCP|TestAgentMCP' -v`
Expected: FAIL because tables/queries do not exist yet.

- [ ] **Step 3: Write minimal migration and sqlc queries**

Create:
- `workspace_mcp_server`
- `agent_mcp_binding`

Add queries for:
- list/get/create/update/delete MCP servers
- list/set agent MCP bindings
- load bound MCPs for daemon claim path

- [ ] **Step 4: Regenerate sqlc output**

Run: `make sqlc`
Expected: generated query/types files update cleanly.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd server && go test ./internal/handler -run 'TestMCP|TestAgentMCP' -v`
Expected: PASS or proceed once later handler logic lands if schema-only assertions are split.

- [ ] **Step 6: Commit**

```bash
git add server/migrations/043_workspace_mcp.up.sql server/migrations/043_workspace_mcp.down.sql server/pkg/db/queries/mcp.sql server/pkg/db/generated
git commit -m "feat(db): add workspace mcp registry schema"
```

## Task 3: Add MCP Registry HTTP Handlers

**Files:**
- Add: `server/internal/handler/mcp.go`
- Modify: `server/cmd/server/router.go`
- Add: `server/internal/handler/mcp_test.go`

- [ ] **Step 1: Write the failing handler tests**

Cover:
- create MCP server
- update MCP server
- list MCP servers
- reject invalid `config`
- reject delete when still bound to agents if that rule is chosen

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd server && go test ./internal/handler -run TestMCP -v`
Expected: FAIL because routes/handlers are missing.

- [ ] **Step 3: Write minimal implementation**

Implement Chi handlers under `/api/mcp-servers` with workspace membership checks matching existing settings resources.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd server && go test ./internal/handler -run TestMCP -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/internal/handler/mcp.go server/internal/handler/mcp_test.go server/cmd/server/router.go
git commit -m "feat(api): add workspace mcp registry endpoints"
```

## Task 4: Add Agent MCP Binding API

**Files:**
- Modify: `server/internal/handler/agent.go`
- Modify: `server/internal/handler/agent_test.go`
- Modify: `server/cmd/server/router.go`

- [ ] **Step 1: Write the failing tests**

Cover:
- get current bindings
- replace bindings with a new list
- reject MCP IDs outside the agent workspace

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd server && go test ./internal/handler -run TestAgentMCPBindings -v`
Expected: FAIL because binding endpoints are missing.

- [ ] **Step 3: Write minimal implementation**

Add:
- `GET /api/agents/:id/mcp-bindings`
- `PUT /api/agents/:id/mcp-bindings`

Use replace-all semantics for `mcp_server_ids`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd server && go test ./internal/handler -run TestAgentMCPBindings -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/internal/handler/agent.go server/internal/handler/agent_test.go server/cmd/server/router.go
git commit -m "feat(agent): add mcp binding endpoints"
```

## Task 5: Switch Daemon to Binding-Driven MCP Aggregation

**Files:**
- Modify: `server/internal/handler/daemon.go`
- Modify: `server/internal/daemon/types.go`
- Modify: `server/internal/daemon/daemon.go`
- Modify: `server/internal/daemon/execenv/context.go`
- Modify: `server/internal/daemon/execenv/execenv_test.go`
- Add or Modify: `server/internal/daemon/mcp_test.go`

- [ ] **Step 1: Write the failing tests**

Cover:
- claim response includes aggregated MCP config from bindings
- `${VAR}` resolves from `CustomEnv`
- missing `${VAR}` returns a clear startup error
- multiple bound MCPs merge under `mcpServers`

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd server && go test ./internal/daemon/... ./internal/handler/... -run 'Test.*MCP' -v`
Expected: FAIL because daemon still reads `runtime_config.mcp_servers`.

- [ ] **Step 3: Write minimal implementation**

Replace the old extraction path with:
- query agent bindings in claim flow
- include bound MCP definitions in task payload
- resolve `${VAR}` from `task.Agent.CustomEnv`
- write final `.mcp.json`

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd server && go test ./internal/daemon/... ./internal/handler/... -run 'Test.*MCP' -v`
Expected: PASS

- [ ] **Step 5: Run broader backend verification**

Run:
- `cd server && go test ./...`
- `cd server && go vet ./...`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/internal/handler/daemon.go server/internal/daemon/types.go server/internal/daemon/daemon.go server/internal/daemon/execenv/context.go server/internal/daemon/execenv/execenv_test.go server/internal/daemon/mcp_test.go
git commit -m "feat(daemon): build claude mcp config from agent bindings"
```

## Task 6: Add Shared Core MCP Types and API Client

**Files:**
- Add: `packages/core/types/mcp.ts`
- Modify: `packages/core/types/index.ts`
- Modify: `packages/core/api/client.ts`
- Add: `packages/core/mcp/queries.ts`

- [ ] **Step 1: Write the failing frontend tests or type usage sites**

Create tests or usage scaffolding that requires MCP types and API helpers.

- [ ] **Step 2: Run targeted frontend tests to verify they fail**

Run: `pnpm --filter @multica/views exec vitest run packages/views/mcp/components/mcp-page.test.tsx`
Expected: FAIL because shared MCP client/types do not exist.

- [ ] **Step 3: Write minimal implementation**

Add:
- MCP entity types
- MCP CRUD client methods
- agent MCP binding client methods
- React Query options helpers

- [ ] **Step 4: Run typecheck/tests to verify they pass**

Run:
- `pnpm typecheck`

Expected: PASS for the touched packages or only remaining failures unrelated to MCP.

- [ ] **Step 5: Commit**

```bash
git add packages/core/types/mcp.ts packages/core/types/index.ts packages/core/api/client.ts packages/core/mcp/queries.ts
git commit -m "feat(core): add mcp client and query types"
```

## Task 7: Build Web MCP Registry Page

**Files:**
- Add: `packages/views/mcp/components/mcp-page.tsx`
- Add: `packages/views/mcp/components/mcp-list.tsx`
- Add: `packages/views/mcp/components/mcp-detail.tsx`
- Add: `packages/views/mcp/index.ts`
- Add: `apps/web/app/(dashboard)/mcp/page.tsx`
- Modify: web sidebar/nav file
- Add: `packages/views/mcp/components/mcp-page.test.tsx`

- [ ] **Step 1: Write the failing UI tests**

Cover:
- list MCP entries
- create new MCP entry
- edit JSON config
- save success/error flows

- [ ] **Step 2: Run tests to verify they fail**

Run: `pnpm --filter @multica/views exec vitest run mcp-page.test.tsx`
Expected: FAIL because MCP page components do not exist.

- [ ] **Step 3: Write minimal implementation**

Implement a two-panel page matching existing Multica list/detail patterns. Use a textarea or editor for raw JSON config in phase 1.

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm --filter @multica/views exec vitest run mcp-page.test.tsx`
Expected: PASS

- [ ] **Step 5: Run typecheck for web/shared packages**

Run: `pnpm typecheck`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add packages/views/mcp apps/web/app/\(dashboard\)/mcp/page.tsx
git commit -m "feat(web): add workspace mcp management page"
```

## Task 8: Add Agent MCP Binding Tab in Web Flow

**Files:**
- Add: `packages/views/agents/components/tabs/mcp-tab.tsx`
- Modify: `packages/views/agents/components/agent-detail.tsx`
- Add: `packages/views/agents/components/tabs/mcp-tab.test.tsx`

- [ ] **Step 1: Write the failing UI tests**

Cover:
- render existing bindings
- add an available MCP
- remove a bound MCP
- save replace-all bindings
- show explanatory text for `${VAR}` resolution via Environment

- [ ] **Step 2: Run tests to verify they fail**

Run: `pnpm --filter @multica/views exec vitest run mcp-tab.test.tsx`
Expected: FAIL because the tab does not exist.

- [ ] **Step 3: Write minimal implementation**

Add a new `MCP` tab in agent detail for web-visible binding management. Reuse shared query/client layer and keep MCP editing out of this tab.

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm --filter @multica/views exec vitest run mcp-tab.test.tsx`
Expected: PASS

- [ ] **Step 5: Run broader frontend verification**

Run:
- `pnpm test`
- `pnpm typecheck`

Expected: PASS or only unrelated pre-existing failures.

- [ ] **Step 6: Commit**

```bash
git add packages/views/agents/components/tabs/mcp-tab.tsx packages/views/agents/components/agent-detail.tsx packages/views/agents/components/tabs/mcp-tab.test.tsx
git commit -m "feat(agents): add mcp binding tab"
```

## Task 9: Final Verification and Docs Touch-Up

**Files:**
- Modify: any touched docs if needed

- [ ] **Step 1: Run end-to-end verification commands**

Run:
- `cd server && go test ./...`
- `cd server && go vet ./...`
- `pnpm test`
- `pnpm typecheck`

Expected: PASS

- [ ] **Step 2: Manual smoke test**

Verify in web:
- create workspace MCP
- bind MCP to agent
- set required env var in Environment
- run a task and confirm `.mcp.json` is created and Claude receives the MCP

- [ ] **Step 3: Commit final fixes**

```bash
git add -A
git commit -m "chore(mcp): finalize workspace mcp integration"
```
