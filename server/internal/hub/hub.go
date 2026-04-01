package hub

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"remotecontrol/server/internal/audit"
	"remotecontrol/server/internal/model"
)

type Client struct {
	ID       string
	Username string
	Role     model.Role
	Conn     *websocket.Conn
	Send     chan []byte
	DeviceID string

	closeOnce sync.Once
}

func NewClient(conn *websocket.Conn, username string, role model.Role) *Client {
	return &Client{
		ID:       uuid.NewString(),
		Username: username,
		Role:     role,
		Conn:     conn,
		Send:     make(chan []byte, 64),
	}
}

func (c *Client) CloseSend() {
	c.closeOnce.Do(func() {
		close(c.Send)
	})
}

type deviceEntry struct {
	info       model.DeviceInfo
	clientID   string
	onlineAt   time.Time
	lastSeenAt time.Time
}

type sessionEntry struct {
	id        string
	deviceID  string
	viewerID  string
	agentID   string
	startedAt time.Time
	displayID int
	fps       int
	quality   string
}

type Hub struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	devices  map[string]deviceEntry
	sessions map[string]sessionEntry
	audit    *audit.Logger
	webRTC   model.WebRTCConfig
}

func New(a *audit.Logger, webRTC model.WebRTCConfig) *Hub {
	return &Hub{
		clients:  make(map[string]*Client),
		devices:  make(map[string]deviceEntry),
		sessions: make(map[string]sessionEntry),
		audit:    a,
		webRTC:   webRTC,
	}
}

func (h *Hub) RegisterClient(c *Client) {
	h.mu.Lock()
	h.clients[c.ID] = c
	h.mu.Unlock()

	h.audit.Log(c.Username, "ws_connected", map[string]any{
		"client_id": c.ID,
		"role":      c.Role,
	})
	log.Printf("[state] websocket_connected user=%s role=%s client_id=%s", c.Username, c.Role, c.ID)
	h.logStateSnapshot("after_websocket_connected")
}

func (h *Hub) UnregisterClient(c *Client) {
	var ended []sessionEntry
	deviceChanged := false
	removedDeviceID := ""

	h.mu.Lock()
	if _, ok := h.clients[c.ID]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.clients, c.ID)

	if c.DeviceID != "" {
		if _, exists := h.devices[c.DeviceID]; exists {
			delete(h.devices, c.DeviceID)
			deviceChanged = true
			removedDeviceID = c.DeviceID
		}
	}

	for sid, session := range h.sessions {
		if session.viewerID == c.ID || session.agentID == c.ID {
			ended = append(ended, session)
			delete(h.sessions, sid)
		}
	}
	h.mu.Unlock()

	c.CloseSend()
	for _, session := range ended {
		h.notifySessionEnded(session, "peer_disconnected")
		h.audit.Log(c.Username, "session_ended", map[string]any{
			"session_id": session.id,
			"device_id":  session.deviceID,
			"reason":     "peer_disconnected",
		})
		log.Printf("[state] session_ended reason=peer_disconnected session_id=%s device_id=%s actor=%s", session.id, session.deviceID, c.Username)
	}
	if deviceChanged {
		h.notifyViewersDevicesUpdated()
		log.Printf("[state] device_offline device_id=%s actor=%s", removedDeviceID, c.Username)
	}
	h.audit.Log(c.Username, "ws_disconnected", map[string]any{
		"client_id": c.ID,
		"role":      c.Role,
	})
	log.Printf("[state] websocket_disconnected user=%s role=%s client_id=%s", c.Username, c.Role, c.ID)
	h.logStateSnapshot("after_websocket_disconnected")
}

func (h *Hub) HandleMessage(c *Client, raw []byte) error {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &base); err != nil {
		return err
	}

	switch base.Type {
	case "register_device":
		return h.handleRegisterDevice(c, raw)
	case "heartbeat":
		return h.handleHeartbeat(c)
	case "start_session":
		return h.handleStartSession(c, raw)
	case "signal":
		return h.handleSignal(c, raw)
	case "control":
		return h.handleControl(c, raw)
	case "update_session":
		return h.handleUpdateSession(c, raw)
	case "session_end":
		return h.handleSessionEnd(c, raw)
	case "client_warning":
		return h.handleClientWarning(c, raw)
	case "list_devices":
		h.enqueue(c, map[string]any{
			"type":    "devices",
			"devices": h.ListDevices(),
		})
		return nil
	default:
		return errors.New("unknown message type")
	}
}

func (h *Hub) handleRegisterDevice(c *Client, raw []byte) error {
	if c.Role != model.RoleAgent {
		return errors.New("only agent can register device")
	}

	var msg struct {
		Type   string           `json:"type"`
		Device model.DeviceInfo `json:"device"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.Device.Name == "" {
		return errors.New("device name is required")
	}
	if msg.Device.ID == "" {
		msg.Device.ID = "dev-" + uuid.NewString()
	}

	now := time.Now()
	h.mu.Lock()
	if c.DeviceID != "" && c.DeviceID != msg.Device.ID {
		delete(h.devices, c.DeviceID)
	}
	c.DeviceID = msg.Device.ID
	h.devices[msg.Device.ID] = deviceEntry{
		info:       msg.Device,
		clientID:   c.ID,
		onlineAt:   now,
		lastSeenAt: now,
	}
	h.mu.Unlock()

	h.enqueue(c, map[string]any{
		"type":      "device_registered",
		"device_id": msg.Device.ID,
	})
	h.notifyViewersDevicesUpdated()
	h.audit.Log(c.Username, "device_registered", map[string]any{
		"device_id": msg.Device.ID,
		"name":      msg.Device.Name,
	})
	log.Printf("[state] device_online device_id=%s name=%s agent=%s", msg.Device.ID, msg.Device.Name, c.Username)
	h.logStateSnapshot("after_device_online")
	return nil
}

func (h *Hub) handleHeartbeat(c *Client) error {
	if c.DeviceID == "" {
		return nil
	}
	h.mu.Lock()
	d, ok := h.devices[c.DeviceID]
	if ok {
		d.lastSeenAt = time.Now()
		h.devices[c.DeviceID] = d
	}
	h.mu.Unlock()
	return nil
}

func (h *Hub) handleStartSession(c *Client, raw []byte) error {
	if c.Role != model.RoleViewer && c.Role != model.RoleAdmin {
		return errors.New("only viewer/admin can start session")
	}

	var msg struct {
		Type     string `json:"type"`
		DeviceID string `json:"device_id"`
		Display  int    `json:"display_id"`
		FPS      int    `json:"fps"`
		Quality  string `json:"quality"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.DeviceID == "" {
		return errors.New("device_id is required")
	}
	if msg.FPS <= 0 {
		msg.FPS = 15
	}
	if msg.FPS > 60 {
		msg.FPS = 60
	}
	msg.Quality = normalizeQuality(msg.Quality)

	h.mu.Lock()
	device, ok := h.devices[msg.DeviceID]
	if !ok {
		h.mu.Unlock()
		return errors.New("device not online")
	}
	agent, ok := h.clients[device.clientID]
	if !ok {
		h.mu.Unlock()
		return errors.New("device agent unavailable")
	}

	session := sessionEntry{
		id:        uuid.NewString(),
		deviceID:  msg.DeviceID,
		viewerID:  c.ID,
		agentID:   agent.ID,
		startedAt: time.Now(),
		displayID: msg.Display,
		fps:       msg.FPS,
		quality:   msg.Quality,
	}
	h.sessions[session.id] = session
	h.mu.Unlock()

	h.enqueue(c, map[string]any{
		"type":       "session_created",
		"session_id": session.id,
		"device_id":  msg.DeviceID,
		"webrtc":     h.webRTC,
	})
	h.enqueue(agent, map[string]any{
		"type":       "session_offer_request",
		"session_id": session.id,
		"viewer":     c.Username,
		"display_id": msg.Display,
		"fps":        msg.FPS,
		"quality":    msg.Quality,
		"webrtc":     h.webRTC,
	})
	h.audit.Log(c.Username, "session_started", map[string]any{
		"session_id": session.id,
		"device_id":  session.deviceID,
		"display_id": session.displayID,
		"fps":        session.fps,
		"quality":    session.quality,
	})
	log.Printf("[state] session_started session_id=%s device_id=%s viewer=%s display=%d fps=%d quality=%s",
		session.id, session.deviceID, c.Username, session.displayID, session.fps, session.quality)
	h.logStateSnapshot("after_session_started")
	return nil
}

func (h *Hub) handleSignal(c *Client, raw []byte) error {
	var msg struct {
		Type      string          `json:"type"`
		SessionID string          `json:"session_id"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.SessionID == "" {
		return errors.New("session_id is required")
	}

	h.mu.RLock()
	session, ok := h.sessions[msg.SessionID]
	if !ok {
		h.mu.RUnlock()
		return errors.New("session not found")
	}
	var target *Client
	if c.ID == session.viewerID {
		target = h.clients[session.agentID]
	} else if c.ID == session.agentID {
		target = h.clients[session.viewerID]
	} else {
		h.mu.RUnlock()
		return errors.New("sender not in session")
	}
	h.mu.RUnlock()

	if target == nil {
		return errors.New("target disconnected")
	}
	h.enqueue(target, map[string]any{
		"type":       "signal",
		"session_id": msg.SessionID,
		"from":       c.Role,
		"payload":    json.RawMessage(msg.Payload),
	})
	return nil
}

func (h *Hub) handleControl(c *Client, raw []byte) error {
	if c.Role != model.RoleViewer && c.Role != model.RoleAdmin {
		return errors.New("only viewer/admin can send control")
	}

	var msg struct {
		Type      string          `json:"type"`
		SessionID string          `json:"session_id"`
		Event     json.RawMessage `json:"event"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.SessionID == "" {
		return errors.New("session_id is required")
	}

	h.mu.RLock()
	session, ok := h.sessions[msg.SessionID]
	if !ok {
		h.mu.RUnlock()
		return errors.New("session not found")
	}
	if session.viewerID != c.ID {
		h.mu.RUnlock()
		return errors.New("only session owner can control")
	}
	target := h.clients[session.agentID]
	h.mu.RUnlock()
	if target == nil {
		return errors.New("agent disconnected")
	}

	var controlMeta struct {
		Kind    string `json:"kind"`
		FPS     int    `json:"fps"`
		Quality string `json:"quality"`
	}
	if err := json.Unmarshal(msg.Event, &controlMeta); err == nil && controlMeta.Kind == "stream_config" {
		nextFPS := controlMeta.FPS
		if nextFPS <= 0 {
			nextFPS = session.fps
		}
		if nextFPS < 5 {
			nextFPS = 5
		}
		if nextFPS > 60 {
			nextFPS = 60
		}

		nextQuality := controlMeta.Quality
		if strings.TrimSpace(nextQuality) == "" {
			nextQuality = session.quality
		}
		nextQuality = normalizeQuality(nextQuality)

		h.mu.Lock()
		s := h.sessions[msg.SessionID]
		s.fps = nextFPS
		s.quality = nextQuality
		h.sessions[msg.SessionID] = s
		h.mu.Unlock()

		h.audit.Log(c.Username, "session_updated", map[string]any{
			"session_id": msg.SessionID,
			"fps":        nextFPS,
			"quality":    nextQuality,
			"source":     "control_stream_config",
		})
		log.Printf("[state] session_updated source=control session_id=%s viewer=%s fps=%d quality=%s",
			msg.SessionID, c.Username, nextFPS, nextQuality)
	}

	h.enqueue(target, map[string]any{
		"type":       "control",
		"session_id": msg.SessionID,
		"event":      json.RawMessage(msg.Event),
	})
	return nil
}

func (h *Hub) handleUpdateSession(c *Client, raw []byte) error {
	if c.Role != model.RoleViewer && c.Role != model.RoleAdmin {
		return errors.New("only viewer/admin can update session")
	}

	var msg struct {
		Type      string `json:"type"`
		SessionID string `json:"session_id"`
		FPS       int    `json:"fps"`
		Quality   string `json:"quality"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.SessionID == "" {
		return errors.New("session_id is required")
	}

	h.mu.Lock()
	session, ok := h.sessions[msg.SessionID]
	if !ok {
		h.mu.Unlock()
		return errors.New("session not found")
	}
	if session.viewerID != c.ID {
		h.mu.Unlock()
		return errors.New("only session owner can update")
	}
	agent := h.clients[session.agentID]
	if agent == nil {
		h.mu.Unlock()
		return errors.New("agent disconnected")
	}

	if msg.FPS <= 0 {
		msg.FPS = session.fps
	}
	if msg.FPS < 5 {
		msg.FPS = 5
	}
	if msg.FPS > 60 {
		msg.FPS = 60
	}

	if strings.TrimSpace(msg.Quality) == "" {
		msg.Quality = session.quality
	}
	msg.Quality = normalizeQuality(msg.Quality)

	session.fps = msg.FPS
	session.quality = msg.Quality
	h.sessions[msg.SessionID] = session
	h.mu.Unlock()

	h.enqueue(agent, map[string]any{
		"type":       "session_update",
		"session_id": msg.SessionID,
		"fps":        msg.FPS,
		"quality":    msg.Quality,
	})
	h.enqueue(c, map[string]any{
		"type":       "session_updated",
		"session_id": msg.SessionID,
		"fps":        msg.FPS,
		"quality":    msg.Quality,
	})

	h.audit.Log(c.Username, "session_updated", map[string]any{
		"session_id": msg.SessionID,
		"fps":        msg.FPS,
		"quality":    msg.Quality,
	})
	log.Printf("[state] session_updated session_id=%s viewer=%s fps=%d quality=%s",
		msg.SessionID, c.Username, msg.FPS, msg.Quality)
	return nil
}

func (h *Hub) handleSessionEnd(c *Client, raw []byte) error {
	var msg struct {
		Type      string `json:"type"`
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.SessionID == "" {
		return errors.New("session_id is required")
	}
	h.endSession(msg.SessionID, "ended_by_"+string(c.Role), c.Username)
	return nil
}

func (h *Hub) handleClientWarning(c *Client, raw []byte) error {
	var msg struct {
		Type      string         `json:"type"`
		SessionID string         `json:"session_id"`
		Code      string         `json:"code"`
		Message   string         `json:"message"`
		Details   map[string]any `json:"details"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	if msg.Code == "" {
		return errors.New("warning code is required")
	}

	payload := map[string]any{
		"client_id":  c.ID,
		"role":       c.Role,
		"session_id": msg.SessionID,
		"code":       msg.Code,
		"message":    msg.Message,
		"details":    msg.Details,
	}
	h.audit.Log(c.Username, "client_warning", payload)
	log.Printf("[warning] client=%s role=%s code=%s session=%s message=%s details=%v",
		c.Username, c.Role, msg.Code, msg.SessionID, msg.Message, msg.Details)
	return nil
}

func (h *Hub) endSession(sessionID, reason, actor string) {
	h.mu.Lock()
	session, ok := h.sessions[sessionID]
	if !ok {
		h.mu.Unlock()
		return
	}
	delete(h.sessions, sessionID)
	h.mu.Unlock()

	h.notifySessionEnded(session, reason)
	h.audit.Log(actor, "session_ended", map[string]any{
		"session_id": sessionID,
		"reason":     reason,
	})
	log.Printf("[state] session_ended reason=%s session_id=%s actor=%s", reason, sessionID, actor)
	h.logStateSnapshot("after_session_ended")
}

func (h *Hub) notifySessionEnded(session sessionEntry, reason string) {
	h.mu.RLock()
	viewer := h.clients[session.viewerID]
	agent := h.clients[session.agentID]
	h.mu.RUnlock()

	msg := map[string]any{
		"type":       "session_ended",
		"session_id": session.id,
		"reason":     reason,
	}
	if viewer != nil {
		h.enqueue(viewer, msg)
	}
	if agent != nil {
		h.enqueue(agent, msg)
	}
}

func (h *Hub) notifyViewersDevicesUpdated() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		if c.Role == model.RoleViewer || c.Role == model.RoleAdmin {
			h.enqueue(c, map[string]any{"type": "devices_updated"})
		}
	}
}

func (h *Hub) enqueue(c *Client, payload any) {
	if c == nil {
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	select {
	case c.Send <- b:
	default:
	}
}

func (h *Hub) ListDevices() []model.DeviceStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]model.DeviceStatus, 0, len(h.devices))
	for _, d := range h.devices {
		out = append(out, model.DeviceStatus{
			ID:         d.info.ID,
			Name:       d.info.Name,
			OS:         d.info.OS,
			Displays:   d.info.Displays,
			Online:     true,
			OnlineAt:   d.onlineAt,
			LastSeenAt: d.lastSeenAt,
		})
	}
	return out
}

func (h *Hub) ListSessions() []model.SessionStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]model.SessionStatus, 0, len(h.sessions))
	for _, s := range h.sessions {
		viewer := "unknown"
		if c, ok := h.clients[s.viewerID]; ok {
			viewer = c.Username
		}
		out = append(out, model.SessionStatus{
			ID:        s.id,
			DeviceID:  s.deviceID,
			Viewer:    viewer,
			StartedAt: s.startedAt,
			DisplayID: s.displayID,
			FPS:       s.fps,
			Quality:   s.quality,
		})
	}
	return out
}

func (h *Hub) WebRTCConfig() model.WebRTCConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.webRTC
}

func (h *Hub) logStateSnapshot(tag string) {
	clients, devices, sessions := h.snapshotForLog()
	log.Printf("[state] %s clients=%d devices=%d sessions=%d device_ids=[%s] session_refs=[%s]",
		tag, clients, len(devices), len(sessions), strings.Join(devices, ", "), strings.Join(sessions, ", "))
}

func (h *Hub) snapshotForLog() (int, []string, []string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	deviceIDs := make([]string, 0, len(h.devices))
	for id := range h.devices {
		deviceIDs = append(deviceIDs, id)
	}
	sort.Strings(deviceIDs)

	sessionRefs := make([]string, 0, len(h.sessions))
	for _, s := range h.sessions {
		viewer := "unknown"
		if c, ok := h.clients[s.viewerID]; ok {
			viewer = c.Username
		}
		ref := s.id + "(viewer=" + viewer + ",device=" + s.deviceID + ")"
		sessionRefs = append(sessionRefs, ref)
	}
	sort.Strings(sessionRefs)

	return len(h.clients), deviceIDs, sessionRefs
}

func normalizeQuality(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "low":
		return "low"
	case "high":
		return "high"
	case "ultra":
		return "ultra"
	default:
		return "medium"
	}
}
