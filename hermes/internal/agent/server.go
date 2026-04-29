package agent

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (a *Agent) runLocalServer(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"agent":  a.Snapshot(),
		})
	})
	router.GET("/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, a.Snapshot())
	})
	router.GET("/skills", func(c *gin.Context) {
		a.mu.Lock()
		skills := append([]SkillInfo(nil), a.lastInventory...)
		a.mu.Unlock()
		c.JSON(http.StatusOK, gin.H{"skills": skills})
	})
	router.POST("/commands/poll", func(c *gin.Context) {
		go a.processNextCommand(context.WithoutCancel(ctx))
		c.JSON(http.StatusAccepted, gin.H{"accepted": true})
	})

	server := &http.Server{
		Addr:              a.cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	a.logger.Info("local Gin health API listening", "addr", a.cfg.HTTPAddr)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return context.Canceled
	}
	return err
}
