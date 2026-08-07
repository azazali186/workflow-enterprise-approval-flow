package server

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/network"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"

	"github.com/aeroxe/approval-flow/docs"
	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/handler"
	"github.com/aeroxe/approval-flow/internal/modules/analytics"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
	"github.com/aeroxe/approval-flow/internal/modules/application"
	"github.com/aeroxe/approval-flow/internal/modules/escalation"
	"github.com/aeroxe/approval-flow/internal/modules/login_log"
	"github.com/aeroxe/approval-flow/internal/modules/notification"
	"github.com/aeroxe/approval-flow/internal/modules/rbac"
	"github.com/aeroxe/approval-flow/internal/modules/report"
	"github.com/aeroxe/approval-flow/internal/modules/template"
	"github.com/aeroxe/approval-flow/internal/modules/workflow"
	auditmod "github.com/aeroxe/approval-flow/internal/modules/audit"
	"github.com/aeroxe/approval-flow/internal/saga"
	"github.com/aeroxe/approval-flow/internal/pkg/auth"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/database"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/middleware"
	wsHub "github.com/aeroxe/approval-flow/internal/pkg/websocket"
)

type Server struct {
	cfg      *config.Config
	engine   *server.Hertz
	db       *database.DB
	redis    *cache.Redis
	nats     *messaging.NATS
	hub      *wsHub.Hub
	tokenSvc *auth.TokenService
	rbacSvc  *rbac.Service
}

func NewServer(cfg *config.Config) (*Server, error) {
	docs.SwaggerInfo.Title = cfg.AppName + " API"
	docs.SwaggerInfo.Description = "Approval Flow Enterprise - Workflow Management System"
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.ServerPort)
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	h := server.New(server.WithHostPorts(fmt.Sprintf(":%d", cfg.ServerPort)))

	db, err := database.New(cfg)
	if err != nil {
		return nil, err
	}

	redis, err := cache.New(cfg)
	if err != nil {
		return nil, err
	}

	nats, err := messaging.New(cfg)
	if err != nil {
		return nil, err
	}

	hub := wsHub.NewHub(cfg)
	go hub.Run()

	tokenSvc := auth.NewTokenService(cfg)
	rbacRepo := rbac.NewRepository(db.Conn)
	rbacSvc := rbac.NewService(rbacRepo, redis, tokenSvc, cfg)

	s := &Server{
		cfg:      cfg,
		engine:   h,
		db:       db,
		redis:    redis,
		nats:     nats,
		hub:      hub,
		tokenSvc: tokenSvc,
		rbacSvc:  rbacSvc,
	}

	s.registerMiddleware()
	s.registerRoutes()

	return s, nil
}

func (s *Server) registerMiddleware() {
	s.engine.Use(middleware.RequestID())
	s.engine.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	s.engine.Use(middleware.Recovery(s.cfg))
	s.engine.Use(middleware.SecurityHeaders())
	s.engine.Use(middleware.RateLimiter(s.redis, 100, time.Minute, s.cfg))
	s.engine.Use(middleware.BodySizeLimit(middleware.MaxBodySize, s.cfg))
	s.engine.Use(middleware.TimeoutMiddleware(s.cfg))
	s.engine.Use(middleware.PrometheusMiddleware(s.cfg))
	s.engine.Use(middleware.DistributedLockMiddleware(s.redis, s.cfg))
	s.engine.Use(middleware.AuditMiddleware(s.db.Conn, s.cfg))

	// Initialize circuit breakers
	middleware.InitCircuitBreakers(s.cfg)
}

func (s *Server) registerRoutes() {
	approvalRepo := approval.NewRepository(s.db.Conn)
	applicationRepo := application.NewRepository(s.db.Conn)
	workflowRepo := workflow.NewRepository(s.db.Conn)
	notificationRepo := notification.NewRepository(s.db.Conn)
	templateRepo := template.NewRepository(s.db.Conn)
	escalationRepo := escalation.NewRepository(s.db.Conn)
	analyticsRepo := analytics.NewRepository(s.db.Conn)
	reportRepo := report.NewRepository(s.db.Conn)

	approvalSvc := approval.NewService(approvalRepo, s.redis, s.nats, s.hub, s.cfg)
	applicationSvc := application.NewService(applicationRepo, s.redis, s.nats, s.hub, approvalSvc, s.cfg)
	workflowSvc := workflow.NewService(workflowRepo, s.redis, s.nats, s.hub, s.cfg)
	notificationSvc := notification.NewService(notificationRepo, s.redis, s.nats, s.cfg)
	templateSvc := template.NewService(templateRepo, s.redis, s.nats, s.cfg)
	escalationSvc := escalation.NewService(escalationRepo, s.nats, s.hub, s.cfg)
	analyticsSvc := analytics.NewService(analyticsRepo, s.cfg)
	reportSvc := report.NewService(reportRepo, s.redis, s.cfg)

	// Initialize audit & login log services
	loginLogRepo := login_log.NewRepository(s.db.Conn)
	loginLogSvc := login_log.NewService(loginLogRepo, s.redis, s.cfg)
	auditSvc := auditmod.NewService(s.db.Conn, s.cfg)

	authHandler := handler.NewAuthHandler(s.rbacSvc, loginLogSvc, auditSvc, s.db.Conn, s.cfg)
	loginLogHandler := handler.NewLoginLogHandler(loginLogSvc, s.cfg)
	rbacHandler := handler.NewRBACHandler(s.rbacSvc, s.cfg)
	approvalHandler := handler.NewApprovalHandler(approvalSvc, s.cfg)
	applicationHandler := handler.NewApplicationHandler(applicationSvc, s.cfg)
	workflowHandler := handler.NewWorkflowHandler(workflowSvc, s.cfg)
	notificationHandler := handler.NewNotificationHandler(notificationSvc, s.cfg)
	templateHandler := handler.NewTemplateHandler(templateSvc, s.cfg)
	escalationHandler := handler.NewEscalationHandler(escalationSvc, s.cfg)
	reportHandler := handler.NewReportHandler(reportSvc, s.cfg)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc, s.cfg)
	healthHandler := handler.NewHealthHandler(s.db, s.redis, s.nats, s.cfg)

	// ==================== Swagger Docs ====================
	s.engine.POST("/docs/swagger.json", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Type", "application/json")
		c.JSON(consts.StatusOK, map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":       s.cfg.AppName + " API",
				"version":     "1.0.0",
				"description": "Approval Flow Enterprise - Workflow Management System",
			},
			"paths": map[string]interface{}{},
		})
	})

	// ==================== Prometheus Metrics ====================
	s.engine.POST("/metrics", func(ctx context.Context, c *app.RequestContext) {
		metrics := middleware.GetPrometheusMetrics()
		c.JSON(consts.StatusOK, metrics.ToJSON())
	})

	// ==================== Health (Public) ====================
	s.engine.POST("/health", healthHandler.HealthCheck)
	s.engine.POST("/health/detailed", healthHandler.DetailedHealthCheck)
	s.engine.POST("/health/ready", healthHandler.ReadyCheck)
	s.engine.POST("/health/live", healthHandler.LiveCheck)
	s.engine.POST("/version", healthHandler.Version)

	// ==================== Auth (Public) ====================
	s.engine.POST("/api/v1/auth/login", authHandler.Login)
	s.engine.POST("/api/v1/auth/register", authHandler.Register)
	s.engine.POST("/api/v1/auth/refresh", authHandler.RefreshToken)
	s.engine.POST("/api/v1/auth/login-history", loginLogHandler.GetLoginHistory)
	s.engine.POST("/api/v1/auth/login-history/email", loginLogHandler.GetLoginHistoryByEmail)
	s.engine.POST("/api/v1/auth/login-stats", loginLogHandler.GetLoginStats)

	// ==================== Protected Routes ====================
	v1 := s.engine.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(s.tokenSvc, s.redis, s.cfg))
	v1.Use(middleware.RBACMiddleware(s.rbacSvc, s.cfg))

	// Profile
	v1.POST("/profile", authHandler.GetProfile)
	v1.POST("/logout", authHandler.Logout)
	v1.POST("/change-password", authHandler.ChangePassword)

	// ==================== RBAC Admin ====================
	admin := v1.Group("/admin")
	admin.Use(middleware.RequireRole("admin", s.rbacSvc, s.cfg))

	// Users
	admin.POST("/users", rbacHandler.ListUsers)
	admin.POST("/users/get", rbacHandler.GetUser)
	admin.POST("/users/update", rbacHandler.UpdateUser)
	admin.POST("/users/delete", rbacHandler.DeleteUser)

	// Roles
	admin.POST("/roles", rbacHandler.ListRoles)
	admin.POST("/roles/get", rbacHandler.GetRole)
	admin.POST("/roles/create", rbacHandler.CreateRole)
	admin.POST("/roles/update", rbacHandler.UpdateRole)
	admin.POST("/roles/delete", rbacHandler.DeleteRole)

	// Role-Permission
	admin.POST("/roles/permissions", rbacHandler.GetRolePermissions)
	admin.POST("/roles/permissions/assign", rbacHandler.AssignPermissionToRole)
	admin.POST("/roles/permissions/remove", rbacHandler.RemovePermissionFromRole)

	// User-Role
	admin.POST("/users/roles", rbacHandler.GetUserRoles)
	admin.POST("/users/roles/assign", rbacHandler.AssignRoleToUser)
	admin.POST("/users/roles/remove", rbacHandler.RemoveRoleFromUser)

	// Permissions
	admin.POST("/permissions", rbacHandler.ListPermissions)
	admin.POST("/permissions/get", rbacHandler.GetPermission)
	admin.POST("/permissions/create", rbacHandler.CreatePermission)
	admin.POST("/permissions/update", rbacHandler.UpdatePermission)
	admin.POST("/permissions/delete", rbacHandler.DeletePermission)

	// Login Logs (admin)
	admin.POST("/login-logs", loginLogHandler.GetLoginLogs)

	// ==================== Applications ====================
	v1.POST("/applications", applicationHandler.GetApplications)
	v1.POST("/applications/get", applicationHandler.GetApplication)
	v1.POST("/applications/submit", applicationHandler.SubmitApplication)
	v1.POST("/applications/update", applicationHandler.UpdateApplication)
	v1.POST("/applications/delete", applicationHandler.DeleteApplication)

	// ==================== Approvals ====================
	v1.POST("/approvals", approvalHandler.GetApprovals)
	v1.POST("/approvals/get", approvalHandler.GetApproval)
	v1.POST("/approvals/create", approvalHandler.CreateApproval)
	v1.POST("/approvals/decide", approvalHandler.DecideApproval)
	v1.POST("/approvals/update", approvalHandler.UpdateApproval)
	v1.POST("/approvals/delete", approvalHandler.DeleteApproval)
	v1.POST("/approvals/pending", approvalHandler.GetPendingApprovals)

	// ==================== Workflows ====================
	v1.POST("/workflows", workflowHandler.GetWorkflows)
	v1.POST("/workflows/get", workflowHandler.GetWorkflow)
	v1.POST("/workflows/create", workflowHandler.CreateWorkflow)
	v1.POST("/workflows/update", workflowHandler.UpdateWorkflow)
	v1.POST("/workflows/delete", workflowHandler.DeleteWorkflow)

	// ==================== Templates ====================
	v1.POST("/templates", templateHandler.GetTemplates)
	v1.POST("/templates/get", templateHandler.GetTemplate)
	v1.POST("/templates/create", templateHandler.CreateTemplate)
	v1.POST("/templates/update", templateHandler.UpdateTemplate)
	v1.POST("/templates/delete", templateHandler.DeleteTemplate)

	// ==================== Escalations ====================
	v1.POST("/escalations", escalationHandler.GetEscalations)
	v1.POST("/escalations/get", escalationHandler.GetEscalation)
	v1.POST("/escalations/create", escalationHandler.CreateEscalation)
	v1.POST("/escalations/resolve", escalationHandler.ResolveEscalation)

	// ==================== Notifications ====================
	v1.POST("/notifications", notificationHandler.GetNotifications)
	v1.POST("/notifications/unread", notificationHandler.GetUnreadNotifications)
	v1.POST("/notifications/send", notificationHandler.SendNotification)
	v1.POST("/notifications/read", notificationHandler.MarkAsRead)
	v1.POST("/notifications/stats", notificationHandler.GetNotificationStats)

	// ==================== Reports ====================
	v1.POST("/reports/statuses", reportHandler.GetStatuses)
	v1.POST("/reports/comments", reportHandler.GetComments)
	v1.POST("/reports/documents", reportHandler.GetDocuments)

	// ==================== Analytics ====================
	v1.POST("/analytics/approvals", analyticsHandler.GetApprovalStats)
	v1.POST("/analytics/workflows", analyticsHandler.GetWorkflowPerformance)
	v1.POST("/analytics/approvers", analyticsHandler.GetApproverPerformance)
	v1.POST("/analytics/escalations", analyticsHandler.GetEscalationMetrics)

	// ==================== WebSocket (Public) ====================
	s.engine.POST("/ws", s.handleWebSocket)
}

func (s *Server) handleWebSocket(c context.Context, ctx *app.RequestContext) {
	// Parse WebSocket upgrade request from body/headers
	userID := string(ctx.GetHeader("X-User-ID"))
	if userID == "" {
		// Fallback: try to read from request body
		var wsReq struct {
			UserID   string `json:"user_id"`
			ClientID string `json:"client_id"`
		}
		if err := ctx.BindAndValidate(&wsReq); err == nil && wsReq.UserID != "" {
			userID = wsReq.UserID
		}
	}
	if userID == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "user_id is required"})
		return
	}

	clientID := string(ctx.GetHeader("X-Client-ID"))
	if clientID == "" {
		clientID = userID
	}
	wsKey := string(ctx.GetHeader("Sec-WebSocket-Key"))

	ctx.Hijack(func(conn network.Conn) {
		if wsKey == "" {
			conn.Close()
			return
		}

		acceptKey := computeWSAcceptKey(wsKey)
		response := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n"+
			"\r\n", acceptKey)

		if _, err := conn.Write([]byte(response)); err != nil {
			conn.Close()
			return
		}

		wsConn := &basicWSConn{conn: conn}

		client := &wsHub.Client{
			ID:     clientID,
			UserID: userID,
			Conn:   wsConn,
			Send:   make(chan []byte, 256),
			Hub:    s.hub,
		}

		s.hub.Register <- client

		go client.ReadPump()
		go client.WritePump()
	})
}

func computeWSAcceptKey(key string) string {
	const wsGUID = "258EAFA5-E914-47DA-95CA-5AB9A8095E44"
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type basicWSConn struct {
	conn   network.Conn
	mu     sync.Mutex
	closed bool
}

func (c *basicWSConn) Read(p []byte) (n int, err error) {
	return c.conn.Read(p)
}

func (c *basicWSConn) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	return c.conn.Write(p)
}

func (c *basicWSConn) Close(code int, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

func (c *basicWSConn) Ping(ctx context.Context) error { return nil }
func (c *basicWSConn) Pong(ctx context.Context) error { return nil }

func (s *Server) Start() error {
	s.cfg.Info("starting server", zap.Int("port", s.cfg.ServerPort))

	// Extract and store route permissions on startup
	if err := middleware.ExtractAndStoreRoutes(s.engine, s.rbacSvc, s.redis, s.cfg); err != nil {
		s.cfg.Error("failed to extract and store route permissions", zap.Error(err))
	}

	s.seedDefaultRoles()

	// Start outbox processor for reliable event publishing
	outbox := middleware.NewOutbox(s.redis, s.nats, s.cfg)
	outbox.StartProcessor(context.Background())
	s.cfg.Info("outbox processor started")

	// Start Saga Orchestrator
	orchestrator := saga.NewOrchestrator(s.nats, s.redis, s.hub, s.cfg)
	if err := orchestrator.Start(context.Background()); err != nil {
		s.cfg.Error("failed to start saga orchestrator", zap.Error(err))
	} else {
		s.cfg.Info("saga orchestrator started")
	}

	return s.engine.Run()
}

func (s *Server) seedDefaultRoles() {
	ctx := context.Background()

	adminRole := &domain.Role{
		Name:        "admin",
		Description: "System administrator with full access",
		IsDefault:   false,
	}
	if err := s.rbacSvc.CreateRole(ctx, adminRole); err != nil {
		s.cfg.Debug("admin role may already exist", zap.Error(err))
	}

	userRole := &domain.Role{
		Name:        "user",
		Description: "Regular user with basic access",
		IsDefault:   true,
	}
	if err := s.rbacSvc.CreateRole(ctx, userRole); err != nil {
		s.cfg.Debug("user role may already exist", zap.Error(err))
	}

	viewerRole := &domain.Role{
		Name:        "viewer",
		Description: "Read-only access",
		IsDefault:   false,
	}
	if err := s.rbacSvc.CreateRole(ctx, viewerRole); err != nil {
		s.cfg.Debug("viewer role may already exist", zap.Error(err))
	}

	s.cfg.Info("default roles seeded")
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.cfg.Info("shutting down server")
	s.engine.Close()
	if s.nats != nil {
		s.nats.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Server) Migrate() error {
	s.cfg.Info("running auto-migration")
	return s.db.AutoMigrate(
		// Note: For production, use golang-migrate instead. See internal/config/migration.go
		&domain.LoginLog{},
		&domain.Workflow{},
		&domain.WorkflowStep{},
		&domain.Template{},
		&domain.Application{},
		&domain.Approval{},
		&domain.Escalation{},
		&domain.Notification{},
		&domain.Status{},
		&domain.Comment{},
		&domain.Document{},
		&domain.AuditLog{},
		&domain.User{},
		&domain.Role{},
		&domain.Permission{},
		&domain.UserRole{},
		&domain.RolePermission{},
	)
}

// RunMigrations runs golang-migrate migrations if available, falls back to AutoMigrate
func (s *Server) RunMigrations(migrationsDir string) error {
	// Try golang-migrate first
	if err := config.RunMigrationsFromMain(s.cfg, migrationsDir); err != nil {
		s.cfg.Warn("golang-migrate failed, falling back to AutoMigrate", zap.Error(err))
		return s.Migrate()
	}

	// Always run AutoMigrate to ensure all models are registered
	// This is safe - AutoMigrate won't drop columns or change types
	return s.Migrate()
}
