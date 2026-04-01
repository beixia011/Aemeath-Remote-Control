package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"remotecontrol/server/internal/audit"
	"remotecontrol/server/internal/auth"
	"remotecontrol/server/internal/config"
	"remotecontrol/server/internal/hub"
	"remotecontrol/server/internal/model"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2 << 20
)

type loginLimiter struct {
	mu      sync.Mutex
	attempt map[string][]time.Time
	max     int
	window  time.Duration
}

func newLoginLimiter(max int, window time.Duration) *loginLimiter {
	return &loginLimiter{
		attempt: make(map[string][]time.Time, 256),
		max:     max,
		window:  window,
	}
}

func (l *loginLimiter) Allow(ip string) bool {
	if ip == "" {
		return true
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	rows := l.compact(ip, now)
	return len(rows) < l.max
}

func (l *loginLimiter) RecordFailure(ip string) {
	if ip == "" {
		return
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	rows := l.compact(ip, now)
	rows = append(rows, now)
	l.attempt[ip] = rows
}

func (l *loginLimiter) Reset(ip string) {
	if ip == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempt, ip)
}

func (l *loginLimiter) compact(ip string, now time.Time) []time.Time {
	rows := l.attempt[ip]
	cut := now.Add(-l.window)
	idx := 0
	for _, t := range rows {
		if t.After(cut) {
			rows[idx] = t
			idx += 1
		}
	}
	rows = rows[:idx]
	l.attempt[ip] = rows
	return rows
}

type Server struct {
	authSvc      *auth.Service
	hub          *hub.Hub
	audit        *audit.Logger
	staticDir    string
	security     config.Security
	webRTC       model.WebRTCConfig
	loginLimiter *loginLimiter
	upgrader     websocket.Upgrader
}

func New(
	authSvc *auth.Service,
	h *hub.Hub,
	a *audit.Logger,
	staticDir string,
	security config.Security,
	webRTC model.WebRTCConfig,
) *Server {
	if staticDir == "" {
		staticDir = "./web"
	}

	s := &Server{
		authSvc:      authSvc,
		hub:          h,
		audit:        a,
		staticDir:    staticDir,
		security:     security,
		webRTC:       webRTC,
		loginLimiter: newLoginLimiter(security.LoginMaxAttempt, security.LoginWindow),
	}
	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     s.checkOrigin,
	}
	return s
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/runtime", s.handleRuntime)
	mux.HandleFunc("/api/devices", s.handleDevices)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/audit", s.handleAudit)
	mux.HandleFunc("/ws", s.handleWS)

	fs := http.FileServer(http.Dir(s.staticDir))
	mux.Handle("/", fs)

	return logRequest(s.withSecurityHeaders(mux))
}

func (s *Server) checkOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		// Non-browser websocket clients (agent) usually have no Origin header.
		return true
	}
	if s.security.AllowAnyOrigin() {
		return true
	}
	return s.security.IsOriginAllowed(origin)
}

func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" && (s.security.AllowAnyOrigin() || s.security.IsOriginAllowed(origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; connect-src 'self' ws: wss:; media-src 'self' blob:; "+
				"img-src 'self' data:; script-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "time": time.Now()})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	ip := clientIP(r.RemoteAddr)
	if !s.loginLimiter.Allow(ip) {
		s.audit.Log("anonymous", "login_rate_limited", map[string]any{"ip": ip})
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "too many login attempts, try again later"})
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid request body"})
		return
	}

	token, role, err := s.authSvc.Login(body.Username, body.Password)
	if err != nil {
		s.loginLimiter.RecordFailure(ip)
		s.audit.Log(body.Username, "login_failed", map[string]any{"ip": ip, "reason": err.Error()})
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}

	s.loginLimiter.Reset(ip)
	s.audit.Log(body.Username, "login_success", map[string]any{"role": role, "ip": ip})
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"role":  role,
	})
}

func (s *Server) handleRuntime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	if _, err := s.requireAuth(r, model.RoleViewer, model.RoleAdmin, model.RoleAgent); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"webrtc": s.webRTC,
	})
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if _, err := s.requireAuth(r, model.RoleViewer, model.RoleAdmin); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"devices": s.hub.ListDevices()})
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if _, err := s.requireAuth(r, model.RoleViewer, model.RoleAdmin); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": s.hub.ListSessions()})
}

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if _, err := s.requireAuth(r, model.RoleAdmin); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": s.audit.List(200)})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(maxMessageSize)

	conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	_, first, err := conn.ReadMessage()
	if err != nil {
		_ = conn.Close()
		return
	}
	var authMsg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(first, &authMsg); err != nil || authMsg.Type != "auth" || authMsg.Token == "" {
		_ = conn.WriteJSON(map[string]any{"type": "error", "message": "first websocket message must be auth token"})
		_ = conn.Close()
		return
	}

	authed, err := s.authSvc.Verify(authMsg.Token)
	if err != nil {
		_ = conn.WriteJSON(map[string]any{"type": "error", "message": "invalid token"})
		_ = conn.Close()
		return
	}

	client := hub.NewClient(conn, authed.Username, authed.Role)
	s.hub.RegisterClient(client)
	defer func() {
		s.hub.UnregisterClient(client)
		_ = conn.Close()
	}()

	go s.writePump(client)
	if b, err := json.Marshal(map[string]any{"type": "ws_ready", "role": authed.Role, "webrtc": s.webRTC}); err == nil {
		select {
		case client.Send <- b:
		default:
		}
	}
	s.readPump(client)
}

func (s *Server) readPump(c *hub.Client) {
	conn := c.Conn
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(_ string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := s.hub.HandleMessage(c, message); err != nil {
			log.Printf("[warning] ws_message_failed user=%s role=%s err=%v", c.Username, c.Role, err)
			b, _ := json.Marshal(map[string]any{
				"type":    "error",
				"message": err.Error(),
			})
			select {
			case c.Send <- b:
			default:
			}
		}
	}
}

func (s *Server) writePump(c *hub.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Server) requireAuth(r *http.Request, allowed ...model.Role) (*auth.AuthenticatedUser, error) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		return nil, errors.New("missing bearer token")
	}
	user, err := s.authSvc.Verify(token)
	if err != nil {
		return nil, errors.New("invalid bearer token")
	}
	for _, role := range allowed {
		if user.Role == role {
			return user, nil
		}
	}
	return nil, errors.New("forbidden")
}

func bearerToken(v string) string {
	parts := strings.SplitN(v, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func clientIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func logRequest(next http.Handler) http.Handler {
	logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
