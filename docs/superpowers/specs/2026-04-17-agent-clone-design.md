# Agent Clone

## Summary

Allow users to clone an existing agent's configuration into a new agent via the agent detail page. Opens the existing CreateAgentDialog pre-filled with the source agent's data (except environment variables), letting the user review and modify before creating.

## Scope

- Frontend only — reuses existing `POST /api/agents` endpoint
- Two files modified: `CreateAgentDialog`, `AgentDetail`
- No backend, store, or API changes

## Design

### 1. CreateAgentDialog — `initialData` prop

Add an optional `initialData` prop of type `Omit<CreateAgentRequest, "custom_env">`:

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
  initialData?: Omit<CreateAgentRequest, "custom_env">;
  runtimes: RuntimeDevice[];
  runtimesLoading?: boolean;
  members: MemberWithUser[];
  currentUserId: string | null;
  onClose: () => void;
  onCreate: (data: CreateAgentRequest) => Promise<void>;
})
```

When `initialData` is provided:

| Field | Behavior |
|---|---|
| `name` | Pre-fill with `initialData.name + " (copy)"` |
| `description` | Pre-fill from `initialData.description` |
| `instructions` | Pre-fill from `initialData.instructions` |
| `visibility` | Pre-fill from `initialData.visibility` |
| `runtime_id` | Pre-fill from `initialData.runtime_id` if runtime exists in list |
| `runtime_config` | Pre-fill from `initialData.runtime_config` |
| `custom_args` | Pre-fill from `initialData.custom_args` |
| `max_concurrent_tasks` | Pre-fill from `initialData.max_concurrent_tasks` |
| `custom_env` | Always empty (not in `initialData` type) |
| `avatar_url` | Pre-fill from `initialData.avatar_url` |

UI differences when `initialData` is set:
- Dialog title: "Clone Agent" instead of "Create Agent"
- Submit button: "Clone" instead of "Create"

### 2. AgentDetail — Clone menu item

Add a "Clone Agent" item to the existing `DropdownMenu` in the detail page header (before the Archive item).

On click:
1. Build `initialData` from the current `agent` prop, omitting `custom_env`
2. Open `CreateAgentDialog` with this `initialData`

State additions to `AgentDetail`:
- `showCloneDialog: boolean` — controls dialog visibility

Prop additions to `AgentDetail`:
- `runtimesLoading?: boolean` — forwarded to `CreateAgentDialog`
- `onCreate: (data: CreateAgentRequest) => Promise<void>` — forwarded to `CreateAgentDialog`

### Data Flow

```
AgentDetail
  → User clicks "Clone Agent" in dropdown
  → showCloneDialog = true
  → CreateAgentDialog opens with initialData (agent config, no env)
  → User reviews/modifies fields
  → User clicks "Clone"
  → onCreate(CreateAgentRequest) called
  → POST /api/agents (existing endpoint)
  → New agent created
```

### Files Changed

| File | Change |
|---|---|
| `packages/views/agents/components/create-agent-dialog.tsx` | Add `initialData` prop, pre-fill state, conditional title/button |
| `packages/views/agents/components/agent-detail.tsx` | Add clone menu item, dialog state, wire up `CreateAgentDialog` |

### Out of Scope

- Backend API changes
- New Zustand store or React Query mutations
- Cloning conversation history or sessions
- Bulk clone operations
