-- MCP Server CRUD

-- List MCP servers for a workspace
-- name: ListMCPServers :many
SELECT * FROM workspace_mcp_server WHERE workspace_id = $1 ORDER BY name;

-- Get a single MCP server by ID
-- name: GetMCPServer :one
SELECT * FROM workspace_mcp_server WHERE id = $1;

-- Create a new MCP server
-- name: CreateMCPServer :one
INSERT INTO workspace_mcp_server (workspace_id, name, description, config, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- Update an MCP server
-- name: UpdateMCPServer :one
UPDATE workspace_mcp_server
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    config = COALESCE(sqlc.narg('config'), config),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- Delete an MCP server
-- name: DeleteMCPServer :exec
DELETE FROM workspace_mcp_server WHERE id = $1;

-- Agent-MCP Binding CRUD

-- List MCP bindings for an agent
-- name: ListAgentMCPBindings :many
SELECT * FROM agent_mcp_binding WHERE agent_id = $1 ORDER BY sort_order;

-- Replace all bindings for an agent (delete existing, insert new)
-- name: DeleteAgentMCPBindings :exec
DELETE FROM agent_mcp_binding WHERE agent_id = $1;

-- name: CreateAgentMCPBinding :one
INSERT INTO agent_mcp_binding (agent_id, mcp_server_id, enabled, sort_order)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- Load full MCP definitions for a specific agent (for daemon claim path)
-- name: GetBoundMCPServersForAgent :many
SELECT ms.* FROM workspace_mcp_server ms
JOIN agent_mcp_binding ab ON ab.mcp_server_id = ms.id
WHERE ab.agent_id = $1 AND ab.enabled = true
ORDER BY ab.sort_order;
