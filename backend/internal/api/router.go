package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"k8s-dashboard/backend/internal/config"
	"k8s-dashboard/backend/internal/kube"
)

func NewRouter(cfg config.Config, client *kube.Client) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigin},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	handlers := NewHandlers(client)
	protected := router.Group("/api", gin.BasicAuth(gin.Accounts{
		cfg.DashboardUsername: cfg.DashboardPassword,
	}))
	{
		protected.GET("/pods", handlers.ListPods)
		protected.GET("/deployments", handlers.ListDeployments)
		protected.GET("/pods/:namespace/:name/logs", handlers.PodLogs)
		protected.POST("/deployments/:namespace/:name/scale", handlers.ScaleDeployment)
		protected.POST("/deployments/:namespace/:name/restart", handlers.RestartDeployment)
	}

	return router
}
