"use client";

import { useState, useMemo } from "react";
import { Server, Plus, Search } from "lucide-react";
import { Button } from "@multica/ui/components/ui/button";
import { Input } from "@multica/ui/components/ui/input";
import { cn } from "@multica/ui/lib/utils";
import type { MCPServer } from "@multica/core/types";

interface MCPListProps {
  servers: MCPServer[];
  selectedId: string | null;
  isCreating: boolean;
  onSelect: (id: string | null) => void;
  onCreate: () => void;
  onCancelCreate: () => void;
}

export function MCPList({
  servers,
  selectedId,
  isCreating,
  onSelect,
  onCreate,
  onCancelCreate,
}: MCPListProps) {
  const [search, setSearch] = useState("");

  const filtered = useMemo(() => {
    if (!search.trim()) return servers;
    const q = search.toLowerCase();
    return servers.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.description.toLowerCase().includes(q),
    );
  }, [servers, search]);

  return (
    <div className="overflow-y-auto h-full border-r">
      <div className="flex h-12 items-center justify-between border-b px-4">
        <h1 className="text-sm font-semibold">MCP Servers</h1>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={onCreate}
          >
            <Plus className="h-4 w-4 text-muted-foreground" />
          </Button>
        </div>
      </div>

      {/* Search */}
      <div className="p-2 border-b">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="Search servers..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-7 pl-8 text-xs"
          />
        </div>
      </div>

      {/* Creating placeholder */}
      {isCreating && (
        <button
          className={cn(
            "flex w-full items-center gap-3 px-4 py-3 text-left border-b bg-accent/50",
          )}
          onClick={onCancelCreate}
        >
          <Server className="h-4 w-4 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium truncate">New server</p>
            <p className="text-xs text-muted-foreground truncate">Unsaved</p>
          </div>
        </button>
      )}

      {filtered.length === 0 && !isCreating ? (
        <div className="flex flex-col items-center justify-center px-4 py-12">
          <Server className="h-8 w-8 text-muted-foreground/40" />
          <p className="mt-3 text-sm text-muted-foreground">
            {servers.length === 0 ? "No MCP servers yet" : "No matching servers"}
          </p>
          {servers.length === 0 && (
            <Button
              onClick={onCreate}
              size="xs"
              className="mt-3"
            >
              <Plus className="h-3 w-3" />
              Create Server
            </Button>
          )}
        </div>
      ) : (
        <div className="divide-y">
          {filtered.map((server) => (
            <button
              key={server.id}
              className={cn(
                "flex w-full items-center gap-3 px-4 py-3 text-left transition-colors",
                server.id === selectedId && !isCreating
                  ? "bg-accent"
                  : "hover:bg-accent/50",
              )}
              onClick={() => onSelect(server.id)}
            >
              <Server className="h-4 w-4 shrink-0 text-muted-foreground" />
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium truncate">{server.name}</p>
                {server.description && (
                  <p className="text-xs text-muted-foreground truncate">
                    {server.description}
                  </p>
                )}
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
