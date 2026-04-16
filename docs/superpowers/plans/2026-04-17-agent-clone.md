# Agent Clone Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a "Clone Agent" action to the agent detail page that opens the existing CreateAgentDialog pre-filled with the source agent's configuration (minus environment variables).

**Architecture:** Extend `CreateAgentDialog` with an optional `initialData` prop to pre-fill form state. Add a "Clone Agent" menu item to `AgentDetail`'s dropdown that opens this dialog. The dialog calls the existing `onCreate` callback — no backend changes needed.

**Tech Stack:** React, TypeScript, Zustand (no changes), TanStack Query (no changes), shadcn UI components

---

## File Structure

| File | Action | Responsibility |
|---|---|---|
| `packages/views/agents/components/create-agent-dialog.tsx` | Modify | Add `initialData` prop, pre-fill state, conditional title/button |
| `packages/views/agents/components/agent-detail.tsx` | Modify | Add "Clone Agent" menu item, manage clone dialog state |
| `packages/views/agents/components/agents-page.tsx` | Modify | Wire `onCreate` + `runtimesLoading` through to `AgentDetail` |

---

### Task 1: Add `initialData` prop to `CreateAgentDialog`

**Files:**
- Modify: `packages/views/agents/components/create-agent-dialog.tsx`

- [ ] **Step 1: Add `initialData` type and prop**

Add the type alias and update the component signature. In `create-agent-dialog.tsx`, add after the existing `RuntimeFilter` type (line 31):

```ts
export type CloneAgentInitialData = Omit<CreateAgentRequest, "custom_env">;
```

Update the component props to include `initialData`:

```ts
export function CreateAgentDialog({
  initialData,
  runtimes,
  runtimesLoading,
  members,
  currentUserId,
  onClose,
  onCreate,
}: {
  initialData?: CloneAgentInitialData;
  runtimes: RuntimeDevice[];
  runtimesLoading?: boolean;
  members: MemberWithUser[];
  currentUserId: string | null;
  onClose: () => void;
  onCreate: (data: CreateAgentRequest) => Promise<void>;
}) {
```

- [ ] **Step 2: Pre-fill state from `initialData`**

Replace the state initializations (lines 48-53) to use `initialData` when provided:

```ts
  const [name, setName] = useState(() => initialData?.name ? `${initialData.name} (copy)` : "");
  const [description, setDescription] = useState(() => initialData?.description ?? "");
  const [visibility, setVisibility] = useState<AgentVisibility>(() => initialData?.visibility ?? "private");
  const [creating, setCreating] = useState(false);
  const [runtimeOpen, setRuntimeOpen] = useState(false);
  const [runtimeFilter, setRuntimeFilter] = useState<RuntimeFilter>("mine");
```

- [ ] **Step 3: Pre-fill runtime selection**

Replace the `selectedRuntimeId` initialization (line 73) to use `initialData.runtime_id`:

```ts
  const [selectedRuntimeId, setSelectedRuntimeId] = useState(
    () => initialData?.runtime_id ?? filteredRuntimes[0]?.id ?? ""
  );
```

- [ ] **Step 4: Update dialog title and button text**

Change the `<DialogTitle>` (line 105):

```tsx
          <DialogTitle>{initialData ? "Clone Agent" : "Create Agent"}</DialogTitle>
```

Change the `<DialogDescription>` (line 106-108):

```tsx
          <DialogDescription>
            {initialData ? "Create a new agent based on an existing configuration." : "Create a new AI agent for your workspace."}
          </DialogDescription>
```

Change the submit button text (line 287):

```tsx
            {creating ? (initialData ? "Cloning..." : "Creating...") : (initialData ? "Clone" : "Create")}
```

- [ ] **Step 5: Commit**

```bash
git add packages/views/agents/components/create-agent-dialog.tsx
git commit -m "feat(agents): add initialData prop to CreateAgentDialog for clone support"
```

---

### Task 2: Add clone action to `AgentDetail`

**Files:**
- Modify: `packages/views/agents/components/agent-detail.tsx`

- [ ] **Step 1: Add imports**

Add `Copy` to the lucide-react import (line 4-16):

```ts
import {
  Cloud,
  Monitor,
  FileText,
  BookOpenText,
  ListTodo,
  Trash2,
  AlertCircle,
  MoreHorizontal,
  Settings,
  KeyRound,
  Terminal,
  Copy,
} from "lucide-react";
```

Add `CreateAgentRequest` to the type imports (line 17):

```ts
import type { Agent, RuntimeDevice, MemberWithUser, CreateAgentRequest } from "@multica/core/types";
```

Add the `CreateAgentDialog` import:

```ts
import { CreateAgentDialog } from "./create-agent-dialog";
```

- [ ] **Step 2: Update props**

Update the component props to include `runtimesLoading` and `onCreate` (lines 57-73):

```ts
export function AgentDetail({
  agent,
  runtimes,
  runtimesLoading,
  members,
  currentUserId,
  onUpdate,
  onArchive,
  onRestore,
  onCreate,
}: {
  agent: Agent;
  runtimes: RuntimeDevice[];
  runtimesLoading?: boolean;
  members: MemberWithUser[];
  currentUserId: string | null;
  onUpdate: (id: string, data: Partial<Agent>) => Promise<void>;
  onArchive: (id: string) => Promise<void>;
  onRestore: (id: string) => Promise<void>;
  onCreate: (data: CreateAgentRequest) => Promise<void>;
}) {
```

- [ ] **Step 3: Add clone dialog state**

After the existing state declarations (after line 77, before the `isArchived` line), add:

```ts
  const [showCloneDialog, setShowCloneDialog] = useState(false);
```

- [ ] **Step 4: Add "Clone Agent" menu item**

In the `DropdownMenuContent` (between line 128-136), add the Clone menu item **before** the Archive item:

```tsx
            <DropdownMenuContent align="end" className="w-auto">
              <DropdownMenuItem
                onClick={() => setShowCloneDialog(true)}
              >
                <Copy className="h-3.5 w-3.5" />
                Clone Agent
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-destructive"
                onClick={() => setConfirmArchive(true)}
              >
                <Trash2 className="h-3.5 w-3.5" />
                Archive Agent
              </DropdownMenuItem>
            </DropdownMenuContent>
```

- [ ] **Step 5: Add the clone dialog**

After the archive confirmation dialog (after line 226, before the closing `</div>`), add:

```tsx
      {showCloneDialog && (
        <CreateAgentDialog
          initialData={{
            name: agent.name,
            description: agent.description,
            instructions: agent.instructions,
            avatar_url: agent.avatar_url ?? undefined,
            runtime_id: agent.runtime_id,
            runtime_config: agent.runtime_config,
            custom_args: agent.custom_args,
            visibility: agent.visibility,
            max_concurrent_tasks: agent.max_concurrent_tasks,
          }}
          runtimes={runtimes}
          runtimesLoading={runtimesLoading}
          members={members}
          currentUserId={currentUserId}
          onClose={() => setShowCloneDialog(false)}
          onCreate={async (data) => {
            await onCreate(data);
            setShowCloneDialog(false);
          }}
        />
      )}
```

- [ ] **Step 6: Commit**

```bash
git add packages/views/agents/components/agent-detail.tsx
git commit -m "feat(agents): add clone agent action to agent detail dropdown"
```

---

### Task 3: Wire new props through `AgentsPage`

**Files:**
- Modify: `packages/views/agents/components/agents-page.tsx`

- [ ] **Step 1: Pass `runtimesLoading` and `onCreate` to `AgentDetail`**

Update the `<AgentDetail>` usage (around line 203-211) to pass the two new props:

```tsx
          <AgentDetail
            key={selected.id}
            agent={selected}
            runtimes={runtimes}
            runtimesLoading={runtimesLoading}
            members={members}
            currentUserId={currentUser?.id ?? null}
            onUpdate={handleUpdate}
            onArchive={handleArchive}
            onRestore={handleRestore}
            onCreate={handleCreate}
          />
```

- [ ] **Step 2: Commit**

```bash
git add packages/views/agents/components/agents-page.tsx
git commit -m "feat(agents): wire runtimesLoading and onCreate through to AgentDetail"
```

---

### Task 4: Verify

- [ ] **Step 1: Run typecheck**

```bash
cd /root/Document/multica && pnpm typecheck
```

Expected: No errors.

- [ ] **Step 2: Run unit tests**

```bash
cd /root/Document/multica && pnpm test
```

Expected: All tests pass.
