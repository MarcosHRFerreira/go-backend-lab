// Command go-tweets starts the HTTP server for the API.
package main

import (
	"context"
	"errors"
	"go-tweets/internal/config"
	commentHandler "go-tweets/internal/handler/comment"
	postHandler "go-tweets/internal/handler/post"
	userHandler "go-tweets/internal/handler/user"
	"go-tweets/internal/middleware"
	"go-tweets/internal/observability/logctx"
	obslogger "go-tweets/internal/observability/logger"
	obsmetrics "go-tweets/internal/observability/metrics"
	obstracing "go-tweets/internal/observability/tracing"
	commentRepo "go-tweets/internal/repository/comment"
	postRepo "go-tweets/internal/repository/post"
	userRepo "go-tweets/internal/repository/user"
	commentService "go-tweets/internal/service/comment"
	postService "go-tweets/internal/service/post"
	userService "go-tweets/internal/service/user"

	"go-tweets/pkg/internalsql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	healthCheckTimeout = 2 * time.Second
	serverReadTimeout  = 10 * time.Second
	serverWriteTimeout = 15 * time.Second
	serverIdleTimeout  = 60 * time.Second
	shutdownTimeout    = 10 * time.Second
)

func main() {
	appLogger := obslogger.New(obslogger.Config{
		Service: "go-tweets",
		Env:     os.Getenv("APP_ENV"),
		Version: os.Getenv("APP_VERSION"),
		Level:   os.Getenv("LOG_LEVEL"),
	})
	slog.SetDefault(appLogger)

	r := gin.New()
	validate := validator.New()
	metricsRegistry := obsmetrics.NewRegistry()
	traceProvider, err := obstracing.NewProvider(obstracing.Config{
		Service: "go-tweets",
		Env:     os.Getenv("APP_ENV"),
		Version: os.Getenv("APP_VERSION"),
	})
	if err != nil {
		appLogger.Error("failed to initialize tracing", slog.String("component", "bootstrap"), slog.String("error", err.Error()))
		os.Exit(1)
	}

	httpTracer := traceProvider.Tracer("go-tweets/http")
	dbTracer := traceProvider.Tracer("go-tweets/db")

	cfg, err := config.LoadConfig()
	if err != nil {
		appLogger.Error("failed to load configuration", slog.String("component", "bootstrap"), slog.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("configuration loaded", slog.String("component", "bootstrap"), slog.String("server_address", cfg.ServerAddress()))

	rawDB, err := internalsql.ConnectMySQL(cfg)
	if err != nil {
		appLogger.Error("failed to connect database", slog.String("component", "bootstrap"), slog.String("error", err.Error()))
		os.Exit(1)
	}
	db := internalsql.NewInstrumentedDB(rawDB, metricsRegistry, dbTracer)
	appLogger.Info("database connection ready", slog.String("component", "bootstrap"))

	r.Use(middleware.RequestID(appLogger))
	r.Use(middleware.Trace(httpTracer))
	r.Use(metricsRegistry.HTTPMiddleware())
	r.Use(middleware.AccessLog())
	r.Use(middleware.Recovery())

	r.GET("/metrics", gin.WrapH(metricsRegistry.Handler()))

	// Keep the health check lightweight and bounded so infrastructure probes do not hang.
	// Mantem o health check leve e com tempo limitado para que as sondas de infraestrutura nao fiquem presas.
	r.GET("/check-health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			logctx.FromContext(c.Request.Context()).Warn("health check failed", slog.String("error", err.Error()))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "service is healthy",
		})
	})

	// Wire repositories first so services can compose persistence dependencies explicitly.
	// Conecta os repositories primeiro para que os services componham as dependencias de persistencia de forma explicita.
	userRepo := userRepo.NewRepository(db)
	postRepo := postRepo.NewPostRepository(db)
	commentRepo := commentRepo.NewCommentRepository(db)

	// Services own business rules while handlers stay focused on HTTP concerns.
	// Os services concentram as regras de negocio enquanto os handlers permanecem focados nas preocupacoes HTTP.
	userService := userService.NewUserService(cfg, userRepo)
	postService := postService.NewPostService(cfg, postRepo, commentRepo)
	commentService := commentService.NewCommentService(cfg, commentRepo, postRepo)

	// Handlers register routes after all dependencies are ready.
	// Os handlers registram as rotas somente depois que todas as dependencias estao prontas.
	userHandler := userHandler.NewHandler(r, validate, userService)
	postHandler := postHandler.NewHandler(r, validate, postService)
	commentHandler := commentHandler.NewHandler(r, validate, commentService)

	userHandler.RouteList(cfg.SecretJwt)
	postHandler.RouteList(cfg.SecretJwt)
	commentHandler.RouteList(cfg.SecretJwt)

	server := &http.Server{
		Addr:              cfg.ServerAddress(),
		Handler:           r,
		ReadHeaderTimeout: healthCheckTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		// Run the HTTP server in a separate goroutine so startup and shutdown can be coordinated.
		// Executa o servidor HTTP em uma goroutine separada para coordenar inicializacao e desligamento.
		appLogger.Info("starting http server", slog.String("component", "bootstrap"), slog.String("address", cfg.ServerAddress()))
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}

		close(serverErrors)
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil {
			_ = db.Close()
			appLogger.Error("http server stopped unexpectedly", slog.String("component", "bootstrap"), slog.String("error", err.Error()))
			os.Exit(1)
		}
		return
	case <-shutdownSignal.Done():
		appLogger.Info("shutdown signal received", slog.String("component", "bootstrap"))
	}

	// Give in-flight requests a bounded window to finish before forcing shutdown.
	// Da um tempo limitado para as requisicoes em andamento terminarem antes de forcar o desligamento.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = db.Close()
		_ = obstracing.Shutdown(shutdownCtx, traceProvider)
		appLogger.Error("failed to shutdown http server", slog.String("component", "bootstrap"), slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := db.Close(); err != nil {
		appLogger.Error("failed to close database connection", slog.String("component", "bootstrap"), slog.String("error", err.Error()))
	}

	if err := obstracing.Shutdown(shutdownCtx, traceProvider); err != nil {
		appLogger.Error("failed to shutdown tracing", slog.String("component", "bootstrap"), slog.String("error", err.Error()))
	}

	appLogger.Info("server stopped gracefully", slog.String("component", "bootstrap"))
}
