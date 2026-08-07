package server

import (
	"context"
	"net/http"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/domain"
	"github.com/aeroxe/approval-flow/internal/handler"
	"github.com/aeroxe/approval-flow/internal/modules/analytics"
	"github.com/aeroxe/approval-flow/internal/modules/approval"
	"github.com/aeroxe/approval-flow/internal/modules/application"
	"github.com/aeroxe/approval-flow/internal/modules/escalation"
	"github.com/aeroxe/approval-flow/internal/modules/notification"
	"github.com/aeroxe/approval-flow/internal/modules/report"
	"github.com/aeroxe/approval-flow/internal/modules/workflow"
	"github.com/aeroxe/approval-flow/internal/pkg/auth"
	"github.com/aeroxe/approval-flow/internal/pkg/cache"
	"github.com/aeroxe/approval-flow/internal/pkg/database"
	"github.com/aeroxe/approval-flow/internal/pkg/messaging"
	"github.com/aeroxe/approval-flow/internal/pkg/websocket"
)

type Server struct {
	cfg      *config.Config
	engine   *server.Engine
	db       *database.DB
	redis    *cache.Redis
	nats     *messaging.NATS
	hub      *websocket.Hub
}

func NewServer(cfg *config.Config) (*Server, error) {
	h := server.Default(server.WithHostPorts(cfg.ServerPort))

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

	hub := websocket.NewHub(cfg)
	go hub.Run()

	s := &Server{
		cfg:    cfg,
		engine: h,
		db:     db,
		redis:  redis,
		nats:   nats,
		hub:    hub,
	}

	s.registerRoutes()

	return s, nil
}

func (s *Server) registerRoutes() {
	approvalRepo := approval.NewRepository(s.db.Conn)
	applicationRepo := application.NewRepository(s.db.Conn)
	workflowRepo := workflow.NewRepository(s.db.Conn)
	notificationRepo := notification.NewRepository(s.db.Conn)
	templateRepo := report.NewRepository(s.db.Conn)
	escalationRepo := escalation.NewRepository(s.db.Conn)
	analyticsRepo := analytics.NewRepository(s.db.Conn)
	reportRepo := report.NewRepository(s.db.Conn)

	approvalSvc := approval.NewService(approvalRepo, s.redis, s.nats, s.hub, s.cfg)
	applicationSvc := application.NewService(applicationRepo, s.redis, s.nats, s.hub, approvalSvc, s.cfg)
	workflowSvc := workflow.NewService(workflowRepo, s.redis, s.nats, s.hub, s.cfg)
	notificationSvc := notification.NewService(notificationRepo, s.redis, s.nats, s.cfg)
	escalationSvc := escalation.NewService(escalationRepo, s.nats, s.hub, s.cfg)
	analyticsSvc := analytics.NewService(analyticsRepo, s.cfg)
	reportSvc := report.NewService(reportRepo, s.redis, s.cfg)

	_ = auth.NewTokenService(s.cfg)

	approvalHandler := handler.NewApprovalHandler(approvalSvc, s.cfg)
	applicationHandler := handler.NewApplicationHandler(applicationSvc, s.cfg)
	workflowHandler := handler.NewWorkflowHandler(workflowSvc, s.cfg)
	notificationHandler := handler.NewNotificationHandler(notificationSvc, s.cfg)

	s.engine.GET("/health", func(ctx app.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/approvals/pending/:approver_id", approvalHandler.GetPendingApprovals)
		v1.POST("/approvals/:id/decide", approvalHandler.DecideApproval)
		v1.GET("/approvals/:id", approvalHandler.GetApproval)

		v1.POST("/applications", applicationHandler.SubmitApplication)
		v1.GET("/applications/:id", applicationHandler.GetApplication)
		v1.GET("/applications/applicant/:applicant_id", applicationHandler.GetApplicationsByApplicant)

		v1.POST("/workflows", workflowHandler.CreateWorkflow)
		v1.GET("/workflows/:id", workflowHandler.GetWorkflow)
		v1.GET("/workflows", workflowHandler.GetWorkflows)

		v1.GET("/notifications/:user_id", notificationHandler.GetNotifications)
		v1.GET("/notifications/:user_id/unread", notificationHandler.GetUnreadNotifications)
		v1.POST("/notifications/:id/read", notificationHandler.MarkAsRead)
	}

	s.engine.GET("/ws", s.handleWebSocket)
}

func (s *Server) handleWebSocket(ctx app.RequestContext) {
	userID := ctx.Query("user_id")
	if userID == "" {
		ctx.AbortWithStatus(consts.StatusBadRequest)
		return
	}

	c, err := websocket.Accept(ctx, ctx.Request.Header, nil)
	if err != nil {
		ctx.AbortWithStatus(consts.StatusBadRequest)
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	client := &websocket.Client{
		ID:     ctx.Query("client_id"),
		UserID: userID,
		Conn:   c,
		Send:   make(chan []byte, 256),
		Hub:    s.hub,
	}

	s.hub.Register <- client

	errCh := make(chan error, 1)
	go func() {
		for msg := range client.Send {
			if err := wsjson.Write(ctx, c, msg); err != nil {
				errCh <- err
				return
			}
		}
	}()

	<-errCh
}

func (s *Server) Start() error {
	s.cfg.Info("starting server", "port", s.cfg.ServerPort)
	return s.engine.Run()
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
	)
}
