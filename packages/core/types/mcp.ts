export type MCPServerConfig = {
  command: string;
  args?: string[];
  env?: Record<string, string>;
};

export type MCPServersConfig = Record<string, MCPServerConfig>;