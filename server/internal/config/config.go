package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"remotecontrol/server/internal/model"
)

const (
	ModeP2P  = "p2p"
	ModeTURN = "turn"
)

type Security struct {
	AllowedOrigins  []string
	allowAnyOrigin  bool
	allowedOriginDB map[string]struct{}
	LoginMaxAttempt int
	LoginWindow     time.Duration
}

type App struct {
	WebRTC   model.WebRTCConfig
	Security Security
}

func Load() (App, []string) {
	var warnings []string

	mode := strings.ToLower(strings.TrimSpace(defaultIfEmpty(os.Getenv("WEBRTC_MODE"), ModeP2P)))
	if mode != ModeP2P && mode != ModeTURN {
		warnings = append(warnings, "WEBRTC_MODE is invalid, fallback to p2p")
		mode = ModeP2P
	}

	stunURLs := csvList(defaultIfEmpty(os.Getenv("STUN_URLS"), "stun:stun.l.google.com:19302"))
	turnURLs := csvList(os.Getenv("TURN_URLS"))
	turnUser := strings.TrimSpace(os.Getenv("TURN_USERNAME"))
	turnCredential := strings.TrimSpace(os.Getenv("TURN_CREDENTIAL"))

	if mode == ModeTURN && len(turnURLs) == 0 {
		warnings = append(warnings, "WEBRTC_MODE=turn but TURN_URLS is empty, fallback to p2p")
		mode = ModeP2P
	}

	webRTC := model.WebRTCConfig{
		Mode:       mode,
		ForceRelay: mode == ModeTURN,
		IceServers: make([]model.ICEServer, 0, 2),
	}

	switch mode {
	case ModeTURN:
		webRTC.IceServers = append(webRTC.IceServers, model.ICEServer{
			URLs:       turnURLs,
			Username:   turnUser,
			Credential: turnCredential,
		})
		if len(stunURLs) > 0 {
			webRTC.IceServers = append(webRTC.IceServers, model.ICEServer{URLs: stunURLs})
		}
		if turnUser == "" || turnCredential == "" {
			warnings = append(warnings, "TURN mode enabled but TURN_USERNAME / TURN_CREDENTIAL is empty")
		}
	default:
		if len(stunURLs) > 0 {
			webRTC.IceServers = append(webRTC.IceServers, model.ICEServer{URLs: stunURLs})
		}
	}

	origins := csvList(os.Getenv("ALLOWED_ORIGINS"))
	if len(origins) == 0 {
		origins = []string{"http://127.0.0.1:8080", "http://localhost:8080"}
		warnings = append(warnings, "ALLOWED_ORIGINS is empty, using localhost defaults")
	}

	security := Security{
		AllowedOrigins:  origins,
		allowAnyOrigin:  len(origins) == 1 && origins[0] == "*",
		allowedOriginDB: make(map[string]struct{}, len(origins)),
		LoginMaxAttempt: clampInt(envInt("LOGIN_MAX_ATTEMPTS", 10), 1, 1000),
		LoginWindow:     time.Duration(clampInt(envInt("LOGIN_WINDOW_SEC", 60), 5, 3600)) * time.Second,
	}
	for _, origin := range origins {
		security.allowedOriginDB[strings.ToLower(strings.TrimSpace(origin))] = struct{}{}
	}

	return App{
		WebRTC:   webRTC,
		Security: security,
	}, warnings
}

func (s Security) AllowAnyOrigin() bool {
	return s.allowAnyOrigin
}

func (s Security) IsOriginAllowed(origin string) bool {
	if s.allowAnyOrigin {
		return true
	}
	_, ok := s.allowedOriginDB[strings.ToLower(strings.TrimSpace(origin))]
	return ok
}

func defaultIfEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func csvList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		item := strings.TrimSpace(p)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
