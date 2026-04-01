package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"remotecontrol/server/internal/audit"
	"remotecontrol/server/internal/auth"
	"remotecontrol/server/internal/config"
	"remotecontrol/server/internal/httpapi"
	"remotecontrol/server/internal/hub"
)

func main() {
	log.Printf("Welcome to Aemeath Remote Control Server!")
	log.Printf("正在检查环境变量......")
	envFile := envOrDefault("ENV_FILE", ".env")
	if err := loadEnvFile(envFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("[warning] failed to load env file %q: %v", envFile, err)
		} else {
			log.Printf("[warning] env file not found: %s (set ENV_FILE if needed)", envFile)
		}
	}

	addr := envOrDefault("HTTP_ADDR", ":8080")
	staticDir := envOrDefault("STATIC_DIR", "./web")
	auditPath := envOrDefault("AUDIT_LOG_PATH", "./audit.log")
	appConfig, cfgWarnings := config.Load()
	for _, w := range cfgWarnings {
		log.Printf("[warning] %s", w)
	}
	logStartupWarnings()

	auditLogger, err := audit.NewLogger(auditPath)
	if err != nil {
		log.Fatalf("init audit logger failed: %v", err)
	}
	defer func() {
		_ = auditLogger.Close()
	}()

	authSvc := auth.NewServiceFromEnv()
	messageHub := hub.New(auditLogger, appConfig.WebRTC)
	server := httpapi.New(authSvc, messageHub, auditLogger, staticDir, appConfig.Security, appConfig.WebRTC)

	log.Printf("WebRTC transport mode=%s force_relay=%v ice_servers=%d",
		appConfig.WebRTC.Mode, appConfig.WebRTC.ForceRelay, len(appConfig.WebRTC.IceServers))

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("signal server listening on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down signal server")
	_ = httpServer.Close()
}

func envOrDefault(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}

func logStartupWarnings() {
	log.Printf("正在检查配置文件......")
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" || secret == "change-me-please" || len(secret) < 16 {
		log.Printf("[warning] JWT_SECRET is weak or default, please configure a strong random value")
	}

	if strings.TrimSpace(os.Getenv("ADMIN_PASS")) == "" || os.Getenv("ADMIN_PASS") == "admin123" {
		log.Printf("[warning] ADMIN_PASS is default or empty")
	}
	if strings.TrimSpace(os.Getenv("VIEWER_PASS")) == "" || os.Getenv("VIEWER_PASS") == "viewer123" {
		log.Printf("[warning] VIEWER_PASS is default or empty")
	}
	if strings.TrimSpace(os.Getenv("AGENT_PASS")) == "" || os.Getenv("AGENT_PASS") == "agent123" {
		log.Printf("[warning] AGENT_PASS is default or empty")
	}
}

func loadEnvFile(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		path = ".env"
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid env format at line %d", lineNo)
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			return fmt.Errorf("empty env key at line %d", lineNo)
		}
		value := strings.TrimSpace(parts[1])
		value = trimWrappedQuotes(value)

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s failed at line %d: %w", key, lineNo, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func trimWrappedQuotes(v string) string {
	if len(v) < 2 {
		return v
	}
	if (strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) ||
		(strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'")) {
		return v[1 : len(v)-1]
	}
	return v
}
