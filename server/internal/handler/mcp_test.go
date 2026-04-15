package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMCPServerCRUD(t *testing.T) {
	// Create
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name":        "Test MCP Server",
		"description": "A test MCP server",
		"config": map[string]any{
			"command": "node",
			"args":    []string{"server.js"},
			"env":     map[string]string{"KEY": "value"},
		},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreateMCPServer: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created MCPServerResponse
	json.NewDecoder(w.Body).Decode(&created)
	if created.Name != "Test MCP Server" {
		t.Fatalf("CreateMCPServer: expected name 'Test MCP Server', got '%s'", created.Name)
	}
	if created.Description != "A test MCP server" {
		t.Fatalf("CreateMCPServer: expected description 'A test MCP server', got '%s'", created.Description)
	}
	if created.WorkspaceID != testWorkspaceID {
		t.Fatalf("CreateMCPServer: expected workspace_id '%s', got '%s'", testWorkspaceID, created.WorkspaceID)
	}
	if created.Config == nil {
		t.Fatal("CreateMCPServer: expected non-nil config")
	}
	serverID := created.ID

	// Get
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/mcp-servers/"+serverID, nil)
	req = withURLParam(req, "id", serverID)
	testHandler.GetMCPServer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetMCPServer: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var fetched MCPServerResponse
	json.NewDecoder(w.Body).Decode(&fetched)
	if fetched.ID != serverID {
		t.Fatalf("GetMCPServer: expected id '%s', got '%s'", serverID, fetched.ID)
	}
	if fetched.Name != "Test MCP Server" {
		t.Fatalf("GetMCPServer: expected name 'Test MCP Server', got '%s'", fetched.Name)
	}

	// Update
	w = httptest.NewRecorder()
	req = newRequest("PUT", "/api/mcp-servers/"+serverID, map[string]any{
		"name":        "Updated MCP Server",
		"description": "Updated description",
		"config": map[string]any{
			"command": "python",
			"args":    []string{"-m", "mcp_server"},
		},
	})
	req = withURLParam(req, "id", serverID)
	testHandler.UpdateMCPServer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("UpdateMCPServer: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated MCPServerResponse
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Name != "Updated MCP Server" {
		t.Fatalf("UpdateMCPServer: expected name 'Updated MCP Server', got '%s'", updated.Name)
	}
	if updated.Description != "Updated description" {
		t.Fatalf("UpdateMCPServer: expected description 'Updated description', got '%s'", updated.Description)
	}

	// List
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/mcp-servers?workspace_id="+testWorkspaceID, nil)
	testHandler.ListMCPServers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListMCPServers: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var listResp []MCPServerResponse
	json.NewDecoder(w.Body).Decode(&listResp)
	if len(listResp) == 0 {
		t.Fatal("ListMCPServers: expected at least 1 MCP server")
	}
	found := false
	for _, s := range listResp {
		if s.ID == serverID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListMCPServers: expected to find server '%s' in list", serverID)
	}

	// Delete
	w = httptest.NewRecorder()
	req = newRequest("DELETE", "/api/mcp-servers/"+serverID, nil)
	req = withURLParam(req, "id", serverID)
	testHandler.DeleteMCPServer(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DeleteMCPServer: expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify deleted
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/mcp-servers/"+serverID, nil)
	req = withURLParam(req, "id", serverID)
	testHandler.GetMCPServer(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("GetMCPServer after delete: expected 404, got %d", w.Code)
	}
}

func TestCreateMCPServerValidation(t *testing.T) {
	tests := []struct {
		name    string
		body    map[string]any
		wantCode int
	}{
		{
			name:    "missing name",
			body:    map[string]any{"config": map[string]any{}},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "config is array",
			body: map[string]any{
				"name":   "Bad Config",
				"config": []string{"not", "an", "object"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "config is string",
			body: map[string]any{
				"name":   "Bad Config",
				"config": "not an object",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "config is number",
			body: map[string]any{
				"name":   "Bad Config",
				"config": 42,
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "args is not array",
			body: map[string]any{
				"name": "Bad Args",
				"config": map[string]any{
					"args": "not-an-array",
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "env is not object",
			body: map[string]any{
				"name": "Bad Env",
				"config": map[string]any{
					"env": []string{"not", "an", "object"},
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "valid config with args and env",
			body: map[string]any{
				"name":        "Valid Server",
				"description": "desc",
				"config": map[string]any{
					"command": "node",
					"args":    []string{"server.js"},
					"env":     map[string]string{"KEY": "val"},
				},
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "valid config empty object",
			body: map[string]any{
				"name":   "Empty Config",
				"config": map[string]any{},
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "valid config without args or env",
			body: map[string]any{
				"name": "No Args Env",
				"config": map[string]any{
					"command": "node",
				},
			},
			wantCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, tt.body)
			testHandler.CreateMCPServer(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			// Clean up if created
			if tt.wantCode == http.StatusCreated {
				var resp MCPServerResponse
				json.NewDecoder(w.Body).Decode(&resp)
				delReq := newRequest("DELETE", "/api/mcp-servers/"+resp.ID, nil)
				delReq = withURLParam(delReq, "id", resp.ID)
				testHandler.DeleteMCPServer(httptest.NewRecorder(), delReq)
			}
		})
	}
}

func TestMCPServerWorkspaceIsolation(t *testing.T) {
	// Create MCP server in test workspace
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name": "Isolated Server",
		"config": map[string]any{
			"command": "node",
		},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreateMCPServer: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created MCPServerResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Try to get with wrong workspace (simulate different workspace)
	// Create a chi route context with the server ID but set a different workspace
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/mcp-servers/"+created.ID, nil)
	req = withURLParam(req, "id", created.ID)
	// Override workspace to a non-member workspace
	req.Header.Set("X-Workspace-ID", "00000000-0000-0000-0000-000000000099")
	testHandler.GetMCPServer(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("GetMCPServer with wrong workspace: expected 404, got %d: %s", w.Code, w.Body.String())
	}

	// Clean up
	delReq := newRequest("DELETE", "/api/mcp-servers/"+created.ID, nil)
	delReq = withURLParam(delReq, "id", created.ID)
	testHandler.DeleteMCPServer(httptest.NewRecorder(), delReq)
}

func TestUpdateMCPServerPartial(t *testing.T) {
	// Create
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name":        "Partial Update Server",
		"description": "Original description",
		"config": map[string]any{
			"command": "node",
		},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreateMCPServer: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created MCPServerResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Update only name (partial)
	w = httptest.NewRecorder()
	req = newRequest("PUT", "/api/mcp-servers/"+created.ID, map[string]any{
		"name": "Partially Updated",
	})
	req = withURLParam(req, "id", created.ID)
	testHandler.UpdateMCPServer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("UpdateMCPServer: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated MCPServerResponse
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Name != "Partially Updated" {
		t.Fatalf("UpdateMCPServer: expected name 'Partially Updated', got '%s'", updated.Name)
	}
	// Description and config should be preserved
	if updated.Description != "Original description" {
		t.Fatalf("UpdateMCPServer: expected description preserved, got '%s'", updated.Description)
	}

	// Clean up
	delReq := newRequest("DELETE", "/api/mcp-servers/"+created.ID, nil)
	delReq = withURLParam(delReq, "id", created.ID)
	testHandler.DeleteMCPServer(httptest.NewRecorder(), delReq)
}

func TestListMCPServersEmpty(t *testing.T) {
	ctx := context.Background()

	// Create a temporary workspace for this test
	var tmpWorkspaceID string
	err := testPool.QueryRow(ctx, `
		INSERT INTO workspace (name, slug, description, issue_prefix)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "MCP Empty Test", "mcp-empty-test", "Temporary for empty list test", "MCP").Scan(&tmpWorkspaceID)
	if err != nil {
		t.Fatalf("failed to create temp workspace: %v", err)
	}

	// Add test user as member
	_, err = testPool.Exec(ctx, `
		INSERT INTO member (workspace_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, tmpWorkspaceID, testUserID)
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	t.Cleanup(func() {
		testPool.Exec(context.Background(), `DELETE FROM workspace WHERE id = $1`, tmpWorkspaceID)
	})

	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/mcp-servers?workspace_id="+tmpWorkspaceID, nil)
	testHandler.ListMCPServers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ListMCPServers empty: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var listResp []MCPServerResponse
	json.NewDecoder(w.Body).Decode(&listResp)
	if len(listResp) != 0 {
		t.Fatalf("ListMCPServers empty: expected 0, got %d", len(listResp))
	}
}

func TestMCPServerRequiresAuth(t *testing.T) {
	// Create request without user ID
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mcp-servers?workspace_id="+testWorkspaceID, nil)
	req.Header.Set("Content-Type", "application/json")
	// No X-User-ID header
	testHandler.ListMCPServers(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("ListMCPServers without auth: expected 401, got %d", w.Code)
	}
}

func TestMCPServerRequiresWorkspaceID(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mcp-servers", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", testUserID)
	// No workspace ID in query, header, or context
	testHandler.ListMCPServers(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ListMCPServers without workspace_id: expected 400, got %d", w.Code)
	}
}

func TestGetMCPServerNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/mcp-servers/00000000-0000-0000-0000-000000000099", nil)
	req = withURLParam(req, "id", "00000000-0000-0000-0000-000000000099")
	testHandler.GetMCPServer(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("GetMCPServer not found: expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateMCPServerValidation(t *testing.T) {
	// Create a server first
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name": "Validation Target",
		"config": map[string]any{
			"command": "node",
		},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreateMCPServer: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created MCPServerResponse
	json.NewDecoder(w.Body).Decode(&created)

	tests := []struct {
		name     string
		body     map[string]any
		wantCode int
	}{
		{
			name: "config is array",
			body: map[string]any{
				"config": []string{"bad"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "args is not array",
			body: map[string]any{
				"config": map[string]any{
					"args": "string",
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "env is not object",
			body: map[string]any{
				"config": map[string]any{
					"env": "string",
				},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "valid partial update",
			body: map[string]any{
				"name": "New Name",
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := newRequest("PUT", "/api/mcp-servers/"+created.ID, tt.body)
			req = withURLParam(req, "id", created.ID)
			testHandler.UpdateMCPServer(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("expected %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}
		})
	}

	// Clean up
	delReq := newRequest("DELETE", "/api/mcp-servers/"+created.ID, nil)
	delReq = withURLParam(delReq, "id", created.ID)
	testHandler.DeleteMCPServer(httptest.NewRecorder(), delReq)
}

// TestMCPServerRouteRegistration verifies that the MCP server routes are wired
// correctly by exercising them through a Chi router mux.
func TestMCPServerRouteRegistration(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/mcp-servers", func(r chi.Router) {
		r.Get("/", testHandler.ListMCPServers)
		r.Post("/", testHandler.CreateMCPServer)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", testHandler.GetMCPServer)
			r.Put("/", testHandler.UpdateMCPServer)
			r.Delete("/", testHandler.DeleteMCPServer)
		})
	})

	// Test GET /api/mcp-servers
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/mcp-servers?workspace_id="+testWorkspaceID, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("route GET /: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Test POST /api/mcp-servers
	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name": "Routed Server",
		"config": map[string]any{
			"command": "test",
		},
	})
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("route POST /: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created MCPServerResponse
	json.NewDecoder(w.Body).Decode(&created)

	// Test GET /api/mcp-servers/{id}
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/mcp-servers/"+created.ID, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("route GET /{id}: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Test DELETE /api/mcp-servers/{id}
	w = httptest.NewRecorder()
	req = newRequest("DELETE", "/api/mcp-servers/"+created.ID, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("route DELETE /{id}: expected 204, got %d: %s", w.Code, w.Body.String())
	}
}
