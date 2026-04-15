"use client";

import { useState, useEffect, useCallback } from "react";
import { Save, Trash2, AlertCircle } from "lucide-react";
import { Button } from "@multica/ui/components/ui/button";
import { Input } from "@multica/ui/components/ui/input";
import { Textarea } from "@multica/ui/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@multica/ui/components/ui/dialog";
import { toast } from "sonner";
import {
  useCreateMCPServer,
  useUpdateMCPServer,
  useDeleteMCPServer,
} from "@multica/core/mcp/mutations";
import type { MCPServer } from "@multica/core/types";

interface MCPDetailProps {
  wsId: string;
  server: MCPServer | null;
  onCreated: (id: string) => void;
  onCancelled: () => void;
  onDeleted: () => void;
}

export function MCPDetail({
  wsId,
  server,
  onCreated,
  onCancelled,
  onDeleted,
}: MCPDetailProps) {
  const isEditing = !!server;

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [configJson, setConfigJson] = useState("{}");
  const [confirmDelete, setConfirmDelete] = useState(false);

  // Populate form when server changes
  useEffect(() => {
    if (server) {
      setName(server.name);
      setDescription(server.description);
      setConfigJson(
        server.config
          ? JSON.stringify(server.config, null, 2)
          : "{}",
      );
    } else {
      setName("");
      setDescription("");
      setConfigJson("{}");
    }
  }, [server]);

  const createMut = useCreateMCPServer(wsId);
  const updateMut = useUpdateMCPServer(wsId);
  const deleteMut = useDeleteMCPServer(wsId);

  const handleSave = useCallback(async () => {
    if (!name.trim()) {
      toast.error("Name is required");
      return;
    }

    let parsedConfig: Record<string, unknown>;
    try {
      parsedConfig = JSON.parse(configJson);
    } catch {
      toast.error("Config must be valid JSON");
      return;
    }

    try {
      if (isEditing && server) {
        await updateMut.mutateAsync({
          id: server.id,
          name: name.trim(),
          description: description.trim(),
          config: parsedConfig,
        });
        toast.success("Server updated");
      } else {
        const created = await createMut.mutateAsync({
          name: name.trim(),
          description: description.trim(),
          config: parsedConfig,
        });
        toast.success("Server created");
        onCreated(created.id);
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to save server");
    }
  }, [
    name,
    description,
    configJson,
    isEditing,
    server,
    createMut,
    updateMut,
    onCreated,
  ]);

  const handleDelete = useCallback(async () => {
    if (!server) return;
    try {
      await deleteMut.mutateAsync(server.id);
      toast.success("Server deleted");
      setConfirmDelete(false);
      onDeleted();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to delete server");
    }
  }, [server, deleteMut, onDeleted]);

  const isSaving = createMut.isPending || updateMut.isPending;

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex h-12 shrink-0 items-center gap-3 border-b px-4">
        <h2 className="text-sm font-semibold">
          {isEditing ? server.name : "New MCP Server"}
        </h2>
        <div className="ml-auto flex items-center gap-2">
          {isEditing && (
            <Button
              variant="ghost"
              size="sm"
              className="text-destructive hover:text-destructive"
              onClick={() => setConfirmDelete(true)}
            >
              <Trash2 className="h-3.5 w-3.5" />
              Delete
            </Button>
          )}
          {!isEditing && (
            <Button variant="ghost" size="sm" onClick={onCancelled}>
              Cancel
            </Button>
          )}
          <Button
            size="sm"
            onClick={handleSave}
            disabled={isSaving || !name.trim()}
          >
            <Save className="h-3.5 w-3.5" />
            {isSaving ? "Saving..." : isEditing ? "Save" : "Create"}
          </Button>
        </div>
      </div>

      {/* Form */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-2xl space-y-6">
          {/* Name */}
          <div className="space-y-2">
            <label className="text-sm font-medium" htmlFor="mcp-name">
              Name
            </label>
            <Input
              id="mcp-name"
              placeholder="e.g. GitHub MCP"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <label className="text-sm font-medium" htmlFor="mcp-description">
              Description
            </label>
            <Textarea
              id="mcp-description"
              placeholder="What does this MCP server do?"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>

          {/* Config JSON */}
          <div className="space-y-2">
            <label className="text-sm font-medium" htmlFor="mcp-config">
              Config (JSON)
            </label>
            <Textarea
              id="mcp-config"
              value={configJson}
              onChange={(e) => setConfigJson(e.target.value)}
              rows={12}
              className="font-mono text-xs"
              placeholder='{"command": "npx", "args": ["-y", "@modelcontextprotocol/server-example"]}'
            />
            <p className="text-xs text-muted-foreground">
              Define the MCP server configuration as JSON. Use{" "}
              <code className="rounded bg-muted px-1 py-0.5 text-xs">
                {"${VAR}"}
              </code>{" "}
              to reference environment variables from the agent Environment tab
              -- these are resolved at runtime.
            </p>
          </div>
        </div>
      </div>

      {/* Delete Confirmation */}
      {confirmDelete && (
        <Dialog open onOpenChange={(v) => { if (!v) setConfirmDelete(false); }}>
          <DialogContent className="max-w-sm" showCloseButton={false}>
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-destructive/10">
                <AlertCircle className="h-5 w-5 text-destructive" />
              </div>
              <DialogHeader className="flex-1 gap-1">
                <DialogTitle className="text-sm font-semibold">Delete MCP server?</DialogTitle>
                <DialogDescription className="text-xs">
                  &quot;{server?.name}&quot; will be permanently deleted. Agents using this server will lose access to its tools.
                </DialogDescription>
              </DialogHeader>
            </div>
            <DialogFooter>
              <Button variant="ghost" onClick={() => setConfirmDelete(false)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteMut.isPending}
              >
                {deleteMut.isPending ? "Deleting..." : "Delete"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
