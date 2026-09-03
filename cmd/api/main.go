package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/portable-chat/internal/middleware"
	"github.com/kouleen/portable-chat/internal/router"
	"github.com/kouleen/portable-chat/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	router.Register(r)

	r.Use(middleware.Trace())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())
	r.Use(middleware.RouteRecover)

	srv := &http.Server{
		Addr:         ":" + "9191",
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()
	logger.Logger.Info("Server started on port 9191")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Logger.Info("Server exiting")
}
