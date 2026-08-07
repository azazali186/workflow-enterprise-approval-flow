package pagination

import (
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// ParseQuery parses pagination parameters from request query
func ParseQuery(c *app.RequestContext) *PaginationRequest {
	cursor := c.Query("cursor")
	
	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}

	return &PaginationRequest{
		Cursor: cursor,
		Limit:  limit,
	}
}

// ParseSortQuery parses sorting parameters from request query
func ParseSortQuery(c *app.RequestContext) *SortRequest {
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")
	
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	return &SortRequest{
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}

// ParseFilterQuery parses filter parameters from request query
func ParseFilterQuery(c *app.RequestContext) map[string]string {
	filters := make(map[string]string)
	
	// Common filter parameters
	filterKeys := []string{
		"status", "priority", "category", "type", "channel",
		"search", "workflow_id", "template_id", "applicant_id",
		"approver_id", "application_id", "decision", "is_active",
	}
	
	for _, key := range filterKeys {
		if value := c.Query(key); value != "" {
			filters[key] = value
		}
	}
	
	return filters
}

// ParseDateRangeQuery parses date range parameters from request query
func ParseDateRangeQuery(c *app.RequestContext) *DateRangeFilter {
	var startDate, endDate *time.Time
	
	if start := c.Query("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = &t
		}
	}
	
	return &DateRangeFilter{
		Start: startDate,
		End:   endDate,
	}
}

// SortRequest represents sorting parameters
type SortRequest struct {
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// Validate validates the sort request
func (s *SortRequest) Validate() {
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"name":       true,
		"status":     true,
		"priority":   true,
		"title":      true,
	}

	if !validSortFields[s.SortBy] {
		s.SortBy = "created_at"
	}

	if s.SortOrder != "asc" && s.SortOrder != "desc" {
		s.SortOrder = "desc"
	}
}
