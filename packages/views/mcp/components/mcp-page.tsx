"use client";

import { useState, useEffect, useMemo } from "react";
import { useDefaultLayout } from "react-resizable-panels";
import { Server, Plus } from "lucide-react";
import {
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle,
} from "@multica/ui/components/ui/resizable";
import { Button } from "@multica/ui/components/ui/button";
import { useQuery } from "@tanstack/react-query";
import { mcpListOptions } from "@multica/core/mcp/queries";
import { useWorkspaceId } from "@multica/core/hooks";
import { MCPList } from "./mcp-list";
import { MCPDetail } from "./mcp-detail";

export function MCPPage() {
  const wsId = useWorkspaceId();
  const { data: servers = [] } = useQuery(mcpListOptions(wsId));
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const { defaultLayout, onLayoutChanged } = useDefaultLayout({
    id: "multica_mcp_layout",
  });

  // Select first server on initial load
  useEffect(() => {
    if (servers.length > 0 && !servers.some((s) => s.id === selectedId)) {
      setSelectedId(servers[0]!.id);
    }
  }, [servers, selectedId]);

  const handleCreate = () => {
    setIsCreating(true);
    setSelectedId(null);
  };

  const handleCancelCreate = () => {
    setIsCreating(false);
    if (servers.length > 0) {
      setSelectedId(servers[0]!.id);
    }
  };

  const handleCreated = (id: string) => {
    setIsCreating(false);
    setSelectedId(id);
  };

  const handleDeleted = () => {
    setSelectedId(null);
    if (servers.length > 0) {
      setSelectedId(servers[0]!.id);
    }
  };

  const selected = useMemo(
    () => servers.find((s) => s.id === selectedId) ?? null,
    [servers, selectedId],
  );

  return (
    <ResizablePanelGroup
      orientation="horizontal"
      className="flex-1 min-h-0"
      defaultLayout={defaultLayout}
      onLayoutChanged={onLayoutChanged}
    >
      <ResizablePanel id="list" defaultSize={280} minSize={240} maxSize={400} groupResizeBehavior="preserve-pixel-size">
        {/* Left column -- MCP list */}
        <MCPList
          servers={servers}
          selectedId={selectedId}
          isCreating={isCreating}
          onSelect={setSelectedId}
          onCreate={handleCreate}
          onCancelCreate={handleCancelCreate}
        />
      </ResizablePanel>

      <ResizableHandle />

      <ResizablePanel id="detail" minSize="50%">
        {/* Right column -- detail editor */}
        {isCreating ? (
          <MCPDetail
            wsId={wsId}
            server={null}
            onCreated={handleCreated}
            onCancelled={handleCancelCreate}
            onDeleted={handleDeleted}
          />
        ) : selected ? (
          <MCPDetail
            key={selected.id}
            wsId={wsId}
            server={selected}
            onCreated={handleCreated}
            onCancelled={handleCancelCreate}
            onDeleted={handleDeleted}
          />
        ) : (
          <div className="flex h-full flex-col items-center justify-center text-muted-foreground">
            <Server className="h-10 w-10 text-muted-foreground/30" />
            <p className="mt-3 text-sm">Select an MCP server to view details</p>
            <Button
              onClick={handleCreate}
              size="xs"
              className="mt-3"
            >
              <Plus className="h-3 w-3" />
              Create Server
            </Button>
          </div>
        )}
      </ResizablePanel>
    </ResizablePanelGroup>
  );
}
