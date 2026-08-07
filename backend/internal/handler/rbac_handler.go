package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/pkg/pagination"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type RBACHandler struct {
	svc *rbac.Service
	cfg *config.Config
}

func NewRBACHandler(svc *rbac.Service, cfg *config.Config) *RBACHandler {
	return &RBACHandler{svc: svc, cfg: cfg}
}

// ==================== User Endpoints ====================

// ListUsers godoc
// @Summary      List users
// @Description  Get a paginated list of all users
// @Tags         RBAC - Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.ListUsersRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/users [post]
func (h *RBACHandler) ListUsers(ctx context.Context, c *app.RequestContext) {
	var req validation.ListUsersRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	users, err := h.svc.ListUsers(ctx, req.Limit, 0)
	if err != nil {
		h.cfg.Error("failed to list users", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	_ = pagination.PaginationResponse{
		TotalCount: int64(len(users)),
		PageSize:   req.Limit,
	}

	response.Success(c, users)
}

// GetUser godoc
// @Summary      Get user
// @Description  Get a specific user by ID
// @Tags         RBAC - Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.GetUserRequest true  "User ID"
// @Success      200  {object}  response.Response{data=domain.User}
// @Failure      400  {object}  response.Response
// @Router       /admin/users/get [post]
func (h *RBACHandler) GetUser(ctx context.Context, c *app.RequestContext) {
	var req validation.GetUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.svc.GetUser(ctx, id)
	if err != nil {
		h.cfg.Error("failed to get user", zap.Error(err))
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.Success(c, user)
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Update a user's information
// @Tags         RBAC - Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.UpdateUserRequest true  "Update details"
// @Success      200  {object}  response.Response{data=domain.User}
// @Failure      400  {object}  response.Response
// @Router       /admin/users/update [post]
func (h *RBACHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	user := &domain.User{
		Email: req.Email,
		Name:  req.Name,
	}
	user.ID = id

	if err := h.svc.UpdateUser(ctx, user); err != nil {
		h.cfg.Error("failed to update user", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	response.Success(c, user)
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Delete a user
// @Tags         RBAC - Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.DeleteUserRequest true  "User ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/users/delete [post]
func (h *RBACHandler) DeleteUser(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.svc.DeleteUser(ctx, id); err != nil {
		h.cfg.Error("failed to delete user", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to delete user")
		return
	}

	response.Success(c, map[string]string{"message": "user deleted"})
}

// ==================== Role Endpoints ====================

// ListRoles godoc
// @Summary      List roles
// @Description  Get a list of all roles
// @Tags         RBAC - Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.ListRolesRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/roles [post]
func (h *RBACHandler) ListRoles(ctx context.Context, c *app.RequestContext) {
	var req validation.ListRolesRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	roles, err := h.svc.ListRoles(ctx)
	if err != nil {
		h.cfg.Error("failed to list roles", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to list roles")
		return
	}

	response.Success(c, roles)
}

// GetRole godoc
// @Summary      Get role
// @Description  Get a specific role by ID
// @Tags         RBAC - Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.GetRoleRequest true  "Role ID"
// @Success      200  {object}  response.Response{data=domain.Role}
// @Failure      400  {object}  response.Response
// @Router       /admin/roles/get [post]
func (h *RBACHandler) GetRole(ctx context.Context, c *app.RequestContext) {
	var req validation.GetRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	role, err := h.svc.GetRole(ctx, id)
	if err != nil {
		h.cfg.Error("failed to get role", zap.Error(err))
		response.Error(c, http.StatusNotFound, "role not found")
		return
	}

	response.Success(c, role)
}

// CreateRole godoc
// @Summary      Create role
// @Description  Create a new role
// @Tags         RBAC - Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.CreateRoleRequest true  "Role details"
// @Success      201  {object}  response.Response{data=domain.Role}
// @Failure      400  {object}  response.Response
// @Router       /admin/roles/create [post]
func (h *RBACHandler) CreateRole(ctx context.Context, c *app.RequestContext) {
	var req validation.CreateRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   req.IsDefault,
	}

	if err := h.svc.CreateRole(ctx, role); err != nil {
		h.cfg.Error("failed to create role", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	response.Success(c, role)
}

// UpdateRole godoc
// @Summary      Update role
// @Description  Update a role
// @Tags         RBAC - Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.UpdateRoleRequest true  "Update details"
// @Success      200  {object}  response.Response{data=domain.Role}
// @Failure      400  {object}  response.Response
// @Router       /admin/roles/update [post]
func (h *RBACHandler) UpdateRole(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdateRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	role := &domain.Role{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.IsDefault != nil {
		role.IsDefault = *req.IsDefault
	}
	role.ID = id

	if err := h.svc.UpdateRole(ctx, role); err != nil {
		h.cfg.Error("failed to update role", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to update role")
		return
	}

	response.Success(c, role)
}

// DeleteRole godoc
// @Summary      Delete role
// @Description  Delete a role
// @Tags         RBAC - Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.DeleteRoleRequest true  "Role ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/roles/delete [post]
func (h *RBACHandler) DeleteRole(ctx context.Context, c *app.RequestContext) {
	var req validation.DeleteRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	if err := h.svc.DeleteRole(ctx, id); err != nil {
		h.cfg.Error("failed to delete role", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to delete role")
		return
	}

	response.Success(c, map[string]string{"message": "role deleted"})
}

// ==================== Permission Endpoints ====================

// ListPermissions godoc
// @Summary      List permissions
// @Description  Get a list of all permissions
// @Tags         RBAC - Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.ListPermissionsRequest true  "List parameters"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/permissions [post]
func (h *RBACHandler) ListPermissions(ctx context.Context, c *app.RequestContext) {
	var req validation.ListPermissionsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	permissions, err := h.svc.ListPermissions(ctx)
	if err != nil {
		h.cfg.Error("failed to list permissions", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to list permissions")
		return
	}

	response.Success(c, permissions)
}

// GetPermission godoc
// @Summary      Get permission
// @Description  Get a specific permission by ID
// @Tags         RBAC - Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.GetPermissionRequest true  "Permission ID"
// @Success      200  {object}  response.Response{data=domain.Permission}
// @Failure      400  {object}  response.Response
// @Router       /admin/permissions/get [post]
func (h *RBACHandler) GetPermission(ctx context.Context, c *app.RequestContext) {
	var req validation.GetPermissionRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.PermissionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID")
		return
	}

	permission, err := h.svc.GetPermission(ctx, id)
	if err != nil {
		h.cfg.Error("failed to get permission", zap.Error(err))
		response.Error(c, http.StatusNotFound, "permission not found")
		return
	}

	response.Success(c, permission)
}

// CreatePermission godoc
// @Summary      Create permission
// @Description  Create a new permission
// @Tags         RBAC - Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.CreatePermissionRequest true  "Permission details"
// @Success      201  {object}  response.Response{data=domain.Permission}
// @Failure      400  {object}  response.Response
// @Router       /admin/permissions/create [post]
func (h *RBACHandler) CreatePermission(ctx context.Context, c *app.RequestContext) {
	var req validation.CreatePermissionRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	permission := &domain.Permission{
		Name:    req.Name,
		Route:   req.Route,
		Path:    req.Path,
		Method:  req.Method,
		Service: req.Service,
	}

	if err := h.svc.CreatePermission(ctx, permission); err != nil {
		h.cfg.Error("failed to create permission", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	response.Success(c, permission)
}

// UpdatePermission godoc
// @Summary      Update permission
// @Description  Update a permission
// @Tags         RBAC - Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.UpdatePermissionRequest true  "Update details"
// @Success      200  {object}  response.Response{data=domain.Permission}
// @Failure      400  {object}  response.Response
// @Router       /admin/permissions/update [post]
func (h *RBACHandler) UpdatePermission(ctx context.Context, c *app.RequestContext) {
	var req validation.UpdatePermissionRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.PermissionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID")
		return
	}

	permission := &domain.Permission{
		Name:    req.Name,
		Route:   req.Route,
		Path:    req.Path,
		Method:  req.Method,
		Service: req.Service,
	}
	permission.ID = id

	if err := h.svc.UpdatePermission(ctx, permission); err != nil {
		h.cfg.Error("failed to update permission", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to update permission")
		return
	}

	response.Success(c, permission)
}

// DeletePermission godoc
// @Summary      Delete permission
// @Description  Delete a permission
// @Tags         RBAC - Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.DeletePermissionRequest true  "Permission ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/permissions/delete [post]
func (h *RBACHandler) DeletePermission(ctx context.Context, c *app.RequestContext) {
	var req validation.DeletePermissionRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.PermissionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID")
		return
	}

	if err := h.svc.DeletePermission(ctx, id); err != nil {
		h.cfg.Error("failed to delete permission", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to delete permission")
		return
	}

	response.Success(c, map[string]string{"message": "permission deleted"})
}

// ==================== Role-Permission Endpoints ====================

// GetRolePermissions godoc
// @Summary      Get role permissions
// @Description  Get all permissions for a role
// @Tags         RBAC - Role Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.GetRolePermissionsRequest true  "Role ID"
// @Success      200  {object}  response.Response{data=[]domain.Permission}
// @Failure      400  {object}  response.Response
// @Router       /admin/roles/permissions [post]
func (h *RBACHandler) GetRolePermissions(ctx context.Context, c *app.RequestContext) {
	var req validation.GetRolePermissionsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	permissions, err := h.svc.GetRolePermissions(ctx, id)
	if err != nil {
		h.cfg.Error("failed to get role permissions", zap.Error(err))
		response.Error(c, http.StatusNotFound, "role not found")
		return
	}

	response.Success(c, permissions)
}

// AssignPermissionToRole godoc
// @Summary      Assign permission to role
// @Description  Assign a permission to a role
// @Tags         RBAC - Role Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.AssignPermissionToRoleRequest true  "Assignment details"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/roles/permissions/assign [post]
func (h *RBACHandler) AssignPermissionToRole(ctx context.Context, c *app.RequestContext) {
	var req validation.AssignPermissionToRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	permissionID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID")
		return
	}

	if err := h.svc.AssignPermissionToRole(ctx, roleID, permissionID); err != nil {
		h.cfg.Error("failed to assign permission", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "permission assigned"})
}

// RemovePermissionFromRole godoc
// @Summary      Remove permission from role
// @Description  Remove a permission from a role
// @Tags         RBAC - Role Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.RemovePermissionFromRoleRequest true  "Removal details"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/roles/permissions/remove [post]
func (h *RBACHandler) RemovePermissionFromRole(ctx context.Context, c *app.RequestContext) {
	var req validation.RemovePermissionFromRoleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	permissionID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID")
		return
	}

	if err := h.svc.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
		h.cfg.Error("failed to remove permission", zap.Error(err))
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "permission removed"})
}

// ==================== User-Role Endpoints ====================

// GetUserRoles godoc
// @Summary      Get user roles
// @Description  Get all roles for a user
// @Tags         RBAC - User Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.GetUserRolesRequest true  "User ID"
// @Success      200  {object}  response.Response{data=[]domain.Role}
// @Failure      400  {object}  response.Response
// @Router       /admin/users/roles [post]
func (h *RBACHandler) GetUserRoles(ctx context.Context, c *app.RequestContext) {
	var req validation.GetUserRolesRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	id, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	roles, err := h.svc.GetUserRoles(ctx, id)
	if err != nil {
		h.cfg.Error("failed to get user roles", zap.Error(err))
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.Success(c, roles)
}

// AssignRoleToUser godoc
// @Summary      Assign role to user
// @Description  Assign a role to a user
// @Tags         RBAC - User Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.AssignRoleToUserRequest true  "Assignment details"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/users/roles/assign [post]
func (h *RBACHandler) AssignRoleToUser(ctx context.Context, c *app.RequestContext) {
	var req validation.AssignRoleToUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	if err := h.svc.AssignRole(ctx, userID, roleID); err != nil {
		h.cfg.Error("failed to assign role", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "role assigned"})
}

// RemoveRoleFromUser godoc
// @Summary      Remove role from user
// @Description  Remove a role from a user
// @Tags         RBAC - User Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     validation.RemoveRoleFromUserRequest true  "Removal details"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /admin/users/roles/remove [post]
func (h *RBACHandler) RemoveRoleFromUser(ctx context.Context, c *app.RequestContext) {
	var req validation.RemoveRoleFromUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID")
		return
	}

	if err := h.svc.RemoveRole(ctx, userID, roleID); err != nil {
		h.cfg.Error("failed to remove role", zap.Error(err))
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "role removed"})
}

// Helper to parse pagination from request body
func parsePaginationFromBody(c *app.RequestContext) (cursor string, limit int, sortBy, sortOrder string) {
	cursor = c.Query("cursor")
	limitStr := c.Query("limit")
	limit = 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}
	sortBy = c.Query("sort_by")
	sortOrder = c.Query("sort_order")
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}
	return
}
