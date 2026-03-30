package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/pixelcraft/api/internal/config"
	"github.com/pixelcraft/api/internal/database"
	"github.com/pixelcraft/api/internal/handlers"
	"github.com/pixelcraft/api/internal/middleware"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
	"github.com/pixelcraft/api/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	// Initialize database connection
	db, err := database.NewPostgresDB(cfg.Database, cfg.CPFEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connection established")

	// Run database migrations (BT-015)
	if err := database.RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("Fatal error running DB migrations: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	productRepo := repository.NewProductRepository(db.DB)
	paymentRepo := repository.NewPaymentRepository(db.DB)
	discountRepo := repository.NewDiscountRepository(db.DB)
	libraryRepo := repository.NewLibraryRepository(db.DB)
	subscriptionRepo := repository.NewSubscriptionRepository(db.DB)
	adminRepo := repository.NewAdminRepository(db.DB)
	transactionRepo := repository.NewTransactionRepository(db)
	gameRepo := repository.NewGameRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)
	supportRepo := repository.NewSupportRepository(db.DB)
	permissionRepo := repository.NewPermissionRepository(db.DB)

	// Initialize services
	// Create Email Service with DB access for system settings
	emailService := service.NewEmailService(db.DB)
	permissionService := service.NewPermissionService(permissionRepo)
	
	authService := service.NewAuthService(userRepo, db.DB, struct{ Secret string }{Secret: cfg.JWT.Secret})
	userService := service.NewUserService(userRepo, cfg.CPFEncryptionKey)
	// File Service and Handler (with upload configuration)
	allowedTypes := []string{
		".jar", ".zip", ".exe",           // Executáveis e pacotes
		".png", ".jpg", ".jpeg", ".pdf",  // Imagens e documentos
		".txt", ".json", ".js",           // Arquivos de texto e código
	}
	fileService := service.NewFileService(db.DB, "./uploads", 500*1024*1024, allowedTypes) // 500MB max
	productService := service.NewProductService(db.DB, cfg, fileService)
	paymentService := service.NewPaymentService(paymentRepo) // Dependency injection (interface-based)
	// Initialize Mercado Pago Auth Service
	mpAuthService := service.NewMercadoPagoAuthService(cfg.MercadoPago)
	depositService := service.NewDepositService(
		transactionRepo,
		userRepo,
		paymentRepo,
		mpAuthService,
		cfg.MercadoPago.WebhookURL,
		service.DepositURLs{
			Success: cfg.MercadoPago.DepositSuccessURL,
			Failure: cfg.MercadoPago.DepositFailureURL,
			Pending: cfg.MercadoPago.DepositPendingURL,
		},
	)

	checkoutService := service.NewCheckoutService(db.DB, productRepo, discountRepo, paymentRepo, userRepo, subscriptionRepo, libraryRepo, depositService)
	depositService.SetCheckoutGateway(checkoutService) // Interface-based decoupling (no circular dependency)
	libraryService := service.NewLibraryService(libraryRepo, productRepo, fileService)
	historyService := service.NewHistoryService(paymentRepo, libraryRepo)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	discountService := service.NewDiscountService(discountRepo)

	// New domain-specific services (SRP compliance)
	balanceService := service.NewBalanceService(transactionRepo, userRepo, db.DB)
	userQueryService := service.NewUserQueryService(userRepo, transactionRepo, subscriptionRepo, libraryRepo, roleRepo)

	// Admin Service (orchestrator only - delegates to domain services)
	adminService := service.NewAdminService(adminRepo, balanceService, userQueryService, depositService)

	// AI Service
	aiService := service.NewAIService()

	// Role and Support Services
	roleService := service.NewRoleService(roleRepo, userRepo)
	supportService := service.NewSupportService(supportRepo, roleService)

	// Analytics Worker
	analyticsWorker := service.NewAnalyticsWorker(db.DB)
	go analyticsWorker.Start()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, emailService, roleService, cfg.JWT.Expiration)
	userHandler := handlers.NewUserHandler(userService, roleService)
	productHandler := handlers.NewProductHandler(productService)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)
	libraryHandler := handlers.NewLibraryHandler(libraryService)
	historyHandler := handlers.NewHistoryHandler(historyService)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)
	dashboardHandler := handlers.NewDashboardHandler(userService, paymentService)
	adminHandler := handlers.NewAdminHandler(adminService, analyticsWorker)
	depositHandler := handlers.NewDepositHandler(depositService, cfg.MercadoPago.WebhookSecret)
	transactionHandler := handlers.NewTransactionHandler(transactionRepo)
	discountHandler := handlers.NewDiscountHandler(discountService)
	
	// Message components
	messageRepo := repository.NewMessageRepository(db.DB)
	messageService := service.NewMessageService(messageRepo, subscriptionRepo)
	messageHandler := handlers.NewMessageHandler(messageService)

	// Admin Subscription Handler (needs messageService)
	adminSubscriptionHandler := handlers.NewAdminSubscriptionHandler(subscriptionService, messageService)

	// Game Handler
	gameHandler := handlers.NewGameHandler(gameRepo)

	// File Service and Handler (with upload configuration)
	fileHandler := handlers.NewFileHandler(fileService)

	// AI Handler
	aiHandler := handlers.NewAIHandler(aiService, userService)

	// WebSocket Hub
	wsHub := handlers.NewWSHub()
	go wsHub.Run()
	wsHandler := handlers.NewWSHandler(wsHub, cfg.JWT.Secret, cfg.CORS.AllowedOrigins)

	// Support Handler
	supportHandler := handlers.NewSupportHandler(supportService, roleService, wsHub)

	// Role Handler
	roleHandler := handlers.NewRoleHandler(roleService)

	// Permission Handler
	permissionHandler := handlers.NewPermissionHandler(permissionService)

	// Permission Advanced Handler
	permissionAdvancedService := service.NewPermissionAdvancedService(permissionRepo, db.DB)
	permissionAdvancedHandler := handlers.NewPermissionAdvancedHandler(permissionAdvancedService)

	// Email Management Handler
	emailManagementHandler := handlers.NewEmailManagementHandler(emailService)

	// Health Check Monitor
	healthHandler := handlers.NewHealthHandler(db)

	// System Handler
	systemHandler := handlers.NewSystemHandler()

	log.Printf("🚀 Server starting on %s", serverAddr)
	log.Printf("📚 API Documentation: http://%s/api/v1/health", serverAddr)

	// Set Gin to release mode (disable debug logs)
	gin.SetMode(gin.ReleaseMode)

	// CORS Configuration
	allowedOrigins := cfg.CORS.AllowedOrigins
	
	// Ensure production origin is included and trimmed
	foundProd := false
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
		if allowedOrigins[i] == "https://pixelcraft-studio.store" {
			foundProd = true
		}
	}
	if !foundProd {
		allowedOrigins = append(allowedOrigins, "https://pixelcraft-studio.store")
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Email-Password"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router := gin.Default()
	router.Use(cors.New(corsConfig))

	// Static file serving for public uploads (e.g. avatars)
	router.Static("/public", "./uploads/public")

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// WS Route
		v1.GET("/ws", wsHandler.ServeWS)

		// Health check (BT-019)
		v1.GET("/health", healthHandler.HealthCheck)

		// Auth routes (public with rate limiting)
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(5, time.Minute)) // 5 requests per minute per IP
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPasswordConfirm)
		}

		// Public Plans (for homepage)
		v1.GET("/plans", subscriptionHandler.ListPlans)

		// Webhook routes (rate limited - BT-034)
		webhook := v1.Group("/webhook")
		webhook.Use(middleware.RateLimitMiddleware(30, time.Minute)) // 30 requests per minute
		{
			webhook.POST("/mercadopago", depositHandler.Webhook)
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			users.GET("/me", userHandler.GetProfile)
			users.PUT("/me", userHandler.UpdateProfile)
			users.POST("/me/avatar", userHandler.UploadAvatar)
		}
		
		// Deposit routes (protected)
		v1.POST("/deposit", middleware.AuthMiddleware(cfg.JWT.Secret), depositHandler.Deposit)
		
		// Transaction routes (protected)
		v1.GET("/transactions", middleware.AuthMiddleware(cfg.JWT.Secret), transactionHandler.GetUserTransactions)
		v1.GET("/wallet/transactions/:id/status", middleware.AuthMiddleware(cfg.JWT.Secret), transactionHandler.GetTransactionStatus)

		// Product routes
		products := v1.Group("/products")
		{
			// Public routes
			products.GET("", productHandler.ListProducts)       // GET /api/v1/products
			products.GET("/:id", productHandler.GetProduct)      // GET /api/v1/products/:id
			
		// Protected routes - Requires Development+ (catalog edit access)
		products.POST("", middleware.AuthMiddleware(cfg.JWT.Secret), middleware.CatalogEditMiddleware(db.DB), productHandler.CreateProduct)       // POST /api/v1/products
		products.PUT("/:id", middleware.AuthMiddleware(cfg.JWT.Secret), middleware.CatalogEditMiddleware(db.DB), productHandler.UpdateProduct)    // PUT /api/v1/products/:id
		products.DELETE("/:id", middleware.AuthMiddleware(cfg.JWT.Secret), middleware.CatalogEditMiddleware(db.DB), productHandler.DeleteProduct) // DELETE /api/v1/products/:id
		}

		// Game routes (public)
		games := v1.Group("/games")
		{
			games.GET("", gameHandler.ListGames)                         // GET /api/v1/games
			games.GET("/with-categories", gameHandler.ListGamesWithCategories) // GET /api/v1/games/with-categories
			games.GET("/:id/categories", gameHandler.GetGameCategories)  // GET /api/v1/games/:id/categories
		}

		// Category routes (public)
		v1.GET("/categories", gameHandler.ListAllCategories) // GET /api/v1/categories



		// Checkout routes (protected)
		checkout := v1.Group("/checkout")
		checkout.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			checkout.POST("", checkoutHandler.ProcessCheckout) // POST /api/v1/checkout
		}

		// Discount routes (protected)
		discounts := v1.Group("/discounts")
		discounts.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			discounts.POST("/validate", checkoutHandler.ValidateDiscount) // POST /api/v1/discounts/validate
		}

		// Dashboard routes (protected)
		dashboard := v1.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			dashboard.GET("/stats", dashboardHandler.GetDashboardStats) // GET /api/v1/dashboard/stats
		}

		// Library routes (protected)
		library := v1.Group("/library")
		library.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			library.GET("", libraryHandler.GetMyLibrary) // GET /api/v1/library
			library.GET("/:id/download", libraryHandler.GetDownloadURL) // GET /api/v1/library/:id/download
		}

		// History routes (protected)
		history := v1.Group("/history")
		history.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			history.GET("", historyHandler.GetMyHistory) // GET /api/v1/history
			history.GET("/invoices", historyHandler.GetMyInvoices) // GET /api/v1/history/invoices
		}

		// Subscription routes (protected)
		subscriptions := v1.Group("/subscriptions")
		subscriptions.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			subscriptions.GET("", subscriptionHandler.ListUserSubscriptions) // GET /api/v1/subscriptions
			subscriptions.GET("/:id", subscriptionHandler.GetSubscriptionDetails) // GET /api/v1/subscriptions/:id
		}

		// Support routes (client area - protected)
		support := v1.Group("/support")
		support.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			support.POST("/tickets", supportHandler.CreateTicket)          // POST /api/v1/support/tickets
			support.GET("/tickets", supportHandler.ListMyTickets)          // GET /api/v1/support/tickets
			support.GET("/tickets/:id", supportHandler.GetTicket)          // GET /api/v1/support/tickets/:id
			support.POST("/tickets/:id/messages", supportHandler.SendMessage) // POST /api/v1/support/tickets/:id/messages
			support.PUT("/tickets/:id/close", supportHandler.CloseTicket)  // PUT /api/v1/support/tickets/:id/close
		}

		// Admin routes (protected + role-based access)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		admin.Use(middleware.AdminPanelMiddleware(db.DB)) // Any staff role can access base admin
		{
			admin.GET("/stats", adminHandler.GetStats)
			admin.POST("/stats/refresh", adminHandler.RefreshStats)
			admin.GET("/orders/recent", adminHandler.GetRecentOrders)
			admin.GET("/products/top", adminHandler.GetTopProducts)

			// Support Channel (Atendimento) - All staff can access
			admin.GET("/support/tickets", supportHandler.ListAllTickets)
			admin.GET("/support/tickets/:id", supportHandler.GetTicket)
			admin.POST("/support/tickets/:id/messages", supportHandler.SendMessage)
			admin.PUT("/support/tickets/:id/assign", supportHandler.AssignTicket)
			admin.PUT("/support/tickets/:id/release", supportHandler.ReleaseTicket)
			admin.PUT("/support/tickets/:id/status", supportHandler.UpdateStatus)
			admin.GET("/support/stats", supportHandler.GetTicketStats)

			// Finance Management - Admin, Engineering, Direction only
			admin.GET("/transactions", adminHandler.ListTransactions)
			admin.GET("/finance/balance", adminHandler.GetMPBalance)
			admin.POST("/transactions/:id/refund", adminHandler.RefundTransaction)
			
			// Admin Subscription Management
			admin.GET("/subscriptions/active", adminSubscriptionHandler.GetActiveSubscriptions)
			admin.GET("/subscriptions/:id", adminSubscriptionHandler.GetSubscriptionDetails)
			admin.PUT("/subscriptions/:id", adminSubscriptionHandler.UpdateSubscription)
			admin.POST("/subscriptions/:id/logs", adminSubscriptionHandler.CreateSubscriptionLog)
			
			// Admin Chat
			admin.GET("/subscriptions/:id/chat", adminSubscriptionHandler.GetSubscriptionChat)
			admin.POST("/subscriptions/:id/chat", adminSubscriptionHandler.SendSubscriptionMessage)

			// Admin Plan Management - Development, Engineering, Direction only
			admin.POST("/plans", adminSubscriptionHandler.CreatePlan)
			admin.PUT("/plans/:id", adminSubscriptionHandler.UpdatePlan)
			admin.DELETE("/plans/:id", adminSubscriptionHandler.DeletePlan)

			// Admin Game Management
			admin.POST("/games", gameHandler.CreateGame)
			admin.PUT("/games/:id", gameHandler.UpdateGame)
			admin.DELETE("/games/:id", gameHandler.DeleteGame)

			// Admin Category Management
			admin.POST("/categories", gameHandler.CreateCategory)
			admin.PUT("/categories/:id", gameHandler.UpdateCategory)
			admin.DELETE("/categories/:id", gameHandler.DeleteCategory)

			// Admin User Management
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/:id", adminHandler.GetUserDetail)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.PUT("/users/:id/password", adminHandler.UpdateUserPassword)

			// Admin File Management
			admin.GET("/files", fileHandler.ListAllFiles)

			// Admin Discount Management
			admin.GET("/discounts", discountHandler.ListDiscounts)
			admin.GET("/discounts/:id", discountHandler.GetDiscount)
			admin.POST("/discounts", discountHandler.CreateDiscount)
			admin.PUT("/discounts/:id", discountHandler.UpdateDiscount)
			admin.DELETE("/discounts/:id", discountHandler.DeleteDiscount)

			// Role Management - GET allowed for any admin
			admin.GET("/users/:id/roles", roleHandler.GetUserRoles)

			// Email Management - Requires EMAILS permission
			admin.POST("/emails/send", emailManagementHandler.SendEmail)
			admin.GET("/emails/logs", emailManagementHandler.GetEmailLogs)
			admin.GET("/emails/logs/:id", emailManagementHandler.GetEmailLogByID)
			admin.POST("/emails/logs/:id/resend", emailManagementHandler.ResendEmail)

			// Permission Management - Requires ROLES permission (DIRECTION only)
			admin.GET("/permissions/roles", permissionHandler.GetAllRolePermissions)
			admin.GET("/permissions/roles/:role", permissionHandler.GetRolePermissions)
			// POST e DELETE requerem ROLES:MANAGE (apenas DIRECTION) — CRÍTICO-04
			admin.POST("/permissions/roles/:role",
				middleware.RequirePermission(permissionService, models.ResourceRoles, models.ActionManage),
				permissionHandler.AddRolePermission)
			admin.DELETE("/permissions/roles/:role",
				middleware.RequirePermission(permissionService, models.ResourceRoles, models.ActionManage),
				permissionHandler.RemoveRolePermission)
			admin.GET("/permissions/resources", permissionHandler.GetAvailableResources)
			admin.GET("/permissions/actions", permissionHandler.GetAvailableActions)
			admin.GET("/permissions/available-roles", permissionHandler.GetAvailableRoles)

			// Advanced Permission Features
			admin.GET("/permissions/audit-log", permissionAdvancedHandler.GetPermissionAuditLog)
			admin.POST("/permissions/roles/:role/inherit", permissionAdvancedHandler.InheritPermissions)
			admin.DELETE("/permissions/roles/:role/inherited", permissionAdvancedHandler.RemoveInheritedPermissions)
			admin.GET("/permissions/custom-roles", permissionAdvancedHandler.GetCustomRoles)
			admin.POST("/permissions/custom-roles", permissionAdvancedHandler.CreateCustomRole)
			admin.DELETE("/permissions/custom-roles/:id", permissionAdvancedHandler.DeleteCustomRole)
			admin.GET("/permissions/export", permissionAdvancedHandler.ExportPermissions)
			admin.POST("/permissions/import", permissionAdvancedHandler.ImportPermissions)
			admin.GET("/permissions/templates", permissionAdvancedHandler.GetPermissionTemplates)
			admin.POST("/permissions/templates", permissionAdvancedHandler.SavePermissionTemplate)
			admin.GET("/permissions/dashboard", permissionAdvancedHandler.GetPermissionDashboard)

			// System Resources - Requires SYSTEM VIEW permission
			admin.GET("/system/metrics", systemHandler.GetSystemMetrics)
		}

		// User permissions route (any authenticated user)
		v1.GET("/permissions/me", middleware.AuthMiddleware(cfg.JWT.Secret), permissionHandler.GetMyPermissions)
		v1.GET("/permissions/notifications", middleware.AuthMiddleware(cfg.JWT.Secret), permissionAdvancedHandler.GetUserNotifications)
		v1.PUT("/permissions/notifications/:id/read", middleware.AuthMiddleware(cfg.JWT.Secret), permissionAdvancedHandler.MarkNotificationAsRead)

		// Role Management - POST/DELETE requer AdminPanel+ (hierarquia validada no handler) — MODERADO-05
		roleManagement := v1.Group("/admin")
		roleManagement.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		roleManagement.Use(middleware.AdminPanelMiddleware(db.DB))
		{
			roleManagement.POST("/users/:id/roles", roleHandler.AddUserRole)
			roleManagement.DELETE("/users/:id/roles/:role", roleHandler.RemoveUserRole)
		}

		// Message routes (Chat)
		// Note: Both users and admins can access these, but permissions are handled in service/handler
		// We use AuthMiddleware to ensure we have a user_id.
		// Admin status is checked via context or DB in handler/service.
		messages := v1.Group("/subscriptions/:id/chat")
		messages.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			messages.POST("", messageHandler.SendMessage)
			messages.GET("", messageHandler.GetMessages)
		}

		// File routes (protected - BT-047: split by authorization)
		files := v1.Group("/files")
		files.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			files.GET("", fileHandler.ListFiles)                      // GET /api/v1/files
			files.GET("/:id", fileHandler.GetFile)                    // GET /api/v1/files/:id
			files.GET("/:id/download", fileHandler.DownloadFile)      // GET /api/v1/files/:id/download
			files.GET("/:id/permissions", fileHandler.GetFilePermissions) // GET /api/v1/files/:id/permissions
			files.GET("/:id/access-logs", fileHandler.GetFileAccessLogs)  // GET /api/v1/files/:id/access-logs
			files.POST("/:id/regenerate-public-link", fileHandler.RegeneratePublicLink) // POST /api/v1/files/:id/regenerate-public-link
			files.POST("/:id/generate-one-time-link", fileHandler.GenerateOneTimeDownloadLink) // POST /api/v1/files/:id/generate-one-time-link
		}
		// File write operations require admin role (BT-047)
		filesAdmin := v1.Group("/files")
		filesAdmin.Use(middleware.AuthMiddleware(cfg.JWT.Secret), middleware.AdminPanelMiddleware(db.DB))
		{
			filesAdmin.POST("", fileHandler.UploadFile)               // POST /api/v1/files
			filesAdmin.GET("/selection", fileHandler.GetFilesForProductSelection) // GET /api/v1/files/selection
			filesAdmin.PUT("/:id", fileHandler.UpdateFile)            // PUT /api/v1/files/:id
			filesAdmin.DELETE("/:id", fileHandler.DeleteFile)         // DELETE /api/v1/files/:id
			filesAdmin.PUT("/:id/permissions", fileHandler.UpdateFilePermissions) // PUT /api/v1/files/:id/permissions
			filesAdmin.POST("/:id/permissions/roles", fileHandler.AddRolePermission) // POST /api/v1/files/:id/permissions/roles
			filesAdmin.DELETE("/:id/permissions/roles/:role", fileHandler.RemoveRolePermission) // DELETE /api/v1/files/:id/permissions/roles/:role
			filesAdmin.POST("/:id/permissions/products", fileHandler.AddProductPermission) // POST /api/v1/files/:id/permissions/products
			filesAdmin.DELETE("/:id/permissions/products/:product_id", fileHandler.RemoveProductPermission) // DELETE /api/v1/files/:id/permissions/products/:product_id
		}

		// Public file download (no auth required, uses token)
		filesPublic := v1.Group("/files/public")
		{
			filesPublic.GET("/:token/download", fileHandler.DownloadFilePublic) // GET /api/v1/files/public/:token/download
		}

		// One-time download links (no auth required, uses token)
		filesOneTime := v1.Group("/files/one-time")
		{
			filesOneTime.GET("/:token/download", fileHandler.DownloadFileOneTime) // GET /api/v1/files/one-time/:token/download
		}

		// AI routes (protected)
		ai := v1.Group("/ai")
		ai.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			ai.POST("/format", aiHandler.FormatText) // POST /api/v1/ai/format
			ai.POST("/generate-avatar", aiHandler.GenerateAvatar) // POST /api/v1/ai/generate-avatar
		}
	}

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 5 seconds timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exiting")
}
