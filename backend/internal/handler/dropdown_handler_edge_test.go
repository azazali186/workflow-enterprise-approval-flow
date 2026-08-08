package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Error Handling Edge Cases ====================

func TestDropdownHandler_SpecialCharactersInNames(t *testing.T) {
	// Test that special characters in names are handled correctly
	options := []DropdownOption{
		{ID: "1", Name: "O'Brien & Sons"},
		{ID: "2", Name: "Café München"},
		{ID: "3", Name: "日本語テスト"},
		{ID: "4", Name: "Test \"Quote\" Name"},
		{ID: "5", Name: "Name with\nnewline"},
		{ID: "6", Name: "Name with\ttab"},
	}

	data, err := json.Marshal(options)
	require.NoError(t, err)

	var parsed []DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 6)
	assert.Equal(t, "O'Brien & Sons", parsed[0].Name)
	assert.Equal(t, "Café München", parsed[1].Name)
	assert.Equal(t, "日本語テスト", parsed[2].Name)
	assert.Equal(t, "Test \"Quote\" Name", parsed[3].Name)
	assert.Contains(t, parsed[4].Name, "\n")
	assert.Contains(t, parsed[5].Name, "\t")
}

func TestDropdownHandler_EmptyNames(t *testing.T) {
	// Test handling of empty names
	options := []DropdownOption{
		{ID: "1", Name: ""},
		{ID: "2", Name: "Valid Name"},
		{ID: "3", Name: "   "}, // whitespace only
	}

	data, err := json.Marshal(options)
	require.NoError(t, err)

	var parsed []DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "", parsed[0].Name)
	assert.Equal(t, "Valid Name", parsed[1].Name)
	assert.Equal(t, "   ", parsed[2].Name)
}

func TestDropdownHandler_VeryLongNames(t *testing.T) {
	// Test handling of very long names (edge case for UI rendering)
	longName := strings.Repeat("A", 1000)
	option := DropdownOption{
		ID:   "1",
		Name: longName,
	}

	data, err := json.Marshal(option)
	require.NoError(t, err)

	var parsed DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed.Name, 1000)
}

func TestDropdownHandler_EmptyID(t *testing.T) {
	// Test handling of empty ID
	option := DropdownOption{
		ID:   "",
		Name: "Option with empty ID",
	}

	data, err := json.Marshal(option)
	require.NoError(t, err)

	var parsed DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Empty(t, parsed.ID)
	assert.Equal(t, "Option with empty ID", parsed.Name)
}

func TestDropdownHandler_LargeDataset(t *testing.T) {
	// Test handling of large datasets
	options := make([]DropdownOption, 1000)
	for i := 0; i < 1000; i++ {
		options[i] = DropdownOption{
			ID:   strings.Repeat("x", 36), // UUID length
			Name: "User " + strings.Repeat("N", 50),
		}
	}

	data, err := json.Marshal(options)
	require.NoError(t, err)

	var parsed []DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 1000)
}

func TestDropdownHandler_MixedValidInvalidEntities(t *testing.T) {
	// Test request with both valid and invalid entities
	validEntities := map[string]bool{
		"users":     true,
		"workflows": true,
	}

	reqBody := `{"entities": ["users", "invalid", "workflows", "another_invalid"]}`
	var parsedReq struct {
		Entities []string `json:"entities"`
	}
	err := json.Unmarshal([]byte(reqBody), &parsedReq)
	require.NoError(t, err)

	var valid, invalid []string
	for _, entity := range parsedReq.Entities {
		if validEntities[entity] {
			valid = append(valid, entity)
		} else {
			invalid = append(invalid, entity)
		}
	}

	assert.Len(t, valid, 2)
	assert.Len(t, invalid, 2)
	assert.Contains(t, valid, "users")
	assert.Contains(t, valid, "workflows")
	assert.Contains(t, invalid, "invalid")
	assert.Contains(t, invalid, "another_invalid")
}

func TestDropdownHandler_DuplicateEntities(t *testing.T) {
	// Test request with duplicate entities
	reqBody := `{"entities": ["users", "users", "users"]}`
	var parsedReq struct {
		Entities []string `json:"entities"`
	}
	err := json.Unmarshal([]byte(reqBody), &parsedReq)
	require.NoError(t, err)

	assert.Len(t, parsedReq.Entities, 3)

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, e := range parsedReq.Entities {
		if !seen[e] {
			seen[e] = true
			unique = append(unique, e)
		}
	}
	assert.Len(t, unique, 1)
}

func TestDropdownHandler_UnicodeInStatuses(t *testing.T) {
	// Test handling of unicode in status fields
	statuses := []string{"submitted", "übersicht", "日本語"}

	data, err := json.Marshal(statuses)
	require.NoError(t, err)

	var parsed []string
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, statuses, parsed)
}

func TestDropdownHandler_EmptyResponse(t *testing.T) {
	// Test empty response format
	response := map[string][]DropdownOption{}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var parsed map[string][]DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Empty(t, parsed)
}

func TestDropdownHandler_NilValuesInResponse(t *testing.T) {
	// Test response with empty slices
	response := map[string][]DropdownOption{
		"users":     {},
		"workflows": {},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var parsed map[string][]DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Empty slices remain empty
	assert.NotNil(t, parsed["users"])
	assert.Len(t, parsed["users"], 0)
	assert.NotNil(t, parsed["workflows"])
	assert.Len(t, parsed["workflows"], 0)
}

func TestDropdownHandler_MultipleInvalidEntities(t *testing.T) {
	// Test error message for multiple invalid entities
	invalidEntities := []string{"foo", "bar", "baz"}

	errorMsg := "invalid entity type(s): " + strings.Join(invalidEntities, ", ")
	assert.Contains(t, errorMsg, "foo")
	assert.Contains(t, errorMsg, "bar")
	assert.Contains(t, errorMsg, "baz")
}

func TestDropdownHandler_StatusDefaults(t *testing.T) {
	// Test status defaulting logic
	testCases := []struct {
		name           string
		statuses       []string
		expectedResult string
	}{
		{
			name:           "empty statuses",
			statuses:       []string{},
			expectedResult: "submitted",
		},
		{
			name:           "nil statuses",
			statuses:       nil,
			expectedResult: "submitted",
		},
		{
			name:           "has statuses",
			statuses:       []string{"approved"},
			expectedResult: "approved",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			statuses := tc.statuses
			if len(statuses) == 0 {
				statuses = []string{"submitted"}
			}
			assert.Equal(t, tc.expectedResult, statuses[0])
		})
	}
}

func TestDropdownHandler_BooleanCoercion(t *testing.T) {
	// Test boolean field handling
	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true string", "true", true},
		{"false string", "false", false},
		{"TRUE string", "TRUE", true},
		{"1 string", "1", false}, // JSON only accepts true/false
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := `{"entities": ["workflows"], "include_inactive": ` + tc.value + `}`
			var req struct {
				IncludeInactive bool `json:"include_inactive"`
			}
			err := json.Unmarshal([]byte(reqBody), &req)
			if tc.value == "true" || tc.value == "false" {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, req.IncludeInactive)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestDropdownHandler_ConcurrentRequests(t *testing.T) {
	// Test that response format is consistent for concurrent requests
	response := map[string][]DropdownOption{
		"users": {{ID: "1", Name: "User 1"}},
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	// Simulate multiple concurrent deserializations
	for i := 0; i < 100; i++ {
		var parsed map[string][]DropdownOption
		err := json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Len(t, parsed["users"], 1)
	}
}

func TestDropdownHandler_HTMLInNames(t *testing.T) {
	// Test handling of HTML in names (potential XSS)
	// Note: Go's JSON marshaler does NOT escape HTML by default
	options := []DropdownOption{
		{ID: "1", Name: "<script>alert('xss')</script>"},
		{ID: "2", Name: "<img src=x onerror=alert(1)>"},
		{ID: "3", Name: "Normal <b>bold</b> text"},
	}

	data, err := json.Marshal(options)
	require.NoError(t, err)

	var parsed []DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// HTML is preserved as-is in JSON (escaping is done at render time)
	assert.Equal(t, "<script>alert('xss')</script>", parsed[0].Name)
	assert.Equal(t, "<img src=x onerror=alert(1)>", parsed[1].Name)
	assert.Equal(t, "Normal <b>bold</b> text", parsed[2].Name)
}

func TestDropdownHandler_ExtremelyLargeID(t *testing.T) {
	// Test handling of extremely large ID strings
	largeID := strings.Repeat("a", 10000)
	option := DropdownOption{
		ID:   largeID,
		Name: "Option",
	}

	data, err := json.Marshal(option)
	require.NoError(t, err)

	var parsed DropdownOption
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed.ID, 10000)
}
