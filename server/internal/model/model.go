package model

import "time"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
	RoleAgent  Role = "agent"
)

type Display struct {
	ID     int    `json:"id"`
	Label  string `json:"label"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type DeviceInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	OS       string    `json:"os"`
	Displays []Display `json:"displays"`
}

type DeviceStatus struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	OS         string    `json:"os"`
	Displays   []Display `json:"displays"`
	Online     bool      `json:"online"`
	OnlineAt   time.Time `json:"online_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type SessionStatus struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Viewer    string    `json:"viewer"`
	StartedAt time.Time `json:"started_at"`
	DisplayID int       `json:"display_id"`
	FPS       int       `json:"fps"`
	Quality   string    `json:"quality"`
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type WebRTCConfig struct {
	Mode       string      `json:"mode"`
	ForceRelay bool        `json:"force_relay"`
	IceServers []ICEServer `json:"ice_servers"`
}
