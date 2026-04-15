export type MCPServerConfig = {
  command: string;
  args?: string[];
  env?: Record<string, string>;
};

export type MCPServersConfig = Record<string, MCPServerConfig>;

// MCP server entity from the API
export type MCPServer = {
  id: string;
  workspace_id: string;
  name: string;
  description: string;
  config: Record<string, unknown>;
  created_by: string | null;
  created_at: string;
  updated_at: string;
};

// Create MCP server request
export type CreateMCPServerRequest = {
  name: string;
  description?: string;
  config?: Record<string, unknown>;
};

// Update MCP server request
export type UpdateMCPServerRequest = {
  name?: string;
  description?: string;
  config?: Record<string, unknown>;
};

// Agent MCP binding
export type AgentMCPBinding = {
  agent_id: string;
  mcp_server_id: string;
  enabled: boolean;
  sort_order: number;
  created_at: string;
};

// Replace agent bindings request
export type ReplaceAgentMCPBindingsRequest = {
  mcp_server_ids: string[];
};