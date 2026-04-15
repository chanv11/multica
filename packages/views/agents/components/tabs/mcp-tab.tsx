"use client";

import { useState } from "react";
import { Save, Plus, Trash2, Loader2 } from "lucide-react";
import type { Agent, MCPServerConfig, MCPServersConfig } from "@multica/core/types";
import { Button } from "@multica/ui/components/ui/button";
import { Input } from "@multica/ui/components/ui/input";
import { Label } from "@multica/ui/components/ui/label";
import { Card, CardContent } from "@multica/ui/components/ui/card";
import { toast } from "sonner";

interface MCPServerEntry {
  id: number;
  name: string;
  command: string;
  args: string; // raw string, parsed to string[] on save
  envEntries: { id: number; key: string; value: string }[];
}

let nextId = 0;

function serversToEntries(servers: MCPServersConfig | undefined): MCPServerEntry[] {
  if (!servers) return [];
  return Object.entries(servers).map(([name, config]) => ({
    id: nextId++,
    name,
    command: config.command ?? "",
    args: (config.args ?? []).join(" "),
    envEntries: config.env
      ? Object.entries(config.env).map(([key, value]) => ({ id: nextId++, key, value }))
      : [],
  }));
}

function entriesToServers(entries: MCPServerEntry[]): MCPServersConfig {
  const servers: MCPServersConfig = {};
  for (const entry of entries) {
    const name = entry.name.trim();
    if (!name) continue;
    const config: MCPServerConfig = { command: entry.command.trim() };
    const args = entry.args.trim().split(/\s+/).filter(Boolean);
    if (args.length > 0) config.args = args;
    const env: Record<string, string> = {};
    for (const e of entry.envEntries) {
      const key = e.key.trim();
      if (key) env[key] = e.value;
    }
    if (Object.keys(env).length > 0) config.env = env;
    servers[name] = config;
  }
  return servers;
}

export function MCPTab({
  agent,
  onSave,
}: {
  agent: Agent;
  onSave: (updates: Partial<Agent>) => Promise<void>;
}) {
  const [entries, setEntries] = useState<MCPServerEntry[]>(() =>
    serversToEntries(agent.runtime_config?.mcp_servers),
  );
  const [saving, setSaving] = useState(false);

  const currentServers = entriesToServers(entries);
  const originalServers = agent.runtime_config?.mcp_servers ?? {};
  const dirty =
    JSON.stringify(currentServers) !== JSON.stringify(originalServers);

  
  const handleSave = async () => {
    // Validate: no duplicate names, no empty names for entries with command
    const names = entries.filter((e) => e.name.trim()).map((e) => e.name.trim());
    const uniqueNames = new Set(names);
    if (uniqueNames.size < names.length) {
      toast.error("Duplicate server names are not allowed");
      return;
    }
    for (const entry of entries) {
      if (entry.name.trim() && !entry.command.trim()) {
        toast.error(`Server "${entry.name}" must have a command`);
        return;
      }
    }

    setSaving(true);
    try {
      await onSave({
        runtime_config: {
          ...(agent.runtime_config as Record<string, unknown>),
          mcp_servers: currentServers,
        },
      });
      toast.success("MCP servers saved");
    } catch {
      toast.error("Failed to save MCP servers");
    } finally {
      setSaving(false);
    }
  };

  const addEntry = () => {
    setEntries([
      ...entries,
      { id: nextId++, name: "", command: "", args: "", envEntries: [] },
    ]);
  };

  const removeEntry = (index: number) => {
    setEntries(entries.filter((_, i) => i !== index));
  };

  const updateEntry = (
    index: number,
    field: "name" | "command" | "args",
    value: string,
  ) => {
    setEntries(
      entries.map((e, i) => (i === index ? { ...e, [field]: value } : e)),
    );
  };

  const addEnvEntry = (serverIndex: number) => {
    setEntries(
      entries.map((e, i) =>
        i === serverIndex
          ? { ...e, envEntries: [...e.envEntries, { id: nextId++, key: "", value: "" }] }
          : e,
      ),
    );
  };

  const removeEnvEntry = (serverIndex: number, envIndex: number) => {
    setEntries(
      entries.map((e, i) =>
        i === serverIndex
          ? { ...e, envEntries: e.envEntries.filter((_, ei) => ei !== envIndex) }
          : e,
      ),
    );
  };

  const updateEnvEntry = (
    serverIndex: number,
    envIndex: number,
    field: "key" | "value",
    val: string,
  ) => {
    setEntries(
      entries.map((e, i) =>
        i === serverIndex
          ? {
              ...e,
              envEntries: e.envEntries.map((ev, ei) =>
                ei === envIndex ? { ...ev, [field]: val } : ev,
              ),
            }
          : e,
      ),
    );
  };

  return (
    <div className="max-w-2xl space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <Label className="text-xs text-muted-foreground">MCP Servers</Label>
          <p className="text-xs text-muted-foreground mt-0.5">
            Configure MCP servers available to this agent during task execution.
          </p>
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={addEntry}
          className="h-7 gap-1 text-xs"
        >
          <Plus className="h-3 w-3" />
          Add Server
        </Button>
      </div>

      {entries.length > 0 && (
        <div className="space-y-3">
          {entries.map((entry, index) => (
            <Card key={entry.id}>
              <CardContent className="space-y-3 pt-4">
                <div className="flex items-center gap-2">
                  <Input
                    value={entry.name}
                    onChange={(e) => updateEntry(index, "name", e.target.value)}
                    placeholder="Server name (e.g. github, filesystem)"
                    className="flex-1 text-xs"
                  />
                  <button
                    type="button"
                    onClick={() => removeEntry(index)}
                    className="shrink-0 text-muted-foreground hover:text-destructive"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
                <Input
                  value={entry.command}
                  onChange={(e) => updateEntry(index, "command", e.target.value)}
                  placeholder="Command (e.g. npx, node, python)"
                  className="text-xs font-mono"
                />
                <Input
                  value={entry.args}
                  onChange={(e) => updateEntry(index, "args", e.target.value)}
                  placeholder="Arguments (space-separated)"
                  className="text-xs font-mono"
                />
                {entry.envEntries.length > 0 && (
                  <div className="space-y-1.5 pl-2 border-l-2 border-muted">
                    {entry.envEntries.map((env, envIndex) => (
                      <div key={env.id} className="flex items-center gap-2">
                        <Input
                          value={env.key}
                          onChange={(e) =>
                            updateEnvEntry(index, envIndex, "key", e.target.value)
                          }
                          placeholder="ENV_KEY"
                          className="w-[40%] font-mono text-xs"
                        />
                        <Input
                          value={env.value}
                          onChange={(e) =>
                            updateEnvEntry(
                              index,
                              envIndex,
                              "value",
                              e.target.value,
                            )
                          }
                          placeholder="value"
                          className="flex-1 font-mono text-xs"
                        />
                        <button
                          type="button"
                          onClick={() => removeEnvEntry(index, envIndex)}
                          className="shrink-0 text-muted-foreground hover:text-destructive"
                        >
                          <Trash2 className="h-3 w-3" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => addEnvEntry(index)}
                  className="h-6 gap-1 text-xs text-muted-foreground"
                >
                  <Plus className="h-3 w-3" />
                  Add env var
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Button onClick={handleSave} disabled={!dirty || saving} size="sm">
        {saving ? (
          <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
        ) : (
          <Save className="h-3.5 w-3.5 mr-1.5" />
        )}
        Save
      </Button>
    </div>
  );
}
