import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { MCPServer } from "@multica/core/types";
import { WorkspaceIdProvider } from "@multica/core/hooks";

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

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

// Mock @multica/core/workspace
vi.mock("@multica/core/workspace", () => ({
  useWorkspaceStore: Object.assign(
    (selector?: any) => {
      const state = {
        workspace: { id: "ws-1", name: "Test WS", slug: "test" },
        agents: [],
        members: [],
      };
      return selector ? selector(state) : state;
    },
    {
      getState: () => ({
        workspace: { id: "ws-1", name: "Test WS", slug: "test" },
        agents: [],
        members: [],
      }),
    },
  ),
  registerWorkspaceStore: vi.fn(),
}));

// Mock @multica/core/api
const mockListMCPServers = vi.hoisted(() => vi.fn().mockResolvedValue([]));
const mockCreateMCPServer = vi.hoisted(() => vi.fn());
const mockUpdateMCPServer = vi.hoisted(() => vi.fn());
const mockDeleteMCPServer = vi.hoisted(() => vi.fn());

vi.mock("@multica/core/api", () => ({
  api: {
    listMCPServers: (...args: any[]) => mockListMCPServers(...args),
    getMCPServer: vi.fn(),
    createMCPServer: (...args: any[]) => mockCreateMCPServer(...args),
    updateMCPServer: (...args: any[]) => mockUpdateMCPServer(...args),
    deleteMCPServer: (...args: any[]) => mockDeleteMCPServer(...args),
  },
  getApi: () => ({
    listMCPServers: (...args: any[]) => mockListMCPServers(...args),
    createMCPServer: (...args: any[]) => mockCreateMCPServer(...args),
    updateMCPServer: (...args: any[]) => mockUpdateMCPServer(...args),
    deleteMCPServer: (...args: any[]) => mockDeleteMCPServer(...args),
  }),
  setApiInstance: vi.fn(),
}));

// Mock @multica/core/modals
vi.mock("@multica/core/modals", () => ({
  useModalStore: Object.assign(
    () => ({ open: vi.fn() }),
    { getState: () => ({ open: vi.fn() }) },
  ),
}));

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

// Mock react-resizable-panels
vi.mock("react-resizable-panels", () => ({
  useDefaultLayout: () => ({ defaultLayout: undefined, onLayoutChanged: vi.fn() }),
  Group: ({ children, ...props }: any) => <div data-slot="resizable-panel-group" {...props}>{children}</div>,
  Panel: ({ children, ...props }: any) => <div data-slot="resizable-panel" {...props}>{children}</div>,
  Separator: ({ children, ...props }: any) => <div data-slot="resizable-handle" {...props}>{children}</div>,
}));

// Mock @multica/views/navigation
vi.mock("../../navigation", () => ({
  AppLink: ({ children, href, ...props }: any) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
  useNavigation: () => ({ push: vi.fn(), pathname: "/mcp" }),
  NavigationProvider: ({ children }: { children: React.ReactNode }) => children,
}));

// Mock workspace avatar
vi.mock("../../workspace/workspace-avatar", () => ({
  WorkspaceAvatar: ({ name }: { name: string }) => <span data-testid="workspace-avatar">{name.charAt(0)}</span>,
}));

// ---------------------------------------------------------------------------
// Test data
// ---------------------------------------------------------------------------

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
];

// ---------------------------------------------------------------------------
// Import component under test (after mocks)
// ---------------------------------------------------------------------------

import { MCPPage } from "./mcp-page";

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

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("MCPPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListMCPServers.mockResolvedValue([]);
  });

  it("shows empty state when there are no MCP servers", async () => {
    renderWithQuery(<MCPPage />);

    await screen.findByText("No MCP servers yet");
    // "Create Server" appears in both the list empty state and the detail empty state
    const createButtons = screen.getAllByText("Create Server");
    expect(createButtons.length).toBeGreaterThanOrEqual(1);
  });

  it("renders list of MCP servers after data loads", async () => {
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    await screen.findByText("GitHub MCP");
    expect(screen.getByText("Filesystem MCP")).toBeInTheDocument();
  });

  it("shows server description in the list", async () => {
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    await screen.findByText("GitHub integration server");
    expect(screen.getByText("Local filesystem access")).toBeInTheDocument();
  });

  it("shows detail panel when a server is selected", async () => {
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    // First server is auto-selected
    await screen.findByText("GitHub MCP");
    // The detail header shows the selected server name
    const headings = screen.getAllByText("GitHub MCP");
    expect(headings.length).toBeGreaterThanOrEqual(2); // list + detail header
  });

  it("opens create form when Create button is clicked", async () => {
    const user = userEvent.setup();
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    // Click the + button in the list header
    const createButtons = screen.getAllByRole("button");
    const headerCreateBtn = createButtons.find(
      (btn) => btn.querySelector("svg.lucide-plus") !== null,
    );
    if (headerCreateBtn) {
      await user.click(headerCreateBtn);
    }

    // Should show the new server form
    await screen.findByText("New MCP Server");
    expect(screen.getByText("Create")).toBeInTheDocument();
  });

  it("shows Config JSON textarea in the detail panel", async () => {
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    await screen.findByText("Config (JSON)");
    expect(screen.getByLabelText("Config (JSON)")).toBeInTheDocument();
  });

  it("shows search input in the list panel", async () => {
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    await screen.findByPlaceholderText("Search servers...");
  });

  it("filters servers by search query", async () => {
    const user = userEvent.setup();
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    await screen.findByText("GitHub MCP");

    const searchInput = screen.getByPlaceholderText("Search servers...");
    await user.type(searchInput, "Filesystem");

    // GitHub is filtered out of the list but may still appear in the detail header (h2).
    // Only Filesystem should appear in the list buttons.
    const listButtons = screen.getAllByRole("button").filter(
      (btn) => btn.textContent?.includes("Filesystem MCP"),
    );
    expect(listButtons.length).toBeGreaterThanOrEqual(1);

    // GitHub should NOT appear as a list item (button), only possibly in the detail header
    const githubButtons = screen.getAllByRole("button").filter(
      (btn) => btn.textContent?.includes("GitHub MCP"),
    );
    expect(githubButtons.length).toBe(0);
  });

  it("shows delete button in detail when editing existing server", async () => {
    mockListMCPServers.mockResolvedValue(mockServers);

    renderWithQuery(<MCPPage />);

    await screen.findByText("Delete");
  });
});
