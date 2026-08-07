package pagination

import (
	"testing"
	"time"
)

func TestEncodeCursor(t *testing.T) {
	// Test nil cursor
	result := EncodeCursor(nil)
	if result != "" {
		t.Errorf("expected empty string for nil cursor, got: %s", result)
	}

	// Test valid cursor
	cursor := &Cursor{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	result = EncodeCursor(cursor)
	if result == "" {
		t.Error("expected non-empty string for valid cursor")
	}

	// Test round trip
	decoded, err := DecodeCursor(result)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if decoded.ID != cursor.ID {
		t.Errorf("expected %s, got: %s", cursor.ID, decoded.ID)
	}
}

func TestDecodeCursor(t *testing.T) {
	// Test empty string
	cursor, err := DecodeCursor("")
	if err != nil {
		t.Errorf("expected no error for empty string, got: %v", err)
	}
	if cursor != nil {
		t.Errorf("expected nil cursor for empty string, got: %v", cursor)
	}

	// Test invalid base64
	_, err = DecodeCursor("invalid-base64")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	// Test invalid JSON
	_, err = DecodeCursor("aW52YWxpZA==") // "invalid" in base64
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestPaginationRequest_Validate(t *testing.T) {
	// Test defaults
	req := &PaginationRequest{}
	req.Validate()
	if req.Limit != 10 {
		t.Errorf("expected default limit 10, got: %d", req.Limit)
	}

	// Test max limit
	req = &PaginationRequest{Limit: 200}
	req.Validate()
	if req.Limit != 100 {
		t.Errorf("expected max limit 100, got: %d", req.Limit)
	}

	// Test valid limit
	req = &PaginationRequest{Limit: 50}
	req.Validate()
	if req.Limit != 50 {
		t.Errorf("expected limit 50, got: %d", req.Limit)
	}

	// Test negative limit
	req = &PaginationRequest{Limit: -1}
	req.Validate()
	if req.Limit != 10 {
		t.Errorf("expected default limit 10 for negative, got: %d", req.Limit)
	}
}

func TestNewPaginationResponse(t *testing.T) {
	// Test basic response
	resp := NewPaginationResponse(100, 1, 10)
	if resp.TotalCount != 100 {
		t.Errorf("expected total count 100, got: %d", resp.TotalCount)
	}
	if resp.Page != 1 {
		t.Errorf("expected page 1, got: %d", resp.Page)
	}
	if resp.PageSize != 10 {
		t.Errorf("expected page size 10, got: %d", resp.PageSize)
	}
	if resp.TotalPages != 10 {
		t.Errorf("expected total pages 10, got: %d", resp.TotalPages)
	}
	if !resp.HasMore {
		t.Error("expected has more to be true")
	}
	if resp.HasPrevious {
		t.Error("expected has previous to be false")
	}

	// Test with remainder
	resp = NewPaginationResponse(105, 1, 10)
	if resp.TotalPages != 11 {
		t.Errorf("expected total pages 11, got: %d", resp.TotalPages)
	}

	// Test last page
	resp = NewPaginationResponse(100, 10, 10)
	if resp.HasMore {
		t.Error("expected has more to be false on last page")
	}

	// Test first page
	resp = NewPaginationResponse(100, 1, 10)
	if resp.HasPrevious {
		t.Error("expected has previous to be false on first page")
	}
}

func TestNewCursorPaginationResponse(t *testing.T) {
	// Test basic response
	resp := NewCursorPaginationResponse(100, 10, true, "next-cursor", "prev-cursor")
	if resp.TotalCount != 100 {
		t.Errorf("expected total count 100, got: %d", resp.TotalCount)
	}
	if resp.PageSize != 10 {
		t.Errorf("expected page size 10, got: %d", resp.PageSize)
	}
	if !resp.HasMore {
		t.Error("expected has more to be true")
	}
	if !resp.HasPrevious {
		t.Error("expected has previous to be true")
	}
	if resp.NextCursor != "next-cursor" {
		t.Errorf("expected next cursor 'next-cursor', got: %s", resp.NextCursor)
	}
	if resp.PreviousCursor != "prev-cursor" {
		t.Errorf("expected previous cursor 'prev-cursor', got: %s", resp.PreviousCursor)
	}

	// Test no more
	resp = NewCursorPaginationResponse(100, 10, false, "", "")
	if resp.HasMore {
		t.Error("expected has more to be false")
	}
	if resp.HasPrevious {
		t.Error("expected has previous to be false")
	}
}
