package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	auditmod "github.com/aeroxe/approval-flow/internal/modules/audit"
	"github.com/aeroxe/approval-flow/internal/modules/login_log"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/pkg/middleware"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
	"gorm.io/gorm"
)

type AuthHandler struct {
	svc      *rbac.Service
	loginLog *login_log.Service
	auditSvc *auditmod.Service
	db       *gorm.DB
	cfg      *config.Config
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
			EntityType: "user",
			EntityID:   result.User.ID,
			Action:     domain.AuditActionLogin,
			ActorID:    &result.User.ID,
			ActorEmail: &result.User.Email,
			AfterState: result.User,
			Request:    reqCtx,
			Status:     domain.AuditStatusSuccess,
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
		// Surface the known business conflict verbatim; any other failure is
		// internal and must never leak wrapped details (e.g. DB drivers).
		if errors.Is(err, rbac.ErrEmailExists) {
			response.Error(c, http.StatusConflict, "user with this email already exists")
			return
		}
		response.Error(c, http.StatusInternalServerError, "registration failed")
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
			EntityType: "user",
			EntityID:   user.ID,
			Action:     domain.AuditActionRegister,
			ActorID:    &user.ID,
			ActorEmail: &user.Email,
			AfterState: user,
			Request:    reqCtx,
			Status:     domain.AuditStatusSuccess,
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

	// Record the presented refresh token's jti for the audit trail.
	tokenID := ""
	if h.svc.Token != nil {
		tokenID = h.svc.Token.ExtractTokenID(req.RefreshToken)
	}

	result, err := h.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		// Distinguish reuse (stolen/rotated token) from ordinary invalidity.
		failureReason := domain.FailureReasonTokenInvalid
		if errors.Is(err, rbac.ErrRefreshTokenReuse) {
			failureReason = domain.FailureReasonTokenReuse
		}

		// Log failed token refresh
		h.loginLog.LogLoginFailure(ctx, "", failureReason,
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
			TokenID:   tokenID,
		})

		// Create audit log
		if h.auditSvc != nil {
			h.auditSvc.Log(ctx, auditmod.LogParams{
				EntityType: "user",
				EntityID:   result.User.ID,
				Action:     domain.AuditActionRefreshToken,
				ActorID:    &result.User.ID,
				ActorEmail: &result.User.Email,
				Request:    reqCtx,
				Status:     domain.AuditStatusSuccess,
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
	// Identity always comes from the authenticated token, never the body, so a
	// user can never read another user's profile (IDOR prevention). The body is
	// tolerated for backward compatibility but ignored.
	var req validation.GetProfileRequest
	_ = c.BindAndValidate(&req)

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr)
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
	// Identity always comes from the authenticated token, never the body.
	var req validation.LogoutRequest
	_ = c.BindAndValidate(&req)

	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	reqCtx := h.extractRequestContext(c)

	userID, _ := uuid.Parse(userIDStr)

	if err := h.svc.Logout(ctx, userIDStr); err != nil {
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
	userIDStr := middleware.GetUserIDFromContext(c)
	if userIDStr == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	reqCtx := h.extractRequestContext(c)
	userID, _ := uuid.Parse(userIDStr)

	// Get user before change (for audit log)
	user, err := h.svc.GetUser(ctx, userID)
	if err != nil {
		h.cfg.Error("failed to get user for password change", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to change password")
		return
	}

	// Change password via RBAC service
	err = h.svc.ChangePassword(ctx, userID, &rbac.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		h.cfg.Error("password change failed", zap.Error(err), zap.String("user_id", userID.String()))

		// Determine appropriate error message
		errMsg := "failed to change password"
		status := http.StatusInternalServerError

		switch err.Error() {
		case "invalid old password":
			errMsg = "current password is incorrect"
			status = http.StatusBadRequest
		case "new password must be different from old password":
			errMsg = "new password must be different from current password"
			status = http.StatusBadRequest
		}

		response.Error(c, status, errMsg)
		return
	}

	// Create audit log for password change
	if h.auditSvc != nil {
		h.auditSvc.Log(ctx, auditmod.LogParams{
			EntityType: "user",
			EntityID:   userID,
			Action:     domain.AuditActionChangePassword,
			ActorID:    &userID,
			ActorEmail: &user.Email,
			BeforeState: map[string]interface{}{
				"email":    user.Email,
				"name":     user.Name,
				"password": "[REDACTED]",
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

	response.Success(c, map[string]string{"message": "password changed successfully"})
}
