import { queryOptions } from "@tanstack/react-query";
import { api } from "../api";

export const mcpKeys = {
  all: (wsId: string) => ["mcp-servers", wsId] as const,
  list: (wsId: string) => [...mcpKeys.all(wsId), "list"] as const,
  detail: (wsId: string, id: string) =>
    [...mcpKeys.all(wsId), "detail", id] as const,
};

export function mcpListOptions(wsId: string) {
  return queryOptions({
    queryKey: mcpKeys.list(wsId),
    queryFn: () => api.listMCPServers(),
  });
}

export function mcpDetailOptions(wsId: string, id: string) {
  return queryOptions({
    queryKey: mcpKeys.detail(wsId, id),
    queryFn: () => api.getMCPServer(id),
  });
}
