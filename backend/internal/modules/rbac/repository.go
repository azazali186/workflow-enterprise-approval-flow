package rbac

import (
	"context"
	"time"

	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles all RBAC database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new RBAC repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ==================== User Operations ====================

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Preload("Roles").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (r *Repository) UpdateUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser soft deletes a user
func (r *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// ListUsers returns a paginated list of users
func (r *Repository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Preload("Roles").Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error
	return users, err
}

// GetUserWithRoles retrieves a user with their roles and permissions
func (r *Repository) GetUserWithRolesAndPermissions(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Permissions")
		}).
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// AssignRoleToUser assigns a role to a user
func (r *Repository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	userRole := domain.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(&userRole).Error
}

// RemoveRoleFromUser removes a role from a user
func (r *Repository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&domain.UserRole{}).Error
}

// GetUserRoles retrieves all roles for a user
func (r *Repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

// ==================== Role Operations ====================

// CreateRole creates a new role
func (r *Repository) CreateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// GetRoleByID retrieves a role by ID
func (r *Repository) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName retrieves a role by name
func (r *Repository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// UpdateRole updates a role
func (r *Repository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// DeleteRole soft deletes a role
func (r *Repository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Role{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// ListRoles returns all roles
func (r *Repository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Order("created_at DESC").Find(&roles).Error
	return roles, err
}

// AssignPermissionToRole assigns a permission to a role
func (r *Repository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	rolePermission := domain.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(&rolePermission).Error
}

// RemovePermissionFromRole removes a permission from a role
func (r *Repository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&domain.RolePermission{}).Error
}

// GetRolePermissions retrieves all permissions for a role
func (r *Repository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

// ==================== Permission Operations ====================

// CreatePermission creates a new permission
func (r *Repository) CreatePermission(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetPermissionByID retrieves a permission by ID
func (r *Repository) GetPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByRoute retrieves a permission by route
func (r *Repository) GetPermissionByRoute(ctx context.Context, route string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).Where("route = ?", route).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// UpdatePermission updates a permission
func (r *Repository) UpdatePermission(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// DeletePermission soft deletes a permission
func (r *Repository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Permission{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// ListPermissions returns all permissions
func (r *Repository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&permissions).Error
	return permissions, err
}

// UpsertPermission creates or updates a permission based on route
func (r *Repository) UpsertPermission(ctx context.Context, permission *domain.Permission) error {
	var existing domain.Permission
	err := r.db.WithContext(ctx).Where("route = ?", permission.Route).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(permission).Error
	}
	if err != nil {
		return err
	}
	// Update existing permission
	existing.Name = permission.Name
	existing.Path = permission.Path
	existing.Method = permission.Method
	existing.Service = permission.Service
	return r.db.WithContext(ctx).Save(&existing).Error
}

// BulkUpsertPermissions creates or updates multiple permissions
func (r *Repository) BulkUpsertPermissions(ctx context.Context, permissions []domain.Permission) error {
	for i := range permissions {
		if err := r.UpsertPermission(ctx, &permissions[i]); err != nil {
			return err
		}
	}
	return nil
}
