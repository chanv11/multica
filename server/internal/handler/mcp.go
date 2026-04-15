package handler

import (
	"encoding/json"
	"net/http"

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
