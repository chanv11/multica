import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { Agent, MCPServer, AgentMCPBinding } from "@multica/core/types";
import { WorkspaceIdProvider } from "@multica/core/hooks";

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

const mockListMCPServers = vi.hoisted(() => vi.fn().mockResolvedValue([]));
const mockGetAgentMCPBindings = vi.hoisted(() => vi.fn().mockResolvedValue([]));
const mockReplaceAgentMCPBindings = vi.hoisted(() => vi.fn().mockResolvedValue([]));

vi.mock("@multica/core/api", () => ({
  api: {
    listMCPServers: (...args: any[]) => mockListMCPServers(...args),
    getMCPServer: vi.fn(),
    createMCPServer: vi.fn(),
    updateMCPServer: vi.fn(),
    deleteMCPServer: vi.fn(),
    getAgentMCPBindings: (...args: any[]) => mockGetAgentMCPBindings(...args),
    replaceAgentMCPBindings: (...args: any[]) => mockReplaceAgentMCPBindings(...args),
  },
  getApi: () => ({
    listMCPServers: (...args: any[]) => mockListMCPServers(...args),
    getAgentMCPBindings: (...args: any[]) => mockGetAgentMCPBindings(...args),
    replaceAgentMCPBindings: (...args: any[]) => mockReplaceAgentMCPBindings(...args),
  }),
  setApiInstance: vi.fn(),
}));

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

// Mock @multica/core/auth
const mockAuthUser = { id: "user-1", email: "test@test.com", name: "Test User" };
vi.mock("@multica/core/auth", () => ({
  useAuthStore: Object.assign(
    (selector?: any) => {
      const state = { user: mockAuthUser, isAuthenticated: true };
      return selector ? selector(state) : state;
    },
    { getState: () => ({ user: mockAuthUser, isAuthenticated: true }) },
  ),
  registerAuthStore: vi.fn(),
  createAuthStore: vi.fn(),
}));

// ---------------------------------------------------------------------------
// Test data
// ---------------------------------------------------------------------------

const mockAgent: Agent = {
  id: "agent-1",
  workspace_id: "ws-1",
  runtime_id: "runtime-1",
  name: "Test Agent",
  description: "A test agent",
  instructions: "",
  avatar_url: null,
  runtime_mode: "cloud",
  runtime_config: {},
  custom_env: {},
  visibility: "workspace",
  status: "idle",
  max_concurrent_tasks: 1,
  owner_id: "user-1",
  skills: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
  archived_at: null,
  archived_by: null,
};

const mockServers: MCPServer[] = [
  {
    id: "mcp-1",
    workspace_id: "ws-1",
    name: "GitHub MCP",
    description: "GitHub integration server",
    config: { command: "npx", args: ["-y", "@modelcontextprotocol/server-github"] },
    created_by: "user-1",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: "mcp-2",
    workspace_id: "ws-1",
    name: "Filesystem MCP",
    description: "Local filesystem access",
    config: { command: "npx", args: ["-y", "@modelcontextprotocol/server-filesystem"] },
    created_by: "user-1",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: "mcp-3",
    workspace_id: "ws-1",
    name: "Slack MCP",
    description: "Slack messaging integration",
    config: { command: "npx", args: ["-y", "@modelcontextprotocol/server-slack"] },
    created_by: "user-1",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const mockBindings: AgentMCPBinding[] = [
  {
    agent_id: "agent-1",
    mcp_server_id: "mcp-1",
    enabled: true,
    sort_order: 0,
    created_at: "2026-01-01T00:00:00Z",
  },
];

// ---------------------------------------------------------------------------
// Import component under test (after mocks)
// ---------------------------------------------------------------------------

import { MCPTab } from "./mcp-tab";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function renderWithQuery(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  });
  return render(
    <QueryClientProvider client={qc}>
      <WorkspaceIdProvider wsId="ws-1">{ui}</WorkspaceIdProvider>
    </QueryClientProvider>,
  );
}

function getAddServerButton(): HTMLElement {
  const buttons = screen.getAllByRole("button").filter(
    (btn) => btn.textContent?.includes("Add Server"),
  );
  expect(buttons.length).toBeGreaterThan(0);
  return buttons[0]!;
}

function getRemoveButtons(): HTMLElement[] {
  return screen.getAllByRole("button").filter(
    (btn) => btn.querySelector("svg.lucide-trash-2") !== null,
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("MCPTab", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListMCPServers.mockResolvedValue(mockServers);
    mockGetAgentMCPBindings.mockResolvedValue(mockBindings);
    mockReplaceAgentMCPBindings.mockResolvedValue([]);
  });

  it("renders existing bindings", async () => {
    renderWithQuery(<MCPTab agent={mockAgent} />);

    await screen.findByText("GitHub MCP");
    expect(screen.getByText("GitHub integration server")).toBeInTheDocument();
  });

  it("shows empty state when no bindings exist", async () => {
    mockGetAgentMCPBindings.mockResolvedValue([]);

    renderWithQuery(<MCPTab agent={mockAgent} />);

    await screen.findByText("No MCP servers bound");
  });

  it("shows explanatory text for ${VAR} resolution", async () => {
    renderWithQuery(<MCPTab agent={mockAgent} />);

    await screen.findByText(/MCP servers provide tools and capabilities/i);
    const infoText = screen.getByText(/\${VAR}/);
    expect(infoText).toBeInTheDocument();
  });

  it("adds an available MCP server from the picker", async () => {
    const user = userEvent.setup();

    renderWithQuery(<MCPTab agent={mockAgent} />);

    // Wait for data to load
    await screen.findByText("GitHub MCP");

    // Click "Add Server" button
    await user.click(getAddServerButton());

    // Picker dialog should appear
    await screen.findByText("Select an MCP server from the workspace registry to bind to this agent.");

    // Unbound servers should appear in the picker
    expect(screen.getByText("Filesystem MCP")).toBeInTheDocument();
    expect(screen.getByText("Slack MCP")).toBeInTheDocument();

    // Click on Filesystem to add it
    await user.click(screen.getByText("Filesystem MCP"));

    // Dialog should close and server should now appear in the bound list
    await waitFor(() => {
      expect(screen.getByText("Local filesystem access")).toBeInTheDocument();
    });
  });

  it("removes a bound MCP server", async () => {
    const user = userEvent.setup();

    renderWithQuery(<MCPTab agent={mockAgent} />);

    // Wait for data to load
    await screen.findByText("GitHub MCP");

    // Find the remove button for the bound server (Trash2 icon)
    const removeButtons = getRemoveButtons();
    expect(removeButtons.length).toBeGreaterThan(0);

    await user.click(removeButtons[0]!);

    // Server should be removed from the bound list — shows empty state
    await screen.findByText("No MCP servers bound");
  });

  it("save button calls replaceAgentMCPBindings with full list", async () => {
    const user = userEvent.setup();

    renderWithQuery(<MCPTab agent={mockAgent} />);

    // Wait for data to load
    await screen.findByText("GitHub MCP");

    // Save button should be disabled initially (no changes)
    const saveButton = screen.getByRole("button", { name: /Save/ });
    expect(saveButton).toBeDisabled();

    // Add another server first
    await user.click(getAddServerButton());

    await screen.findByText("Slack MCP");
    await user.click(screen.getByText("Slack MCP"));

    // Now save should be enabled
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Save/ })).toBeEnabled();
    });

    await user.click(screen.getByRole("button", { name: /Save/ }));

    await waitFor(() => {
      expect(mockReplaceAgentMCPBindings).toHaveBeenCalledWith(
        "agent-1",
        expect.objectContaining({
          mcp_server_ids: expect.arrayContaining(["mcp-1", "mcp-3"]),
        }),
      );
    });
  });

  it("shows toast on successful save", async () => {
    const { toast } = await import("sonner");
    const user = userEvent.setup();

    renderWithQuery(<MCPTab agent={mockAgent} />);

    await screen.findByText("GitHub MCP");

    // Add a server to make dirty
    await user.click(getAddServerButton());

    await screen.findByText("Slack MCP");
    await user.click(screen.getByText("Slack MCP"));

    await user.click(screen.getByRole("button", { name: /Save/ }));

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("MCP bindings saved");
    });
  });

  it("shows toast on failed save", async () => {
    const { toast } = await import("sonner");
    const user = userEvent.setup();
    mockReplaceAgentMCPBindings.mockRejectedValue(new Error("Network error"));

    renderWithQuery(<MCPTab agent={mockAgent} />);

    await screen.findByText("GitHub MCP");

    // Add a server to make dirty
    await user.click(getAddServerButton());

    await screen.findByText("Slack MCP");
    await user.click(screen.getByText("Slack MCP"));

    await user.click(screen.getByRole("button", { name: /Save/ }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("Failed to save MCP bindings");
    });
  });

  it("shows all servers in picker when no bindings exist", async () => {
    const user = userEvent.setup();
    mockGetAgentMCPBindings.mockResolvedValue([]);

    renderWithQuery(<MCPTab agent={mockAgent} />);

    await screen.findByText("No MCP servers bound");

    // There's an Add Server in the empty state
    await user.click(getAddServerButton());

    // All three servers should be available
    await screen.findByText("GitHub MCP");
    expect(screen.getByText("Filesystem MCP")).toBeInTheDocument();
    expect(screen.getByText("Slack MCP")).toBeInTheDocument();
  });

  it("does not show already-bound servers in the picker", async () => {
    const user = userEvent.setup();

    renderWithQuery(<MCPTab agent={mockAgent} />);

    await screen.findByText("GitHub MCP");

    // Open picker
    await user.click(getAddServerButton());

    await screen.findByText("Select an MCP server from the workspace registry to bind to this agent.");

    // GitHub MCP is already bound, so it should NOT appear in the picker
    const pickerButtons = screen.getAllByRole("button").filter(
      (btn) => btn.textContent?.includes("GitHub MCP"),
    );
    expect(pickerButtons.length).toBe(0);
  });
});
