package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/taishanglaojun/core-services/intelligent-learning/internal/application/services/infrastructure"
)

func main() {
	// Create service manager config
	config := &infrastructure.ServiceManagerConfig{
		ConfigPath:          "./config/infrastructure.json",
		LogLevel:            infrastructure.LogLevelInfo,
		ShutdownTimeout:     30 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableHealthCheck:   true,
		EnableMetrics:       true,
		MetricsPort:         9090,
		EnableProfiling:     false,
		ProfilingPort:       6060,
	}

	// Create service manager
	serviceManager := infrastructure.NewServiceManager(config)

	// Start HTTP server for health checks and metrics
	go startHTTPServer(serviceManager)

	// Run service manager with graceful shutdown
	if err := serviceManager.RunWithGracefulShutdown(); err != nil {
		log.Fatalf("Service manager failed: %v", err)
	}
}

// startHTTPServer starts HTTP server
func startHTTPServer(sm *infrastructure.ServiceManager) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := sm.GetHealthStatus()
		w.Header().Set("Content-Type", "application/json")

		status := health["status"].(string)
		if status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(health)
	})

	// Status endpoint
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := sm.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
	})

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := sm.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(metrics)
	})

	// Learning request processing endpoint
	mux.HandleFunc("/api/v1/learning/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		result, err := sm.ProcessRequest(ctx, request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  result,
		})
	})

	// 获取错误历史端点
	mux.HandleFunc("/api/v1/errors", func(w http.ResponseWriter, r *http.Request) {
		errorHandler := sm.GetErrorHandler()
		errors := errorHandler.GetErrorHistory(50) // 获取最近50个错误
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
			"count":  len(errors),
		})
	})

	// 根路径重定向到健康检查
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/health", http.StatusFound)
	})

	// 根路径HTML文档
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Intelligent Learning Services</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 10px 0; padding: 10px; background: #f5f5f5; border-radius: 5px; }
        .method { font-weight: bold; color: #007acc; }
    </style>
</head>
<body>
    <h1>Intelligent Learning Services</h1>
    <p>Welcome to the Intelligent Learning Services API</p>
    
    <h2>Available Endpoints:</h2>
    
    <div class="endpoint">
        <span class="method">GET</span> <code>/health</code> - Health check
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <code>/status</code> - Service status
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <code>/metrics</code> - Service metrics
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <code>/api/v1/learning/process</code> - Process learning request
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <code>/api/v1/errors</code> - Get error history
    </div>
    
    <h2>Service Information:</h2>
    <p>This service provides intelligent learning capabilities including:</p>
    <ul>
        <li>Cross-modal content processing</li>
        <li>Intelligent relation inference</li>
        <li>Adaptive learning algorithms</li>
        <li>Real-time learning analytics</li>
        <li>Automated knowledge graph construction</li>
        <li>Learning analytics reporting</li>
        <li>Intelligent content recommendation</li>
    </ul>
</body>
</html>
		`)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("HTTP server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server error: %v", err)
	}
}
