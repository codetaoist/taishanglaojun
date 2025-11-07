package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codetaoist/taishanglaojun/gateway/internal/config"
	"github.com/codetaoist/taishanglaojun/gateway/internal/discovery"
	"github.com/codetaoist/taishanglaojun/gateway/internal/proxy"
	"github.com/codetaoist/taishanglaojun/gateway/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	defer cfg.Close()

	// Initialize service discovery
	discoveryClient, err := discovery.NewClient(cfg.Discovery)
	if err != nil {
		log.Fatalf("Failed to initialize service discovery: %v", err)
	}

	// Initialize proxy
	proxyManager := proxy.NewManager(discoveryClient, cfg.Proxy)

	// Setup router
	r := router.Setup(cfg, proxyManager)

	// Start HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Gateway server started on port %s", cfg.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}