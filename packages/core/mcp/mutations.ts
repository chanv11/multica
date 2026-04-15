import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api";
import { mcpKeys } from "./queries";
import type { CreateMCPServerRequest, UpdateMCPServerRequest, ReplaceAgentMCPBindingsRequest } from "../types";

export function useCreateMCPServer(wsId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateMCPServerRequest) => api.createMCPServer(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: mcpKeys.list(wsId) });
    },
  });
}

export function useUpdateMCPServer(wsId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & UpdateMCPServerRequest) =>
      api.updateMCPServer(id, data),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: mcpKeys.detail(wsId, vars.id) });
      qc.invalidateQueries({ queryKey: mcpKeys.list(wsId) });
    },
  });
}

export function useDeleteMCPServer(wsId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteMCPServer(id),
    onSuccess: (_data, id) => {
      qc.removeQueries({ queryKey: mcpKeys.detail(wsId, id) });
      qc.invalidateQueries({ queryKey: mcpKeys.list(wsId) });
    },
  });
}

export function useReplaceAgentMCPBindings(wsId: string, agentId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: ReplaceAgentMCPBindingsRequest) =>
      api.replaceAgentMCPBindings(agentId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: mcpKeys.list(wsId) });
      qc.invalidateQueries({ queryKey: mcpKeys.agentBindings(wsId, agentId) });
    },
  });
}
