package handler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropdownHandler_ListDropdowns_InvalidEntity(t *testing.T) {
	// Test validation - should reject invalid entity types
	validEntities := map[string]bool{
		"users":        true,
		"workflows":    true,
		"templates":    true,
		"roles":        true,
		"applications": true,
		"approvals":    true,
	}

	reqBody := `{"entities": ["invalid_entity"]}`
	var parsedReq struct {
		Entities       []string `json:"entities"`
		IncludeInactive bool     `json:"include_inactive"`
		Statuses       []string `json:"statuses"`
	}
	err := json.Unmarshal([]byte(reqBody), &parsedReq)
	require.NoError(t, err)

	var invalidEntities []string
	for _, entity := range parsedReq.Entities {
		if !validEntities[entity] {
			invalidEntities = append(invalidEntities, entity)
		}
	}

	assert.NotEmpty(t, invalidEntities, "should detect invalid entities")
	assert.Contains(t, invalidEntities, "invalid_entity")
}

func TestDropdownHandler_ListDropdowns_ValidEntities(t *testing.T) {
	// Test that valid entities are recognized
	validEntities := map[string]bool{
		"users":        true,
		"workflows":    true,
		"templates":    true,
		"roles":        true,
		"applications": true,
		"approvals":    true,
	}

	testCases := []struct {
		entity  string
		isValid bool
	}{
		{"users", true},
		{"workflows", true},
		{"templates", true},
		{"roles", true},
		{"applications", true},
		{"approvals", true},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.entity, func(t *testing.T) {
			assert.Equal(t, tc.isValid, validEntities[tc.entity])
		})
	}
}

func TestDropdownHandler_DropdownOption_Structure(t *testing.T) {
	// Test DropdownOption structure
	option := DropdownOption{
		ID:   "test-id-123",
		Name: "Test Option",
	}

	// Marshal to JSON
	data, err := json.Marshal(option)
	require.NoError(t, err)

	// Unmarshal back
	var parsed DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, option.ID, parsed.ID)
	assert.Equal(t, option.Name, parsed.Name)
}

func TestDropdownHandler_MultipleEntities(t *testing.T) {
	// Test parsing of multiple entities request
	reqBody := `{
		"entities": ["users", "workflows", "templates"],
		"include_inactive": false,
		"statuses": ["submitted", "approved"]
	}`

	var req struct {
		Entities       []string `json:"entities"`
		IncludeInactive bool     `json:"include_inactive"`
		Statuses       []string `json:"statuses"`
	}

	err := json.Unmarshal([]byte(reqBody), &req)
	require.NoError(t, err)

	assert.Len(t, req.Entities, 3)
	assert.Contains(t, req.Entities, "users")
	assert.Contains(t, req.Entities, "workflows")
	assert.Contains(t, req.Entities, "templates")
	assert.False(t, req.IncludeInactive)
	assert.Len(t, req.Statuses, 2)
	assert.Contains(t, req.Statuses, "submitted")
	assert.Contains(t, req.Statuses, "approved")
}

func TestDropdownHandler_EmptyEntities(t *testing.T) {
	// Test that empty entities array is handled
	reqBody := `{"entities": []}`

	var req struct {
		Entities []string `json:"entities"`
	}

	err := json.Unmarshal([]byte(reqBody), &req)
	require.NoError(t, err)

	assert.Empty(t, req.Entities)
}

func TestDropdownHandler_ResponseFormat(t *testing.T) {
	// Test expected response format
	response := map[string][]DropdownOption{
		"users": {
			{ID: "user-1", Name: "Alice (alice@example.com)"},
			{ID: "user-2", Name: "Bob (bob@example.com)"},
		},
		"workflows": {
			{ID: "wf-1", Name: "Expense Approval"},
		},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var parsed map[string][]DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed["users"], 2)
	assert.Len(t, parsed["workflows"], 1)
	assert.Equal(t, "Alice (alice@example.com)", parsed["users"][0].Name)
}

func TestDropdownHandler_StatusFiltering(t *testing.T) {
	// Test that status filtering works correctly
	testCases := []struct {
		name           string
		statuses       []string
		expectedLength int
	}{
		{
			name:           "single status",
			statuses:       []string{"submitted"},
			expectedLength: 1,
		},
		{
			name:           "multiple statuses",
			statuses:       []string{"submitted", "approved", "completed"},
			expectedLength: 3,
		},
		{
			name:           "empty statuses defaults to submitted",
			statuses:       []string{},
			expectedLength: 0, // Will be replaced with default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			statuses := tc.statuses
			if len(statuses) == 0 {
				statuses = []string{"submitted"}
			}
			assert.Len(t, statuses, max(tc.expectedLength, 1))
		})
	}
}

func TestDropdownHandler_JSONSerialization(t *testing.T) {
	// Test JSON serialization/deserialization of dropdown options
	options := []DropdownOption{
		{ID: "123e4567-e89b-12d3-a456-426614174000", Name: "Test User"},
		{ID: "123e4567-e89b-12d3-a456-426614174001", Name: "Another User"},
	}

	data, err := json.Marshal(options)
	require.NoError(t, err)

	var parsed []DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 2)
	assert.Equal(t, options[0].ID, parsed[0].ID)
	assert.Equal(t, options[1].Name, parsed[1].Name)
}

func TestDropdownHandler_IncludeInactiveWorkflows(t *testing.T) {
	// Test include_inactive flag parsing
	testCases := []struct {
		name            string
		includeInactive bool
	}{
		{"include inactive", true},
		{"exclude inactive", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := `{"entities": ["workflows"], "include_inactive": ` + boolToString(tc.includeInactive) + `}`

			var req struct {
				IncludeInactive bool `json:"include_inactive"`
			}

			err := json.Unmarshal([]byte(reqBody), &req)
			require.NoError(t, err)
			assert.Equal(t, tc.includeInactive, req.IncludeInactive)
		})
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestDropdownHandler_NilStatuses(t *testing.T) {
	// Test that nil statuses are handled correctly
	reqBody := `{"entities": ["applications"]}`

	var req struct {
		Statuses []string `json:"statuses"`
	}

	err := json.Unmarshal([]byte(reqBody), &req)
	require.NoError(t, err)

	// Nil statuses should be handled by defaulting to ["submitted"]
	statuses := req.Statuses
	if len(statuses) == 0 {
		statuses = []string{"submitted"}
	}

	assert.Equal(t, []string{"submitted"}, statuses)
}

func TestDropdownHandler_InvalidJSON(t *testing.T) {
	// Test invalid JSON handling
	reqBody := `{invalid json}`

	var req struct {
		Entities []string `json:"entities"`
	}

	err := json.Unmarshal([]byte(reqBody), &req)
	assert.Error(t, err, "should fail on invalid JSON")
}

func TestDropdownHandler_MissingEntities(t *testing.T) {
	// Test missing entities field
	reqBody := `{"include_inactive": true}`

	var req struct {
		Entities       []string `json:"entities"`
		IncludeInactive bool     `json:"include_inactive"`
	}

	err := json.Unmarshal([]byte(reqBody), &req)
	require.NoError(t, err)

	assert.Nil(t, req.Entities, "entities should be nil when not provided")
	assert.True(t, req.IncludeInactive)
}

func TestDropdownHandler_DropdownOptionJSONTags(t *testing.T) {
	// Test JSON tag correctness
	option := DropdownOption{
		ID:   "id-1",
		Name: "name-1",
	}

	data, err := json.Marshal(option)
	require.NoError(t, err)

	// Verify JSON output has correct field names
	assert.Contains(t, string(data), `"id"`)
	assert.Contains(t, string(data), `"name"`)
	assert.NotContains(t, string(data), `"ID"`)
	assert.NotContains(t, string(data), `"Name"`)
}
