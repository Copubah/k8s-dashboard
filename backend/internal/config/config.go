package config

import (
	"os"
	"path/filepath"
)

// Config keeps every runtime setting in one place so deployments can be changed
// with environment variables instead of code edits.
type Config struct {
	Port              string
	Namespace         string
	Kubeconfig        string
	KubeContext       string
	InCluster         bool
	DashboardUsername string
	DashboardPassword string
	AllowedOrigin     string
}

func Load() Config {
	return Config{
		Port:              env("PORT", "8080"),
		Namespace:         env("K8S_NAMESPACE", ""),
		Kubeconfig:        env("KUBECONFIG", defaultKubeconfig()),
		KubeContext:       os.Getenv("KUBE_CONTEXT"),
		InCluster:         env("IN_CLUSTER", "false") == "true",
		DashboardUsername: env("DASHBOARD_USERNAME", "admin"),
		DashboardPassword: env("DASHBOARD_PASSWORD", "change-me"),
		AllowedOrigin:     env("ALLOWED_ORIGIN", "http://localhost:5173"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func defaultKubeconfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}
