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

func TestCreateMCPServerDuplicateName(t *testing.T) {
	// Create first server
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name":        "Duplicate Test Server",
		"description": "First server",
		"config":      map[string]any{"command": "node"},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var first MCPServerResponse
	json.NewDecoder(w.Body).Decode(&first)

	// Try to create a second server with the same name in the same workspace
	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name":        "Duplicate Test Server",
		"description": "Second server",
		"config":      map[string]any{"command": "python"},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate create: expected 409, got %d: %s", w.Code, w.Body.String())
	}

	// Clean up
	delReq := newRequest("DELETE", "/api/mcp-servers/"+first.ID, nil)
	delReq = withURLParam(delReq, "id", first.ID)
	testHandler.DeleteMCPServer(httptest.NewRecorder(), delReq)
}

func TestUpdateMCPServerDuplicateName(t *testing.T) {
	// Create two servers with different names
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name":   "Server A",
		"config": map[string]any{"command": "node"},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create server A: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var serverA MCPServerResponse
	json.NewDecoder(w.Body).Decode(&serverA)

	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name":   "Server B",
		"config": map[string]any{"command": "python"},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create server B: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var serverB MCPServerResponse
	json.NewDecoder(w.Body).Decode(&serverB)

	// Try to rename server B to server A's name
	w = httptest.NewRecorder()
	req = newRequest("PUT", "/api/mcp-servers/"+serverB.ID, map[string]any{
		"name": "Server A",
	})
	req = withURLParam(req, "id", serverB.ID)
	testHandler.UpdateMCPServer(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("update with duplicate name: expected 409, got %d: %s", w.Code, w.Body.String())
	}

	// Clean up
	for _, id := range []string{serverA.ID, serverB.ID} {
		delReq := newRequest("DELETE", "/api/mcp-servers/"+id, nil)
		delReq = withURLParam(delReq, "id", id)
		testHandler.DeleteMCPServer(httptest.NewRecorder(), delReq)
	}
}

func TestUpdateMCPServerNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("PUT", "/api/mcp-servers/00000000-0000-0000-0000-000000000099", map[string]any{
		"name": "Does Not Exist",
	})
	req = withURLParam(req, "id", "00000000-0000-0000-0000-000000000099")
	testHandler.UpdateMCPServer(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("update non-existent: expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Agent MCP Binding tests ---

// helperAgentID returns the ID of the agent created by the test fixture.
func helperAgentID(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	var agentID string
	err := testPool.QueryRow(ctx,
		`SELECT id FROM agent WHERE workspace_id = $1 AND name = $2`,
		testWorkspaceID, "Handler Test Agent",
	).Scan(&agentID)
	if err != nil {
		t.Fatalf("failed to find test agent: %v", err)
	}
	return agentID
}

// helperCreateMCPServer creates an MCP server in the test workspace and returns its ID.
func helperCreateMCPServer(t *testing.T, name string) string {
	t.Helper()
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/mcp-servers?workspace_id="+testWorkspaceID, map[string]any{
		"name":        name,
		"description": "test server",
		"config":      map[string]any{"command": "node"},
	})
	testHandler.CreateMCPServer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create MCP server: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp MCPServerResponse
	json.NewDecoder(w.Body).Decode(&resp)
	return resp.ID
}

// helperDeleteMCPServer removes an MCP server by ID.
func helperDeleteMCPServer(t *testing.T, id string) {
	t.Helper()
	req := newRequest("DELETE", "/api/mcp-servers/"+id, nil)
	req = withURLParam(req, "id", id)
	testHandler.DeleteMCPServer(httptest.NewRecorder(), req)
}

func TestGetAgentMCPBindingsEmpty(t *testing.T) {
	agentID := helperAgentID(t)

	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/agents/"+agentID+"/mcp-bindings", nil)
	req = withURLParam(req, "id", agentID)
	testHandler.GetAgentMCPBindings(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetAgentMCPBindings: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var bindings []AgentMCPBindingResponse
	json.NewDecoder(w.Body).Decode(&bindings)
	if len(bindings) != 0 {
		t.Fatalf("expected 0 bindings, got %d", len(bindings))
	}
}

func TestReplaceAgentMCPBindings(t *testing.T) {
	agentID := helperAgentID(t)
	server1 := helperCreateMCPServer(t, "Binding Test Server 1")
	server2 := helperCreateMCPServer(t, "Binding Test Server 2")
	defer func() {
		helperDeleteMCPServer(t, server1)
		helperDeleteMCPServer(t, server2)
	}()

	// Replace with two servers
	w := httptest.NewRecorder()
	req := newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{server1, server2},
	})
	req = withURLParam(req, "id", agentID)
	testHandler.ReplaceAgentMCPBindings(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ReplaceAgentMCPBindings: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var bindings []AgentMCPBindingResponse
	json.NewDecoder(w.Body).Decode(&bindings)
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
	if bindings[0].MCPServerID != server1 {
		t.Fatalf("expected first binding server_id %s, got %s", server1, bindings[0].MCPServerID)
	}
	if bindings[0].SortOrder != 0 {
		t.Fatalf("expected first binding sort_order 0, got %d", bindings[0].SortOrder)
	}
	if bindings[1].MCPServerID != server2 {
		t.Fatalf("expected second binding server_id %s, got %s", server2, bindings[1].MCPServerID)
	}
	if bindings[1].SortOrder != 1 {
		t.Fatalf("expected second binding sort_order 1, got %d", bindings[1].SortOrder)
	}
	if !bindings[0].Enabled || !bindings[1].Enabled {
		t.Fatal("expected all bindings to be enabled")
	}

	// Verify GET returns the same
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/agents/"+agentID+"/mcp-bindings", nil)
	req = withURLParam(req, "id", agentID)
	testHandler.GetAgentMCPBindings(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetAgentMCPBindings: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	json.NewDecoder(w.Body).Decode(&bindings)
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings from GET, got %d", len(bindings))
	}
}

func TestReplaceAgentMCPBindingsOverwrites(t *testing.T) {
	agentID := helperAgentID(t)
	server1 := helperCreateMCPServer(t, "Overwrite Server 1")
	server2 := helperCreateMCPServer(t, "Overwrite Server 2")
	defer func() {
		helperDeleteMCPServer(t, server1)
		helperDeleteMCPServer(t, server2)
	}()

	// Set initial binding
	w := httptest.NewRecorder()
	req := newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{server1},
	})
	req = withURLParam(req, "id", agentID)
	testHandler.ReplaceAgentMCPBindings(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first replace: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Replace with different set
	w = httptest.NewRecorder()
	req = newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{server2},
	})
	req = withURLParam(req, "id", agentID)
	testHandler.ReplaceAgentMCPBindings(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("second replace: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var bindings []AgentMCPBindingResponse
	json.NewDecoder(w.Body).Decode(&bindings)
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding after overwrite, got %d", len(bindings))
	}
	if bindings[0].MCPServerID != server2 {
		t.Fatalf("expected binding to server2, got %s", bindings[0].MCPServerID)
	}
}

func TestReplaceAgentMCPBindingsEmptyList(t *testing.T) {
	agentID := helperAgentID(t)
	server1 := helperCreateMCPServer(t, "Empty List Server")
	defer helperDeleteMCPServer(t, server1)

	// Set initial binding
	w := httptest.NewRecorder()
	req := newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{server1},
	})
	req = withURLParam(req, "id", agentID)
	testHandler.ReplaceAgentMCPBindings(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("initial replace: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Clear bindings with empty list
	w = httptest.NewRecorder()
	req = newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{},
	})
	req = withURLParam(req, "id", agentID)
	testHandler.ReplaceAgentMCPBindings(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("clear replace: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var bindings []AgentMCPBindingResponse
	json.NewDecoder(w.Body).Decode(&bindings)
	if len(bindings) != 0 {
		t.Fatalf("expected 0 bindings after clear, got %d", len(bindings))
	}
}

func TestReplaceAgentMCPBindingsInvalidServerID(t *testing.T) {
	agentID := helperAgentID(t)

	w := httptest.NewRecorder()
	req := newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{"00000000-0000-0000-0000-000000000099"},
	})
	req = withURLParam(req, "id", agentID)
	testHandler.ReplaceAgentMCPBindings(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid server ID, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetAgentMCPBindingsAgentNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/agents/00000000-0000-0000-0000-000000000099/mcp-bindings", nil)
	req = withURLParam(req, "id", "00000000-0000-0000-0000-000000000099")
	testHandler.GetAgentMCPBindings(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent agent, got %d: %s", w.Code, w.Body.String())
	}
}

func TestReplaceAgentMCPBindingsAgentNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("PUT", "/api/agents/00000000-0000-0000-0000-000000000099/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{},
	})
	req = withURLParam(req, "id", "00000000-0000-0000-0000-000000000099")
	testHandler.ReplaceAgentMCPBindings(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent agent, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentMCPBindingRouteRegistration(t *testing.T) {
	agentID := helperAgentID(t)
	server1 := helperCreateMCPServer(t, "Route Test Server")
	defer helperDeleteMCPServer(t, server1)

	r := chi.NewRouter()
	r.Route("/api/agents/{id}", func(r chi.Router) {
		r.Get("/mcp-bindings", testHandler.GetAgentMCPBindings)
		r.Put("/mcp-bindings", testHandler.ReplaceAgentMCPBindings)
	})

	// Test PUT via router
	w := httptest.NewRecorder()
	req := newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{server1},
	})
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("route PUT /mcp-bindings: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var bindings []AgentMCPBindingResponse
	json.NewDecoder(w.Body).Decode(&bindings)
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	// Test GET via router
	w = httptest.NewRecorder()
	req = newRequest("GET", "/api/agents/"+agentID+"/mcp-bindings", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("route GET /mcp-bindings: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Clean up bindings
	clearReq := newRequest("PUT", "/api/agents/"+agentID+"/mcp-bindings", map[string]any{
		"mcp_server_ids": []string{},
	})
	clearReq = withURLParam(clearReq, "id", agentID)
	testHandler.ReplaceAgentMCPBindings(httptest.NewRecorder(), clearReq)
}
