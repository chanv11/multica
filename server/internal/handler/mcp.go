package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// --- Response structs ---

type MCPServerResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      any    `json:"config"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func mcpServerToResponse(s db.WorkspaceMcpServer) MCPServerResponse {
	var config any
	if s.Config != nil {
		json.Unmarshal(s.Config, &config)
	}
	if config == nil {
		config = map[string]any{}
	}

	return MCPServerResponse{
		ID:          uuidToString(s.ID),
		WorkspaceID: uuidToString(s.WorkspaceID),
		Name:        s.Name,
		Description: s.Description,
		Config:      config,
		CreatedBy:   uuidToString(s.CreatedBy),
		CreatedAt:   timestampToString(s.CreatedAt),
		UpdatedAt:   timestampToString(s.UpdatedAt),
	}
}

// --- Request structs ---

type CreateMCPServerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      any    `json:"config"`
}

type UpdateMCPServerRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Config      any     `json:"config"`
}

// --- Validation helpers ---

// validateMCPConfig checks that config is a JSON object (not array, string,
// number, etc.) and that optional "args" and "env" fields have correct types.
func validateMCPConfig(config any) (bool, string) {
	if config == nil {
		return true, ""
	}

	// Config must be a map (JSON object), not array/string/number/bool.
	m, ok := config.(map[string]any)
	if !ok {
		return false, "config must be a JSON object"
	}

	// If "args" is present, it must be an array.
	if args, exists := m["args"]; exists {
		if _, isArray := args.([]any); !isArray {
			return false, "args must be an array"
		}
	}

	// If "env" is present, it must be an object.
	if env, exists := m["env"]; exists {
		if _, isMap := env.(map[string]any); !isMap {
			return false, "env must be an object"
		}
	}

	return true, ""
}

// --- Handlers ---

func (h *Handler) ListMCPServers(w http.ResponseWriter, r *http.Request) {
	workspaceID := resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}

	servers, err := h.Queries.ListMCPServers(r.Context(), parseUUID(workspaceID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list MCP servers")
		return
	}

	resp := make([]MCPServerResponse, len(servers))
	for i, s := range servers {
		resp[i] = mcpServerToResponse(s)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetMCPServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}

	server, err := h.Queries.GetMCPServerInWorkspace(r.Context(), db.GetMCPServerInWorkspaceParams{
		ID:          parseUUID(id),
		WorkspaceID: parseUUID(workspaceID),
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "MCP server not found")
		return
	}

	writeJSON(w, http.StatusOK, mcpServerToResponse(server))
}

func (h *Handler) CreateMCPServer(w http.ResponseWriter, r *http.Request) {
	workspaceID := resolveWorkspaceID(r)

	member, ok := h.workspaceMember(w, r, workspaceID)
	if !ok {
		return
	}

	var req CreateMCPServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if ok, msg := validateMCPConfig(req.Config); !ok {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	config, _ := json.Marshal(req.Config)
	if req.Config == nil {
		config = []byte("{}")
	}

	server, err := h.Queries.CreateMCPServer(r.Context(), db.CreateMCPServerParams{
		WorkspaceID: parseUUID(workspaceID),
		Name:        req.Name,
		Description: req.Description,
		Config:      config,
		CreatedBy:   member.ID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "an MCP server with this name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create MCP server")
		return
	}

	writeJSON(w, http.StatusCreated, mcpServerToResponse(server))
}

func (h *Handler) UpdateMCPServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}

	var req UpdateMCPServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Config != nil {
		if ok, msg := validateMCPConfig(req.Config); !ok {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
	}

	params := db.UpdateMCPServerParams{
		ID: parseUUID(id),
	}

	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	}
	if req.Config != nil {
		configBytes, _ := json.Marshal(req.Config)
		params.Config = configBytes
	}

	server, err := h.Queries.UpdateMCPServer(r.Context(), params)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "MCP server not found")
			return
		}
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "an MCP server with this name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update MCP server")
		return
	}

	writeJSON(w, http.StatusOK, mcpServerToResponse(server))
}

func (h *Handler) DeleteMCPServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workspaceID := resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}

	// Pre-read is required for workspace-scoped authorization: DeleteMCPServer SQL
	// only filters by id, not workspace_id, so without this check a user could
	// delete an MCP server belonging to another workspace.
	_, err := h.Queries.GetMCPServerInWorkspace(r.Context(), db.GetMCPServerInWorkspaceParams{
		ID:          parseUUID(id),
		WorkspaceID: parseUUID(workspaceID),
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "MCP server not found")
		return
	}

	if err := h.Queries.DeleteMCPServer(r.Context(), parseUUID(id)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete MCP server")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Agent-MCP Binding handlers ---

type AgentMCPBindingResponse struct {
	AgentID     string `json:"agent_id"`
	MCPServerID string `json:"mcp_server_id"`
	Enabled     bool   `json:"enabled"`
	SortOrder   int32  `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
}

type ReplaceAgentMCPBindingsRequest struct {
	MCPServerIDs []string `json:"mcp_server_ids"`
}

func bindingToResponse(b db.AgentMcpBinding) AgentMCPBindingResponse {
	return AgentMCPBindingResponse{
		AgentID:     uuidToString(b.AgentID),
		MCPServerID: uuidToString(b.McpServerID),
		Enabled:     b.Enabled,
		SortOrder:   b.SortOrder,
		CreatedAt:   timestampToString(b.CreatedAt),
	}
}

func (h *Handler) GetAgentMCPBindings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	agent, ok := h.loadAgentForUser(w, r, id)
	if !ok {
		return
	}

	bindings, err := h.Queries.ListAgentMCPBindings(r.Context(), agent.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list agent MCP bindings")
		return
	}

	resp := make([]AgentMCPBindingResponse, len(bindings))
	for i, b := range bindings {
		resp[i] = bindingToResponse(b)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ReplaceAgentMCPBindings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	agent, ok := h.loadAgentForUser(w, r, id)
	if !ok {
		return
	}
	if !h.canManageAgent(w, r, agent) {
		return
	}

	var req ReplaceAgentMCPBindingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := resolveWorkspaceID(r)

	// Verify all MCP server IDs exist in the workspace.
	for _, serverID := range req.MCPServerIDs {
		_, err := h.Queries.GetMCPServerInWorkspace(r.Context(), db.GetMCPServerInWorkspaceParams{
			ID:          parseUUID(serverID),
			WorkspaceID: parseUUID(workspaceID),
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("MCP server %s not found in workspace", serverID))
			return
		}
	}

	tx, err := h.TxStarter.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	qtx := h.Queries.WithTx(tx)

	if err := qtx.DeleteAgentMCPBindings(r.Context(), agent.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to clear agent MCP bindings")
		return
	}

	for i, serverID := range req.MCPServerIDs {
		_, err := qtx.CreateAgentMCPBinding(r.Context(), db.CreateAgentMCPBindingParams{
			AgentID:     agent.ID,
			McpServerID: parseUUID(serverID),
			Enabled:     true,
			SortOrder:   int32(i),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create agent MCP binding: "+err.Error())
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit")
		return
	}

	// Return the updated bindings list.
	bindings, err := h.Queries.ListAgentMCPBindings(r.Context(), agent.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list agent MCP bindings")
		return
	}

	resp := make([]AgentMCPBindingResponse, len(bindings))
	for i, b := range bindings {
		resp[i] = bindingToResponse(b)
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// Binding-driven MCP aggregation for daemon claim
// ---------------------------------------------------------------------------

// mcpVarPattern matches ${VAR_NAME} placeholders in MCP config strings.
var mcpVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// buildMCPConfigFromBindings queries the agent's bound MCP servers and
// aggregates them into the Claude-native format:
//
//	{"mcpServers": {"<name>": <resolved_config>, ...}}
//
// ${VAR} placeholders in each MCP server's config are resolved from the
// provided env map. Missing variables cause an explicit error naming the
// MCP server, the variable, and the agent.
func (h *Handler) buildMCPConfigFromBindings(ctx context.Context, agentID pgtype.UUID, env map[string]string) (any, error) {
	servers, err := h.Queries.GetBoundMCPServersForAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("query MCP bindings: %w", err)
	}
	if len(servers) == 0 {
		return nil, nil
	}

	agentIDStr := uuidToString(agentID)
	mcpServers := make(map[string]any, len(servers))
	for _, srv := range servers {
		resolved, resolveErr := handlerResolveMCPVars(srv.Config, env)
		if resolveErr != nil {
			slog.Warn("MCP config has unresolved variables",
				"agent_id", agentIDStr,
				"mcp_name", srv.Name,
				"error", resolveErr,
			)
			return nil, fmt.Errorf("agent %s: MCP server %q: %w", agentIDStr, srv.Name, resolveErr)
		}

		var configObj any
		if err := json.Unmarshal(resolved, &configObj); err != nil {
			return nil, fmt.Errorf("agent %s: MCP server %q: invalid JSON config: %w", agentIDStr, srv.Name, err)
		}
		mcpServers[srv.Name] = configObj
	}

	return map[string]any{"mcpServers": mcpServers}, nil
}

// handlerResolveMCPVars walks a JSON config recursively and replaces ${VAR}
// placeholders with values from the env map. Returns an error listing all
// variables that could not be resolved.
func handlerResolveMCPVars(config []byte, env map[string]string) ([]byte, error) {
	if env == nil {
		env = map[string]string{}
	}

	var root any
	if err := json.Unmarshal(config, &root); err != nil {
		return nil, fmt.Errorf("invalid JSON config: %w", err)
	}

	var missing []string
	resolved := mcpResolveValue(root, env, &missing)

	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("unresolved ${%s} in MCP config", strings.Join(missing, "}, ${"))
	}

	return json.Marshal(resolved)
}

// mcpResolveValue recursively walks a JSON value and resolves placeholders in strings.
func mcpResolveValue(v any, env map[string]string, missing *[]string) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			result[k] = mcpResolveValue(v, env, missing)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, elem := range val {
			result[i] = mcpResolveValue(elem, env, missing)
		}
		return result
	case string:
		return mcpResolveString(val, env, missing)
	default:
		return v
	}
}

// mcpResolveString replaces ${VAR} patterns in a single string value.
func mcpResolveString(s string, env map[string]string, missing *[]string) string {
	if !strings.Contains(s, "${") {
		return s
	}

	vars := mcpVarPattern.FindAllStringSubmatch(s, -1)
	seen := make(map[string]bool, len(vars))
	for _, match := range vars {
		name := match[1]
		if _, ok := env[name]; !ok && !seen[name] {
			*missing = append(*missing, name)
			seen[name] = true
		}
	}

	if len(*missing) > 0 {
		return s
	}

	return mcpVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		sub := mcpVarPattern.FindStringSubmatch(match)
		if len(sub) >= 2 {
			return env[sub[1]]
		}
		return match
	})
}
