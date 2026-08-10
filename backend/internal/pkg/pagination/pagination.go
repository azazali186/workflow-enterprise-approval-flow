package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Cursor represents a cursor for pagination
type Cursor struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// EncodeCursor encodes a cursor to a base64 string
func EncodeCursor(cursor *Cursor) string {
	if cursor == nil {
		return ""
	}
	data, _ := json.Marshal(cursor)
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeCursor decodes a base64 string to a cursor
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format")
	}
	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor data")
	}
	return &cursor, nil
}

// PaginationRequest represents common pagination parameters
type PaginationRequest struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

// Validate validates the pagination request
func (p *PaginationRequest) Validate() {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

// PaginationResponse represents the pagination metadata
type PaginationResponse struct {
	NextCursor     string `json:"next_cursor,omitempty"`
	PreviousCursor string `json:"previous_cursor,omitempty"`
	HasMore        bool   `json:"has_more"`
	HasPrevious    bool   `json:"has_previous"`
	TotalCount     int64  `json:"total_count"`
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	TotalPages     int    `json:"total_pages"`
}

// ListResponse represents a generic list response with pagination
type ListResponse struct {
	Data       interface{}         `json:"data"`
	Pagination *PaginationResponse `json:"pagination"`
	Summary    interface{}         `json:"summary,omitempty"`
}

// NewPaginationResponse creates a new pagination response
func NewPaginationResponse(totalCount int64, page, pageSize int) *PaginationResponse {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &PaginationResponse{
		TotalCount:  totalCount,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasMore:     page < totalPages,
		HasPrevious: page > 1,
	}
}

// NewCursorPaginationResponse creates a cursor-based pagination response
func NewCursorPaginationResponse(totalCount int64, pageSize int, hasMore bool, nextCursor, prevCursor string) *PaginationResponse {
	return &PaginationResponse{
		NextCursor:     nextCursor,
		PreviousCursor: prevCursor,
		HasMore:        hasMore,
		HasPrevious:    prevCursor != "",
		TotalCount:     totalCount,
		PageSize:       pageSize,
	}
}
