"use client";

import { useState, useMemo } from "react";
import { Plus, Trash2, Info, Loader2, Server } from "lucide-react";
import type { Agent } from "@multica/core/types";
import { useQuery } from "@tanstack/react-query";
import { mcpListOptions, agentMCPBindingsOptions } from "@multica/core/mcp/queries";
import { useReplaceAgentMCPBindings } from "@multica/core/mcp/mutations";
import { useWorkspaceId } from "@multica/core/hooks";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@multica/ui/components/ui/dialog";
import { Button } from "@multica/ui/components/ui/button";
import { toast } from "sonner";

interface MCPTabProps {
  agent: Agent;
  onSave?: (updates: Partial<Agent>) => Promise<void>;
}

export function MCPTab({ agent }: MCPTabProps) {
  const wsId = useWorkspaceId();
  const { data: availableServers = [] } = useQuery(mcpListOptions(wsId));
  const { data: bindings = [] } = useQuery(agentMCPBindingsOptions(wsId, agent.id));
  const replaceBindings = useReplaceAgentMCPBindings(wsId, agent.id);

  // Track bound server IDs locally for optimistic add/remove before save
  const boundServerIds = useMemo(
    () => new Set(bindings.map((b) => b.mcp_server_id)),
    [bindings],
  );
  const [localBoundIds, setLocalBoundIds] = useState<Set<string>>(boundServerIds);

  // Reset local state when server data arrives
  const [initialized, setInitialized] = useState(false);
  if (bindings.length > 0 && !initialized) {
    setLocalBoundIds(boundServerIds);
    setInitialized(true);
  } else if (bindings.length === 0 && !initialized && availableServers.length > 0) {
    // No bindings exist yet — local state starts empty
    setInitialized(true);
  }

  const currentlyBound = useMemo(
    () => availableServers.filter((s) => localBoundIds.has(s.id)),
    [availableServers, localBoundIds],
  );

  const unboundServers = useMemo(
    () => availableServers.filter((s) => !localBoundIds.has(s.id)),
    [availableServers, localBoundIds],
  );

  const dirty = useMemo(() => {
    const localArr = Array.from(localBoundIds).sort();
    const serverArr = Array.from(boundServerIds).sort();
    return JSON.stringify(localArr) !== JSON.stringify(serverArr);
  }, [localBoundIds, boundServerIds]);

  const [showPicker, setShowPicker] = useState(false);

  const handleAdd = (serverId: string) => {
    setLocalBoundIds(new Set([...localBoundIds, serverId]));
    setShowPicker(false);
  };

  const handleRemove = (serverId: string) => {
    const next = new Set(localBoundIds);
    next.delete(serverId);
    setLocalBoundIds(next);
  };

  const handleSave = async () => {
    try {
      await replaceBindings.mutateAsync({
        mcp_server_ids: Array.from(localBoundIds),
      });
      toast.success("MCP bindings saved");
    } catch {
      toast.error("Failed to save MCP bindings");
    }
  };

  const saving = replaceBindings.isPending;

  return (
    <div className="max-w-2xl space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold">MCP Servers</h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            Bind MCP servers from the workspace registry to this agent.
          </p>
        </div>
        <Button
          variant="outline"
          size="xs"
          onClick={() => setShowPicker(true)}
          disabled={saving || unboundServers.length === 0}
        >
          <Plus className="h-3 w-3" />
          Add Server
        </Button>
      </div>

      <div className="flex items-start gap-2 rounded-md border border-info/20 bg-info/5 px-3 py-2.5">
        <Info className="h-3.5 w-3.5 shrink-0 text-info mt-0.5" />
        <p className="text-xs text-muted-foreground">
          MCP servers provide tools and capabilities to the agent. Use {"${VAR}"} in
          MCP config to reference environment variables defined in the Environment tab.
        </p>
      </div>

      {currentlyBound.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12">
          <Server className="h-8 w-8 text-muted-foreground/40" />
          <p className="mt-3 text-sm text-muted-foreground">No MCP servers bound</p>
          <p className="mt-1 text-xs text-muted-foreground">
            Add MCP servers from the workspace registry to give this agent access to their tools.
          </p>
          {unboundServers.length > 0 && (
            <Button
              onClick={() => setShowPicker(true)}
              size="xs"
              className="mt-3"
              disabled={saving}
            >
              <Plus className="h-3 w-3" />
              Add Server
            </Button>
          )}
        </div>
      ) : (
        <div className="space-y-2">
          {currentlyBound.map((server) => (
            <div
              key={server.id}
              className="flex items-center gap-3 rounded-lg border px-4 py-3"
            >
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-muted">
                <Server className="h-4 w-4 text-muted-foreground" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-sm font-medium">{server.name}</div>
                {server.description && (
                  <div className="text-xs text-muted-foreground truncate">
                    {server.description}
                  </div>
                )}
              </div>
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={() => handleRemove(server.id)}
                disabled={saving}
                className="text-muted-foreground hover:text-destructive"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          ))}
        </div>
      )}

      <Button onClick={handleSave} disabled={!dirty || saving} size="sm">
        {saving ? (
          <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
        ) : null}
        Save
      </Button>

      {/* Add Server Picker Dialog */}
      {showPicker && (
        <Dialog open onOpenChange={(v) => { if (!v) setShowPicker(false); }}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle className="text-sm">Add MCP Server</DialogTitle>
              <DialogDescription className="text-xs">
                Select an MCP server from the workspace registry to bind to this agent.
              </DialogDescription>
            </DialogHeader>
            <div className="max-h-64 overflow-y-auto space-y-1">
              {unboundServers.map((server) => (
                <button
                  key={server.id}
                  onClick={() => handleAdd(server.id)}
                  disabled={saving}
                  className="flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent/50"
                >
                  <Server className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <div className="min-w-0 flex-1">
                    <div className="font-medium">{server.name}</div>
                    {server.description && (
                      <div className="text-xs text-muted-foreground truncate">
                        {server.description}
                      </div>
                    )}
                  </div>
                </button>
              ))}
              {unboundServers.length === 0 && (
                <p className="py-6 text-center text-xs text-muted-foreground">
                  All workspace MCP servers are already bound.
                </p>
              )}
            </div>
            <DialogFooter>
              <Button variant="ghost" onClick={() => setShowPicker(false)}>
                Cancel
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
