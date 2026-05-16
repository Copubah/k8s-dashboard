package main

import (
	"log"

	"k8s-dashboard/backend/internal/api"
	"k8s-dashboard/backend/internal/config"
	"k8s-dashboard/backend/internal/kube"
)

func main() {
	cfg := config.Load()

	client, err := kube.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to create Kubernetes client: %v", err)
	}

	router := api.NewRouter(cfg, client)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
