# Workspace MCP Design

**Date:** 2026-04-15

**Goal:** Add workspace-shared MCP management for Claude Code agents on web, with agent-level binding and `${VAR}` resolution from the agent Environment tab.

## Constraints

- Workspace-shared resource model, not runtime-scoped resource model
- All runtimes in scope are Claude Code only
- Web UI only for phase 1; desktop UI stays unchanged
- Existing agent `Environment` remains the source of per-agent secrets and env vars
- MCP config supports both plaintext values and `${VAR}` placeholders

## Product Shape

### Workspace MCP Registry

Add a workspace-level MCP module where users can create and maintain reusable Claude MCP server definitions.

Each MCP entry stores:

- `name`
- `description`
- `config`

`config` is the Claude-native `mcpServers.<name>` object, for example:

```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-github"],
  "env": {
    "GITHUB_TOKEN": "${GITHUB_TOKEN}"
  }
}
```

### Agent Binding

Agents do not own full MCP config. Instead, each agent binds to one or more workspace MCP entries.

This gives:

- reuse across multiple agents
- one-place updates for shared MCP definitions
- clear “which agents use this MCP” relationships

### Environment Integration

The existing agent `Environment` tab remains the place to manage agent-specific env vars and secrets.

Resolution rule:

- plaintext string values remain unchanged
- strings matching `${VAR}` are resolved from the selected agent’s `custom_env`
- missing referenced variables fail task startup with an explicit error

## Data Model

### `workspace_mcp_server`

Workspace-owned reusable MCP definition.

Suggested fields:

- `id`
- `workspace_id`
- `name`
- `description`
- `config`
- `created_by`
- `created_at`
- `updated_at`

Suggested constraints:

- unique `(workspace_id, name)`
- `config` must be a JSON object

### `agent_mcp_binding`

Join table between agents and workspace MCP definitions.

Suggested fields:

- `agent_id`
- `mcp_server_id`
- `enabled`
- `sort_order`
- `created_at`

Suggested constraints:

- unique `(agent_id, mcp_server_id)`

## Backend API

### MCP Registry

- `GET /api/mcp-servers?workspace_id=...`
- `POST /api/mcp-servers`
- `GET /api/mcp-servers/:id`
- `PUT /api/mcp-servers/:id`
- `DELETE /api/mcp-servers/:id`

Validation:

- `config` must be an object
- `args`, if present, must be an array
- `env`, if present, must be an object

### Agent Bindings

- `GET /api/agents/:id/mcp-bindings`
- `PUT /api/agents/:id/mcp-bindings`

Binding update is replace-all:

```json
{
  "mcp_server_ids": ["mcp_1", "mcp_2"]
}
```

## Runtime Flow

The PR #754 daemon/output path is still useful, but its input source changes.

Instead of reading `agent.runtime_config.mcp_servers`, daemon startup should:

1. load the agent’s MCP bindings
2. load each bound workspace MCP definition
3. resolve `${VAR}` placeholders from `agent.custom_env`
4. aggregate into:

```json
{
  "mcpServers": {
    "...": {}
  }
}
```

5. write `.mcp.json`
6. launch Claude with `--mcp-config`

Security requirement:

- `.mcp.json` must be written with mode `0600`

Failure handling:

- if `${VAR}` is referenced but missing, fail task startup with a message that names the MCP, the variable, and the agent

## Web UI

### MCP Page

Add a web-only workspace MCP page.

Responsibilities:

- list MCP entries
- create MCP entry
- edit MCP entry
- delete MCP entry

Layout:

- left column: MCP list + search + create button
- right column: detail editor with `name`, `description`, `config`, and save/delete actions

Initial editing model:

- raw JSON textarea/editor for `config`
- explanatory text that `${VAR}` values resolve from the agent Environment tab

### Agent Detail MCP Tab

Add a web-only MCP binding tab to the agent detail flow.

Responsibilities:

- show bound MCPs
- add MCP binding from workspace registry
- remove MCP binding
- save the full binding set

This tab does not edit MCP definitions.

## Scope Exclusions

Not in phase 1:

- desktop entry points or desktop-specific UI
- provider abstraction beyond Claude Code
- MCP connection testing
- structured form builder for MCP config
- runtime-scoped MCP registry

## Recommended Delivery Order

1. Fix `.mcp.json` file mode to `0600`
2. Add DB tables and sqlc queries
3. Add MCP registry and binding APIs
4. Switch daemon claim/startup to binding-based aggregation
5. Add web MCP registry page
6. Add web agent MCP binding UI
