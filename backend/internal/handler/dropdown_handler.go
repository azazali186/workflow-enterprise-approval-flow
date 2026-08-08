package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

// DropdownOption represents a simple {id, name} pair for dropdown menus.
type DropdownOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// validEntities is the set of entity types the API supports.
var validEntities = map[string]bool{
	"users":        true,
	"workflows":    true,
	"templates":    true,
	"roles":        true,
	"applications": true,
	"approvals":    true,
}

// DropdownHandler handles dropdown option requests.
type DropdownHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewDropdownHandler creates a new DropdownHandler.
func NewDropdownHandler(db *gorm.DB, cfg *config.Config) *DropdownHandler {
	return &DropdownHandler{db: db, cfg: cfg}
}

// ListDropdowns returns dropdown options for the requested entity types.
// @Summary      List dropdown options
// @Description  Returns id/name pairs for dropdown menus (users, workflows, templates, roles, etc.)
// @Tags         Dropdowns
// @Accept       json
// @Produce      json
// @Param        request body validation.DropdownListRequest true "Entity types to fetch"
// @Success      200 {object} map[string][]DropdownOption
// @Failure      400 {object} response.ErrorResponse
// @Router       /api/v1/dropdowns [post]
func (h *DropdownHandler) ListDropdowns(ctx context.Context, c *app.RequestContext) {
	var req validation.DropdownListRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate entity types and collect invalid ones.
	var invalidEntities []string
	for _, entity := range req.Entities {
		if !validEntities[entity] {
			invalidEntities = append(invalidEntities, entity)
		}
	}
	if len(invalidEntities) > 0 {
		response.Error(c, http.StatusBadRequest,
			fmt.Sprintf("invalid entity type(s): %v. Valid types: users, workflows, templates, roles, applications, approvals", invalidEntities))
		return
	}

	result := make(map[string][]DropdownOption, len(req.Entities))

	for _, entity := range req.Entities {
		var (
			options []DropdownOption
			err     error
		)

		switch entity {
		case "users":
			options, err = h.listUsers(ctx)
		case "workflows":
			options, err = h.listWorkflows(ctx, req.IncludeInactive)
		case "templates":
			options, err = h.listTemplates(ctx)
		case "roles":
			options, err = h.listRoles(ctx)
		case "applications":
			statuses := req.Statuses
			if len(statuses) == 0 {
				statuses = []string{"submitted"}
			}
			options, err = h.listApplications(ctx, statuses)
		case "approvals":
			options, err = h.listApprovals(ctx)
		}

		if err != nil {
			h.cfg.Error("failed to list dropdown options",
				zap.String("entity", entity),
				zap.Error(err),
			)
			response.Error(c, http.StatusInternalServerError,
				fmt.Sprintf("failed to fetch %s options", entity))
			return
		}

		result[entity] = options
	}

	response.Success(c, result)
}

func (h *DropdownHandler) listUsers(ctx context.Context) ([]DropdownOption, error) {
	type row struct {
		ID   string
		Name string
	}
	var rows []row
	err := h.db.WithContext(ctx).
		Table("users").
		Select("id, name || ' (' || email || ')' as name").
		Where("deleted_at IS NULL AND status = ?", "active").
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	opts := make([]DropdownOption, len(rows))
	for i, r := range rows {
		opts[i] = DropdownOption{ID: r.ID, Name: r.Name}
	}
	return opts, nil
}

func (h *DropdownHandler) listWorkflows(ctx context.Context, includeInactive bool) ([]DropdownOption, error) {
	type row struct {
		ID   string
		Name string
	}
	var rows []row
	q := h.db.WithContext(ctx).
		Table("workflows").
		Select("id, name").
		Where("deleted_at IS NULL")
	if !includeInactive {
		q = q.Where("is_active = ?", true)
	}
	err := q.Order("name ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	opts := make([]DropdownOption, len(rows))
	for i, r := range rows {
		opts[i] = DropdownOption{ID: r.ID, Name: r.Name}
	}
	return opts, nil
}

func (h *DropdownHandler) listTemplates(ctx context.Context) ([]DropdownOption, error) {
	type row struct {
		ID   string
		Name string
	}
	var rows []row
	err := h.db.WithContext(ctx).
		Table("templates").
		Select("id, name").
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	opts := make([]DropdownOption, len(rows))
	for i, r := range rows {
		opts[i] = DropdownOption{ID: r.ID, Name: r.Name}
	}
	return opts, nil
}

func (h *DropdownHandler) listRoles(ctx context.Context) ([]DropdownOption, error) {
	type row struct {
		ID   string
		Name string
	}
	var rows []row
	err := h.db.WithContext(ctx).
		Table("roles").
		Select("id, name").
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	opts := make([]DropdownOption, len(rows))
	for i, r := range rows {
		opts[i] = DropdownOption{ID: r.ID, Name: r.Name}
	}
	return opts, nil
}

func (h *DropdownHandler) listApplications(ctx context.Context, statuses []string) ([]DropdownOption, error) {
	type row struct {
		ID    string
		Title string
	}
	var rows []row
	err := h.db.WithContext(ctx).
		Table("applications").
		Select("id, COALESCE(title, 'Untitled') as title").
		Where("deleted_at IS NULL").
		Where("status IN ?", statuses).
		Order("created_at DESC").
		Limit(100).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	opts := make([]DropdownOption, len(rows))
	for i, r := range rows {
		opts[i] = DropdownOption{ID: r.ID, Name: r.Title}
	}
	return opts, nil
}

func (h *DropdownHandler) listApprovals(ctx context.Context) ([]DropdownOption, error) {
	type row struct {
		ID     string
		Status string
	}
	var rows []row
	err := h.db.WithContext(ctx).
		Table("approvals").
		Select("id, status").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(100).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	opts := make([]DropdownOption, len(rows))
	for i, r := range rows {
		opts[i] = DropdownOption{ID: r.ID, Name: r.Status}
	}
	return opts, nil
}
