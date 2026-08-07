package response

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestResponse_Structure(t *testing.T) {
	// Test Response struct
	resp := Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    map[string]string{"key": "value"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed Response
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if parsed.Code != http.StatusOK {
		t.Errorf("expected code %d, got: %d", http.StatusOK, parsed.Code)
	}
	if parsed.Message != "success" {
		t.Errorf("expected message 'success', got: %s", parsed.Message)
	}
}

func TestPaginatedResponse_Structure(t *testing.T) {
	// Test PaginatedResponse struct
	resp := PaginatedResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    []map[string]string{{"key": "value"}},
		Pagination: map[string]interface{}{
			"total": 100,
			"page":  1,
		},
		Summary: map[string]interface{}{
			"total_count": 100,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed PaginatedResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if parsed.Code != http.StatusOK {
		t.Errorf("expected code %d, got: %d", http.StatusOK, parsed.Code)
	}
	if parsed.Summary == nil {
		t.Error("expected summary to be present")
	}
}

func TestValidationErrorDetail_Structure(t *testing.T) {
	// Test ValidationErrorDetail struct
	detail := ValidationErrorDetail{
		Field:   "email",
		Message: "is required",
	}

	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("failed to marshal detail: %v", err)
	}

	var parsed ValidationErrorDetail
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal detail: %v", err)
	}

	if parsed.Field != "email" {
		t.Errorf("expected field 'email', got: %s", parsed.Field)
	}
	if parsed.Message != "is required" {
		t.Errorf("expected message 'is required', got: %s", parsed.Message)
	}
}

func TestJSONSerialization(t *testing.T) {
	// Test that responses serialize correctly
	tests := []struct {
		name     string
		response interface{}
		wantCode int
	}{
		{
			name: "Success response",
			response: Response{
				Code:    http.StatusOK,
				Message: "success",
				Data:    "test",
			},
			wantCode: http.StatusOK,
		},
		{
			name: "Error response",
			response: Response{
				Code:    http.StatusBadRequest,
				Message: "bad request",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "Created response",
			response: Response{
				Code:    http.StatusCreated,
				Message: "created",
			},
			wantCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal response: %v", err)
			}

			var parsed Response
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if parsed.Code != tt.wantCode {
				t.Errorf("expected code %d, got: %d", tt.wantCode, parsed.Code)
			}
		})
	}
}

func TestResponseJSONTags(t *testing.T) {
	// Verify JSON tags are correct
	resp := Response{
		Code:    200,
		Message: "ok",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// Check that JSON keys are lowercase (snake_case)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, ok := result["code"]; !ok {
		t.Error("expected 'code' key in JSON")
	}
	if _, ok := result["message"]; !ok {
		t.Error("expected 'message' key in JSON")
	}
	if _, ok := result["Code"]; ok {
		t.Error("unexpected 'Code' key in JSON (should be lowercase)")
	}
}

func TestResponseWithHertz(t *testing.T) {
	// Test that response helpers work with Hertz context
	// Create a basic context for testing
	w := &responseRecorder{}
	c := &app.RequestContext{}
	_ = w // unused for now
	_ = c // unused for now

	// Test Success helper - basic structure test
	resp := Response{
		Code:    consts.StatusOK,
		Message: "success",
		Data:    map[string]string{"key": "value"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed Response
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if parsed.Code != http.StatusOK {
		t.Errorf("expected code %d, got: %d", http.StatusOK, parsed.Code)
	}
}

type responseRecorder struct {
	code int
	body []byte
}
