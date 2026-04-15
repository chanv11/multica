-- Workspace MCP: reusable MCP server definitions scoped to a workspace,
-- and agent-to-MCP bindings.

CREATE TABLE IF NOT EXISTS workspace_mcp_server (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    config JSONB NOT NULL,
    created_by UUID REFERENCES member(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, name)
);

CREATE INDEX IF NOT EXISTS idx_workspace_mcp_server_workspace ON workspace_mcp_server(workspace_id);

CREATE TABLE IF NOT EXISTS agent_mcp_binding (
    agent_id UUID NOT NULL REFERENCES agent(id) ON DELETE CASCADE,
    mcp_server_id UUID NOT NULL REFERENCES workspace_mcp_server(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (agent_id, mcp_server_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_mcp_binding_mcp_server ON agent_mcp_binding(mcp_server_id);
