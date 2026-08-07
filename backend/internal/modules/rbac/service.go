package rbac

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/pkg/auth"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// TokenCachePrefix is the prefix for token cache keys
	TokenCachePrefix = "auth:token:"
	// DefaultTokenTTL is the default token TTL
	DefaultTokenTTL = 24 * time.Hour
	// RefreshThreshold is the threshold for token refresh
	RefreshThreshold = 30 * time.Minute
)

// Service handles RBAC business logic
type Service struct {
	Repo    *Repository
	Cache   *cache.Redis
	Token   *auth.TokenService
	Logger  *config.Config
}

// NewService creates a new RBAC service
func NewService(repo *Repository, cache *cache.Redis, token *auth.TokenService, cfg *config.Config) *Service {
	return &Service{
		Repo:   repo,
		Cache:  cache,
		Token:  token,
		Logger: cfg,
	}
}

// ==================== Auth Operations ====================

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.Repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	accessToken, err := s.Token.Generate(user.ID.String(), user.Email, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.Token.GenerateRefresh(user.ID.String(), user.Email, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store token in Redis for single sign-on
	tokenHash := computeTokenHash(accessToken, user.ID.String())
	tokenKey := TokenCachePrefix + user.ID.String()
	if err := s.Cache.Set(ctx, tokenKey, tokenHash, DefaultTokenTTL); err != nil {
		s.Logger.Error("failed to cache token", zap.Error(err))
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.Repo.UpdateUser(ctx, user); err != nil {
		s.Logger.Error("failed to update last login", zap.Error(err))
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(DefaultTokenTTL),
	}, nil
}

// Register creates a new user
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*domain.User, error) {
	// Check if user already exists
	existing, _ := s.Repo.GetUserByEmail(ctx, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Create new user
	user := &domain.User{
		Email:  req.Email,
		Name:   req.Name,
		Status: "active",
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Assign default role if exists
	defaultRole, err := s.Repo.GetRoleByName(ctx, "user")
	if err == nil {
		user.Roles = []domain.Role{*defaultRole}
	}

	if err := s.Repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// RefreshToken refreshes an expired token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Parse the refresh token
	claims, err := s.Token.Validate(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	user, err := s.Repo.GetUserWithRolesAndPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new tokens
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	accessToken, err := s.Token.Generate(user.ID.String(), user.Email, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.Token.GenerateRefresh(user.ID.String(), user.Email, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update token cache
	tokenHash := computeTokenHash(accessToken, user.ID.String())
	tokenKey := TokenCachePrefix + user.ID.String()
	if err := s.Cache.Set(ctx, tokenKey, tokenHash, DefaultTokenTTL); err != nil {
		s.Logger.Error("failed to cache token", zap.Error(err))
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(DefaultTokenTTL),
	}, nil
}

// Logout invalidates a user's token
func (s *Service) Logout(ctx context.Context, userID string) error {
	tokenKey := TokenCachePrefix + userID
	return s.Cache.Delete(ctx, tokenKey)
}

// ValidateToken validates a token and checks if it's still valid in Redis
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	// Parse token
	claims, err := s.Token.Validate(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if token is in Redis (single sign-on)
	tokenHash := computeTokenHash(tokenString, claims.UserID)
	tokenKey := TokenCachePrefix + claims.UserID
	cachedHash, err := s.Cache.Get(ctx, tokenKey)
	if err != nil || cachedHash == "" {
		return nil, fmt.Errorf("token expired or invalidated")
	}

	// Verify token hash matches
	if cachedHash != tokenHash {
		return nil, fmt.Errorf("token invalidated (single sign-on)")
	}

	return claims, nil
}

// RenewToken renews the token cache TTL if needed
func (s *Service) RenewToken(ctx context.Context, tokenString string) error {
	claims, err := s.Token.Validate(tokenString)
	if err != nil {
		return err
	}

	tokenHash := computeTokenHash(tokenString, claims.UserID)
	tokenKey := TokenCachePrefix + claims.UserID

	// Get current TTL
	_, ttl, err := s.Cache.GetWithTTL(ctx, tokenKey)
	if err != nil {
		return err
	}

	// Renew if less than threshold remaining
	if ttl < RefreshThreshold {
		if err := s.Cache.Set(ctx, tokenKey, tokenHash, DefaultTokenTTL); err != nil {
			return err
		}
	}

	return nil
}

// ==================== User Operations ====================

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.Repo.GetUserWithRolesAndPermissions(ctx, id)
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.Repo.UpdateUser(ctx, user)
}

// DeleteUser soft deletes a user
func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.Repo.DeleteUser(ctx, id)
}

// ListUsers returns a paginated list of users
func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	return s.Repo.ListUsers(ctx, limit, offset)
}

// AssignRole assigns a role to a user
func (s *Service) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.Repo.AssignRoleToUser(ctx, userID, roleID)
}

// RemoveRole removes a role from a user
func (s *Service) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.Repo.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetUserRoles retrieves all roles for a user
func (s *Service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	return s.Repo.GetUserRoles(ctx, userID)
}

// ==================== Role Operations ====================

// CreateRole creates a new role
func (s *Service) CreateRole(ctx context.Context, role *domain.Role) error {
	return s.Repo.CreateRole(ctx, role)
}

// GetRole retrieves a role by ID
func (s *Service) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return s.Repo.GetRoleByID(ctx, id)
}

// UpdateRole updates a role
func (s *Service) UpdateRole(ctx context.Context, role *domain.Role) error {
	return s.Repo.UpdateRole(ctx, role)
}

// DeleteRole soft deletes a role
func (s *Service) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.Repo.DeleteRole(ctx, id)
}

// ListRoles returns all roles
func (s *Service) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.Repo.ListRoles(ctx)
}

// AssignPermissionToRole assigns a permission to a role
func (s *Service) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.Repo.AssignPermissionToRole(ctx, roleID, permissionID)
}

// RemovePermissionFromRole removes a permission from a role
func (s *Service) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.Repo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

// GetRolePermissions retrieves all permissions for a role
func (s *Service) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	return s.Repo.GetRolePermissions(ctx, roleID)
}

// ==================== Permission Operations ====================

// CreatePermission creates a new permission
func (s *Service) CreatePermission(ctx context.Context, permission *domain.Permission) error {
	return s.Repo.CreatePermission(ctx, permission)
}

// GetPermission retrieves a permission by ID
func (s *Service) GetPermission(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	return s.Repo.GetPermissionByID(ctx, id)
}

// UpdatePermission updates a permission
func (s *Service) UpdatePermission(ctx context.Context, permission *domain.Permission) error {
	return s.Repo.UpdatePermission(ctx, permission)
}

// DeletePermission soft deletes a permission
func (s *Service) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return s.Repo.DeletePermission(ctx, id)
}

// ListPermissions returns all permissions
func (s *Service) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.Repo.ListPermissions(ctx)
}

// ==================== Route Permission Operations ====================

// RouteInfo represents a route's permission info
type RouteInfo struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Guard  string `json:"guard"`
	Method string `json:"method"`
}

// StoreRoutePermissions stores route permissions in database and Redis
func (s *Service) StoreRoutePermissions(ctx context.Context, routes []RouteInfo) error {
	permissions := make([]domain.Permission, len(routes))
	for i, route := range routes {
		permissions[i] = domain.Permission{
			Name:    route.Name,
			Route:   fmt.Sprintf("%s %s", route.Method, route.URL),
			Path:    route.URL,
			Method:  route.Method,
			Service: "approval-flow",
		}
	}

	// Store in database
	if err := s.Repo.BulkUpsertPermissions(ctx, permissions); err != nil {
		return fmt.Errorf("failed to store permissions in database: %w", err)
	}

	// Store in Redis
	redisKey := "api-gateway-permissions"
	if err := s.Cache.SetJSON(ctx, redisKey, permissions, 0); err != nil {
		s.Logger.Error("failed to cache permissions in Redis", zap.Error(err))
	}

	return nil
}

// IsExcludedRoute checks if a route should be excluded from permission checks
func IsExcludedRoute(path string) bool {
	excludedRoutes := []string{
		"/health",
		"/ws",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/metrics",
	}

	for _, excluded := range excludedRoutes {
		if path == excluded {
			return true
		}
	}
	return false
}

// FormatRouteName formats a route path into a human-readable name
func FormatRouteName(path string) string {
	if path == "/ws" {
		return "WebSocket Connection"
	}
	cleaned := strings.Replace(path, "/api/v1", "", 1)
	cleaned = strings.ReplaceAll(cleaned, "/", " ")
	titleCaser := cases.Title(language.English)
	return titleCaser.String(strings.TrimSpace(cleaned))
}

// computeTokenHash computes a hash of the token and user ID for SSO validation
func computeTokenHash(token, userID string) string {
	value := md5.Sum([]byte(token + userID))
	return fmt.Sprintf("%x", value)
}

// CheckPermission checks if a user has a specific permission
func (s *Service) CheckPermission(ctx context.Context, userID uuid.UUID, permissionRoute string) (bool, error) {
	user, err := s.Repo.GetUserWithRolesAndPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.Route == permissionRoute {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetAllUserPermissions retrieves all unique permissions for a user across all roles
func (s *Service) GetAllUserPermissions(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	user, err := s.Repo.GetUserWithRolesAndPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[uuid.UUID]domain.Permission)
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissionMap[permission.ID] = permission
		}
	}

	permissions := make([]domain.Permission, 0, len(permissionMap))
	for _, p := range permissionMap {
		permissions = append(permissions, p)
	}

	return permissions, nil
}
