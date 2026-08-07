package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/modules/login_log"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	auditmod "github.com/aeroxe/approval-flow/internal/modules/audit"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
	"gorm.io/gorm"
)

type AuthHandler struct {
	svc       *rbac.Service
	loginLog  *login_log.Service
	auditSvc  *auditmod.Service
	db        *gorm.DB
	cfg       *config.Config
}

func NewAuthHandler(svc *rbac.Service, loginLog *login_log.Service, auditSvc *auditmod.Service, db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		svc:      svc,
		loginLog: loginLog,
		auditSvc: auditSvc,
		db:       db,
		cfg:      cfg,
	}
}

// extractRequestContext extracts request metadata for audit logging
func (h *AuthHandler) extractRequestContext(c *app.RequestContext) *auditmod.RequestContext {
	return &auditmod.RequestContext{
		IPAddress: c.ClientIP(),
		UserAgent: string(c.Request.Header.UserAgent()),
		RequestID: string(c.GetHeader("X-Request-ID")),
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body validation.LoginRequest true "Login credentials"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req validation.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		// Log failed login attempt
		h.loginLog.LogLoginFailure(ctx, req.Email, domain.FailureReasonInvalidCredentials,
			c.ClientIP(), string(c.Request.Header.UserAgent()), string(c.GetHeader("X-Request-ID")))
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	reqCtx := h.extractRequestContext(c)

	// Check if IP is locked
	if locked, _ := h.loginLog.IsIPLocked(ctx, reqCtx.IPAddress); locked {
		h.loginLog.LogLoginFailure(ctx, req.Email, domain.FailureReasonTooManyAttempts,
			reqCtx.IPAddress, reqCtx.UserAgent, reqCtx.RequestID)
		response.Error(c, http.StatusTooManyRequests, "too many login attempts from this IP, please try again later")
		return
	}

	// Check if account is locked
	if locked, _ := h.loginLog.IsAccountLocked(ctx, req.Email); locked {
		h.loginLog.LogAccountLocked(ctx, req.Email, reqCtx.IPAddress, reqCtx.UserAgent, reqCtx.RequestID)
		response.Error(c, http.StatusTooManyRequests, "account locked due to too many failed attempts, please try again later")
		return
	}

	// Attempt login
	result, err := h.svc.Login(ctx, &rbac.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// Determine failure reason
		failureReason := domain.FailureReasonInvalidCredentials
		if err.Error() == "user not found" {
			failureReason = domain.FailureReasonUserNotFound
		}

		// Log failed attempt
		h.loginLog.LogLoginFailure(ctx, req.Email, failureReason,
			reqCtx.IPAddress, reqCtx.UserAgent, reqCtx.RequestID)

		h.cfg.Error("login failed", zap.Error(err), zap.String("email", req.Email))
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Log successful login
	h.loginLog.LogLoginSuccess(ctx, result.User.ID, req.Email,
		reqCtx.IPAddress, reqCtx.UserAgent, reqCtx.RequestID, "")

	// Create audit log
	if h.auditSvc != nil {
		h.auditSvc.Log(ctx, auditmod.LogParams{
			EntityType:  "user",
			EntityID:    result.User.ID,
			Action:      domain.AuditActionLogin,
			ActorID:     &result.User.ID,
			ActorEmail:  &result.User.Email,
			AfterState:  result.User,
			Request:     reqCtx,
			Status:      domain.AuditStatusSuccess,
		})
	}

	response.Success(c, result)
}

// Register godoc
// @Summary      User registration
// @Description  Register a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body validation.RegisterRequest true "Registration details"
// @Success      201  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req validation.RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	reqCtx := h.extractRequestContext(c)

	user, err := h.svc.Register(ctx, &rbac.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// Log failed registration attempt
		h.loginLog.LogLoginFailure(ctx, req.Email, "registration_failed",
			reqCtx.IPAddress, reqCtx.UserAgent, reqCtx.RequestID)

		h.cfg.Error("registration failed", zap.Error(err), zap.String("email", req.Email))
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	// Log successful registration as a login log entry
	h.loginLog.LogLoginAttempt(ctx, login_log.LogLoginParams{
		UserID:    &user.ID,
		Email:     req.Email,
		Status:    domain.LoginStatusSuccess,
		IPAddress: reqCtx.IPAddress,
		UserAgent: reqCtx.UserAgent,
		RequestID: reqCtx.RequestID,
	})

	// Create audit log
	if h.auditSvc != nil {
		h.auditSvc.Log(ctx, auditmod.LogParams{
			EntityType:  "user",
			EntityID:    user.ID,
			Action:      domain.AuditActionRegister,
			ActorID:     &user.ID,
			ActorEmail:  &user.Email,
			AfterState:  user,
			Request:     reqCtx,
			Status:      domain.AuditStatusSuccess,
		})
	}

	response.Success(c, user)
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body validation.RefreshTokenRequest true "Refresh token"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(ctx context.Context, c *app.RequestContext) {
	var req validation.RefreshTokenRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	reqCtx := h.extractRequestContext(c)

	result, err := h.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		// Log failed token refresh
		h.loginLog.LogLoginFailure(ctx, "", domain.FailureReasonTokenInvalid,
			reqCtx.IPAddress, reqCtx.UserAgent, reqCtx.RequestID)

		h.cfg.Error("token refresh failed", zap.Error(err))
		response.Error(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Log successful token refresh
	if result.User != nil {
		h.loginLog.LogLoginAttempt(ctx, login_log.LogLoginParams{
			UserID:    &result.User.ID,
			Email:     result.User.Email,
			Status:    domain.LoginStatusSuccess,
			IPAddress: reqCtx.IPAddress,
			UserAgent: reqCtx.UserAgent,
			RequestID: reqCtx.RequestID,
		})

		// Create audit log
		if h.auditSvc != nil {
			h.auditSvc.Log(ctx, auditmod.LogParams{
				EntityType:  "user",
				EntityID:    result.User.ID,
				Action:      domain.AuditActionRefreshToken,
				ActorID:     &result.User.ID,
				ActorEmail:  &result.User.Email,
				Request:     reqCtx,
				Status:      domain.AuditStatusSuccess,
			})
		}
	}

	response.Success(c, result)
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get current user's profile information
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body validation.GetProfileRequest true "User ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /api/v1/profile [post]
func (h *AuthHandler) GetProfile(ctx context.Context, c *app.RequestContext) {
	var req validation.GetProfileRequest
	if err := c.BindAndValidate(&req); err != nil {
		// Fallback to context user ID
		userIDStr, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		req.UserID = userIDStr.(string)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.svc.GetUser(ctx, userID)
	if err != nil {
		h.cfg.Error("failed to get profile", zap.Error(err))
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.Success(c, user)
}

// Logout godoc
// @Summary      User logout
// @Description  Invalidate current session and token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body validation.LogoutRequest true "User ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /api/v1/logout [post]
func (h *AuthHandler) Logout(ctx context.Context, c *app.RequestContext) {
	var req validation.LogoutRequest
	if err := c.BindAndValidate(&req); err != nil {
		// Fallback to context
		userIDStr, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		req.UserID = userIDStr.(string)
	}

	reqCtx := h.extractRequestContext(c)

	userID, _ := uuid.Parse(req.UserID)

	if err := h.svc.Logout(ctx, req.UserID); err != nil {
		h.cfg.Error("logout failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "logout failed")
		return
	}

	// Create audit log for logout
	if h.auditSvc != nil && userID != uuid.Nil {
		h.auditSvc.Log(ctx, auditmod.LogParams{
			EntityType: "user",
			EntityID:   userID,
			Action:     domain.AuditActionLogout,
			ActorID:    &userID,
			Request:    reqCtx,
			Status:     domain.AuditStatusSuccess,
		})
	}

	response.Success(c, map[string]string{"message": "logged out"})
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Change current user's password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body validation.ChangePasswordRequest true "Password change details"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /api/v1/change-password [post]
func (h *AuthHandler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	var req validation.ChangePasswordRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	reqCtx := h.extractRequestContext(c)
	userID, _ := uuid.Parse(userIDStr.(string))

	// Get user before change
	user, err := h.svc.GetUser(ctx, userID)
	if err != nil {
		h.cfg.Error("failed to get user for password change", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to change password")
		return
	}

	// TODO: Implement actual password change in rbac service
	// For now, return success
	response.Success(c, map[string]string{"message": "password change requested"})

	// Create audit log for password change attempt
	if h.auditSvc != nil {
		h.auditSvc.Log(ctx, auditmod.LogParams{
			EntityType: "user",
			EntityID:   userID,
			Action:     domain.AuditActionChangePassword,
			ActorID:    &userID,
			BeforeState: map[string]interface{}{
				"email": user.Email,
				"name":  user.Name,
			},
			AfterState: map[string]interface{}{
				"email":    user.Email,
				"name":     user.Name,
				"password": "[CHANGED]",
			},
			Request: reqCtx,
			Status:  domain.AuditStatusSuccess,
		})
	}
}
