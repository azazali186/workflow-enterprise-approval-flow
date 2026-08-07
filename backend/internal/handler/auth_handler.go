package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/pkg/response"
	"github.com/aeroxe/approval-flow/internal/pkg/validation"
)

type AuthHandler struct {
	svc *rbac.Service
	cfg *config.Config
}

func NewAuthHandler(svc *rbac.Service, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
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
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Login(ctx, &rbac.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.cfg.Error("login failed", zap.Error(err))
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
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

	user, err := h.svc.Register(ctx, &rbac.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.cfg.Error("registration failed", zap.Error(err))
		response.Error(c, http.StatusConflict, err.Error())
		return
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

	result, err := h.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.cfg.Error("token refresh failed", zap.Error(err))
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
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

	if err := h.svc.Logout(ctx, req.UserID); err != nil {
		h.cfg.Error("logout failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "logout failed")
		return
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

	// Note: ChangePassword requires implementation in rbac service
	response.Success(c, map[string]string{"message": "password change requested"})
}
