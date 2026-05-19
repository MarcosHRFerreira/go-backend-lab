// Command go-tweets starts the HTTP server for the API.
package main

import (
	"context"
	"errors"
	"go-tweets/internal/config"
	commentHandler "go-tweets/internal/handler/comment"
	postHandler "go-tweets/internal/handler/post"
	userHandler "go-tweets/internal/handler/user"
	commentRepo "go-tweets/internal/repository/comment"
	postRepo "go-tweets/internal/repository/post"
	userRepo "go-tweets/internal/repository/user"
	commentService "go-tweets/internal/service/comment"
	postService "go-tweets/internal/service/post"
	userService "go-tweets/internal/service/user"

	"go-tweets/pkg/internalsql"
	"log"
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
	r := gin.New()
	validate := validator.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := internalsql.ConnectMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/check-health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "service is healthy",
		})
	})

	userRepo := userRepo.NewRepository(db)
	postRepo := postRepo.NewPostRepository(db)
	commentRepo := commentRepo.NewCommentRepository(db)

	userService := userService.NewUserService(cfg, userRepo)
	postService := postService.NewPostService(cfg, postRepo, commentRepo)
	commentService := commentService.NewCommentService(cfg, commentRepo, postRepo)

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
		log.Printf("starting HTTP server on %s", cfg.ServerAddress())
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
			log.Fatal(err)
		}
		return
	case <-shutdownSignal.Done():
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = db.Close()
		log.Fatal(err)
	}

	if err := db.Close(); err != nil {
		log.Printf("failed to close database connection: %v", err)
	}

	log.Println("server stopped gracefully")
}
